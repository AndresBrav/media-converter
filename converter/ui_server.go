package converter

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

//go:embed ui ui/assets/style.css ui/assets/app.js
var uiFS embed.FS

type StartRequest struct {
	InputDir  string `json:"inputDir"`
	OutputDir string `json:"outputDir"`
	Format    string `json:"format"`
	Workers   int    `json:"workers"`
	Quality   int    `json:"quality"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Watermark string `json:"watermark"`
	Thumbnail bool   `json:"thumbnail"`
	Recursive bool   `json:"recursive"`
}

var (
	activeCtx    context.Context
	activeCancel context.CancelFunc
	activeMu     sync.Mutex
	isRunning    bool

	clientsMu sync.Mutex
	clients   = make(map[chan ProgressEvent]bool)
)

func broadcastEvent(event ProgressEvent) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	for ch := range clients {
		select {
		case ch <- event:
		default:
		}
	}
}

// StartUIServer inicia el servidor HTTP local y abre el navegador predeterminado.
func StartUIServer(port int) error {
	subFS, err := fs.Sub(uiFS, "ui")
	if err != nil {
		return fmt.Errorf("error al acceder a los assets embebidos: %w", err)
	}

	mux := http.NewServeMux()

	// SSE: eventos en tiempo real
	mux.HandleFunc("/api/events", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming no soportado", http.StatusInternalServerError)
			return
		}

		ch := make(chan ProgressEvent, 64)
		clientsMu.Lock()
		clients[ch] = true
		clientsMu.Unlock()
		defer func() {
			clientsMu.Lock()
			delete(clients, ch)
			clientsMu.Unlock()
			close(ch)
		}()

		for {
			select {
			case <-r.Context().Done():
				return
			case ev := <-ch:
				data, err := json.Marshal(ev)
				if err != nil {
					continue
				}
				fmt.Fprintf(w, "data: %s\n\n", data)
				flusher.Flush()
			}
		}
	})

	// Diálogo nativo de selección
	mux.HandleFunc("/api/browse", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		t := r.URL.Query().Get("type")
		if t == "" {
			t = "directory"
		}
		path, err := browseDialog(t)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"path": path})
	})

	// Iniciar conversión
	mux.HandleFunc("/api/start", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var req StartRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "JSON inválido"})
			return
		}

		activeMu.Lock()
		if isRunning {
			activeMu.Unlock()
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Ya hay un proceso activo"})
			return
		}
		isRunning = true
		activeCtx, activeCancel = context.WithCancel(context.Background())
		activeMu.Unlock()

		go func() {
			defer func() {
				activeMu.Lock()
				isRunning = false
				if activeCancel != nil {
					activeCancel()
				}
				activeMu.Unlock()
			}()

			opts := Options{
				Quality:   req.Quality,
				Width:     req.Width,
				Height:    req.Height,
				Watermark: req.Watermark,
				Thumbnail: req.Thumbnail,
			}

			if err := RunConversion(activeCtx, req.InputDir, req.OutputDir, req.Format, opts, req.Workers, req.Recursive, broadcastEvent); err != nil {
				broadcastEvent(ProgressEvent{
					Type:  "complete",
					Error: err.Error(),
				})
			}
		}()

		json.NewEncoder(w).Encode(map[string]string{"status": "started"})
	})

	// Detener conversión
	mux.HandleFunc("/api/stop", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		activeMu.Lock()
		if isRunning && activeCancel != nil {
			activeCancel()
		}
		activeMu.Unlock()
		json.NewEncoder(w).Encode(map[string]string{"status": "stopped"})
	})

	// Servir assets estáticos embebidos
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		data, err := fs.ReadFile(subFS, path)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "No encontrado: %s", path)
			return
		}

		switch {
		case strings.HasSuffix(path, ".css"):
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
		case strings.HasSuffix(path, ".js"):
			w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		default:
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		}
		w.Write(data)
	})

	// Abrir navegador
	go func() {
		time.Sleep(350 * time.Millisecond)
		openBrowser(fmt.Sprintf("http://localhost:%d", port))
	}()

	log.Printf("Media Converter UI → http://localhost:%d\n", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = exec.Command("xdg-open", url).Start()
	}
	if err != nil {
		fmt.Printf("Abre manualmente: %s\n", url)
	}
}

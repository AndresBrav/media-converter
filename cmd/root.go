package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"

	"media-converter/converter"

	"github.com/spf13/cobra"
)

var (
	inputDir      string
	outputDir     string
	format        string
	numWorkers    int
	quality       int
	width         int
	height        int
	watermarkPath string
	thumbnail     bool
	recursive     bool
)

var rootCmd = &cobra.Command{
	Use:   "media-converter",
	Short: "Convert images and videos",
	Long: `Media Converter is a CLI tool built with Go
to process files concurrently using worker pools.`,

	Run: func(cmd *cobra.Command, args []string) {
		// 1. Validar directorio de entrada
		if !converter.ValidateInput(inputDir) {
			return
		}

		// 2. Preparar/Validar directorio de salida
		if !converter.PrepareOutput(outputDir) {
			return
		}

		// 3. Validar formato solicitado
		normalizedFormat, ok := converter.ValidateFormat(format)
		if !ok {
			return
		}
		format = normalizedFormat

		// 5. Validar número de workers
		if numWorkers < 1 {
			fmt.Println("Error: --workers debe ser mayor a 0")
			return
		}

		// 6. Validar calidad
		if quality < 1 || quality > 100 {
			fmt.Println("Error: --quality debe estar entre 1 y 100")
			return
		}

		// 6.1 Validar watermark si se proporciona
		if watermarkPath != "" {
			if _, err := os.Stat(watermarkPath); os.IsNotExist(err) {
				fmt.Printf("Error: archivo de watermark '%s' no existe\n", watermarkPath)
				return
			}
		}

		// 6.2 Validar width/height
		if width < 0 {
			fmt.Println("Error: --width debe ser un número positivo o 0")
			return
		}
		if height < 0 {
			fmt.Println("Error: --height debe ser un número positivo o 0")
			return
		}

		// 7. Mostrar configuración de ejecución
		converter.ShowConfig(inputDir, outputDir, format, numWorkers)

		// 8. Obtener y listar trabajos (jobs) a procesar
		jobsToProcess, err := converter.GetJobs(inputDir, outputDir, format, recursive)
		if err != nil {
			fmt.Println("Failed to list input files")
			return
		}

		// Verificar si se requiere FFmpeg (solo si el formato de salida es .mp4 y hay videos en la lista)
		requiresFFmpeg := false
		if format == ".mp4" {
			for _, job := range jobsToProcess {
				if converter.IsVideoInput(filepath.Ext(job.InputPath)) {
					requiresFFmpeg = true
					break
				}
			}
		}

		if requiresFFmpeg {
			if err := exec.Command("ffmpeg", "-version").Run(); err != nil {
				fmt.Println("ERROR: FFmpeg no está instalado o no está en el PATH.")
				fmt.Println("FFmpeg es necesario para convertir videos.")
				fmt.Println()
				fmt.Println("Puedes instalarlo con el siguiente comando:")
				fmt.Println("  winget install ffmpeg")
				fmt.Println()
				fmt.Println("O descárgalo desde: https://ffmpeg.org/download.html")
				return
			}
		}

		converter.Resume(jobsToProcess)

		// 9. Calcular workers efectivos
		effectiveWorkers := numWorkers
		if effectiveWorkers > len(jobsToProcess) {
			effectiveWorkers = len(jobsToProcess)
		}

		opts := converter.Options{
			Quality:   quality,
			Width:     width,
			Height:    height,
			Watermark: watermarkPath,
			Thumbnail: thumbnail,
		}

		if width > 0 || height > 0 {
			fmt.Printf("Resize : %dx%d\n", width, height)
		}
		if watermarkPath != "" {
			fmt.Println("Watermark :", watermarkPath)
		}
		if thumbnail {
			fmt.Println("Thumbnail : 150x150")
		}

		fmt.Printf("Lanzando %d workers...\n\n", effectiveWorkers)

		var printMu sync.Mutex

		convErr := converter.RunConversion(
			context.Background(),
			inputDir, outputDir, format,
			opts, numWorkers, recursive,
			func(ev converter.ProgressEvent) {
				printMu.Lock()
				defer printMu.Unlock()
				switch ev.Type {
				case "worker_start":
					fmt.Printf("Worker %d ▶ %s\n", ev.WorkerID, filepath.Base(ev.InputPath))
				case "worker_end":
					if ev.Status == "success" {
						fmt.Printf("[%d/%d] ✓ %s → %s\n", ev.Current, ev.Total, filepath.Base(ev.InputPath), filepath.Base(ev.OutputPath))
					} else {
						fmt.Printf("[%d/%d] ✗ %s: %s\n", ev.Current, ev.Total, filepath.Base(ev.InputPath), ev.Error)
					}
				case "complete":
					fmt.Printf("\nCompletado — %d exitosos, %d errores, %.1fs\n",
						ev.Current-ev.Failed, ev.Failed, ev.Elapsed)
				}
			},
		)
		if convErr != nil {
			fmt.Printf("Error: %v\n", convErr)
		}

	},
}

// Execute añade todos los comandos hijos al comando raíz y configura las banderas adecuadamente.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&inputDir, "input", "i", "", "Input directory")
	rootCmd.Flags().StringVarP(&outputDir, "output", "o", "", "Output directory")
	rootCmd.Flags().StringVarP(&format, "format", "f", "", "Output format")
	rootCmd.Flags().IntVarP(&numWorkers, "workers", "w", runtime.NumCPU(), "Number of parallel workers (default: number of CPU cores)")
	rootCmd.Flags().IntVarP(&quality, "quality", "q", 80, "Quality of the output image (1-100)")
	rootCmd.Flags().IntVarP(&width, "width", "", 0, "Max width in pixels (0 = keep original)")
	rootCmd.Flags().IntVarP(&height, "height", "", 0, "Max height in pixels (0 = keep original)")
	rootCmd.Flags().StringVarP(&watermarkPath, "watermark", "", "", "Path to watermark image")
	rootCmd.Flags().BoolVarP(&thumbnail, "thumbnail", "", false, "Generate 150x150 thumbnail")
	rootCmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Process directories recursively")

	if err := rootCmd.MarkFlagRequired("input"); err != nil {
		panic(err)
	}
	if err := rootCmd.MarkFlagRequired("output"); err != nil {
		panic(err)
	}
	if err := rootCmd.MarkFlagRequired("format"); err != nil {
		panic(err)
	}
}

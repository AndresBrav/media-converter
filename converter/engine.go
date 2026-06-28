package converter

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ProgressEvent define la estructura de los eventos de progreso para la UI y CLI.
type ProgressEvent struct {
	Type       string  `json:"type"`                  // "start", "worker_start", "worker_end", "complete", "stopped", "log"
	WorkerID   int     `json:"worker_id,omitempty"`
	InputPath  string  `json:"input_path,omitempty"`
	OutputPath string  `json:"output_path,omitempty"`
	Status     string  `json:"status,omitempty"`       // "success", "failed"
	Error      string  `json:"error,omitempty"`
	Current    int     `json:"current,omitempty"`
	Failed     int     `json:"failed,omitempty"`
	Total      int     `json:"total,omitempty"`
	Elapsed    float64 `json:"elapsed,omitempty"`
}

// RunConversion ejecuta el procesamiento por lotes concurrente con soporte para cancelación y eventos en tiempo real.
func RunConversion(ctx context.Context, inputDir, outputDir, format string, opts Options, numWorkers int, recursive bool, onEvent func(ProgressEvent)) error {
	// 1. Validar directorio de entrada
	if !ValidateInput(inputDir) {
		return fmt.Errorf("directorio de entrada no válido: %s", inputDir)
	}

	// 2. Preparar/Validar directorio de salida
	if !PrepareOutput(outputDir) {
		return fmt.Errorf("no se pudo preparar el directorio de salida: %s", outputDir)
	}

	// 3. Validar formato solicitado
	normFormat, ok := ValidateFormat(format)
	if !ok {
		return fmt.Errorf("formato no soportado: %s", format)
	}

	// 4. Obtener trabajos
	jobsToProcess, err := GetJobs(inputDir, outputDir, normFormat, recursive)
	if err != nil {
		return fmt.Errorf("error al listar archivos: %w", err)
	}

	total := len(jobsToProcess)
	if total == 0 {
		onEvent(ProgressEvent{
			Type:    "complete",
			Current: 0,
			Total:   0,
			Elapsed: 0,
		})
		return nil
	}

	// Notificar inicio
	onEvent(ProgressEvent{
		Type:  "start",
		Total: total,
	})

	// Calcular workers efectivos
	effectiveWorkers := numWorkers
	if effectiveWorkers > total {
		effectiveWorkers = total
	}

	jobsChan := make(chan Job, total)
	var wg sync.WaitGroup
	var completed int32
	var failed int32

	startTime := time.Now()

	// Lanzar workers
	for i := 1; i <= effectiveWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for job := range jobsChan {
				// Comprobar cancelación antes de procesar
				if ctx.Err() != nil {
					return
				}

				onEvent(ProgressEvent{
					Type:      "worker_start",
					WorkerID:  workerID,
					InputPath: job.InputPath,
				})

				convertErr := convert(job, opts)

				var status string
				var errStr string
				if convertErr != nil {
					status = "failed"
					errStr = convertErr.Error()
					atomic.AddInt32(&failed, 1)
				} else {
					status = "success"
					atomic.AddInt32(&completed, 1)
				}

				curr := atomic.LoadInt32(&completed) + atomic.LoadInt32(&failed)
				elapsedSecs := time.Since(startTime).Seconds()

				onEvent(ProgressEvent{
					Type:       "worker_end",
					WorkerID:   workerID,
					InputPath:  job.InputPath,
					OutputPath: job.OutputPath,
					Status:     status,
					Error:      errStr,
					Current:    int(curr),
					Failed:     int(atomic.LoadInt32(&failed)),
					Total:      total,
					Elapsed:    elapsedSecs,
				})
			}
		}(i)
	}

	// Llenar el canal con soporte de cancelación
	go func() {
		defer close(jobsChan)
		for _, job := range jobsToProcess {
			select {
			case <-ctx.Done():
				return
			case jobsChan <- job:
			}
		}
	}()

	// Esperar a que terminen todos los workers
	wg.Wait()

	elapsed := time.Since(startTime).Seconds()

	if ctx.Err() != nil {
		onEvent(ProgressEvent{
			Type:    "stopped",
			Current: int(atomic.LoadInt32(&completed) + atomic.LoadInt32(&failed)),
			Failed:  int(atomic.LoadInt32(&failed)),
			Total:   total,
			Elapsed: elapsed,
		})
	} else {
		onEvent(ProgressEvent{
			Type:    "complete",
			Current: int(atomic.LoadInt32(&completed) + atomic.LoadInt32(&failed)),
			Failed:  int(atomic.LoadInt32(&failed)),
			Total:   total,
			Elapsed: elapsed,
		})
	}

	return nil
}

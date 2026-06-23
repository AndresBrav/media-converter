package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"

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
		jobsToProcess, err := converter.GetJobs(inputDir, outputDir, format)
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

		// 9. Calcular workers efectivos (no más workers que jobs disponibles)
		effectiveWorkers := numWorkers
		if effectiveWorkers > len(jobsToProcess) {
			effectiveWorkers = len(jobsToProcess)
		}

		jobs := make(chan converter.Job, len(jobsToProcess))
		var waitGroup sync.WaitGroup
		var completed int32
		var failed int32
		totalJobs := len(jobsToProcess)

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

		startTime := time.Now()

		// Lanzar workers
		fmt.Printf("Lanzando %d workers...\n\n", effectiveWorkers)
		for i := 0; i < effectiveWorkers; i++ {
			waitGroup.Add(1)
			go converter.Worker(i+1, jobs, &waitGroup, &completed, &failed, totalJobs, opts)
		}

		//Llenar el canal con los jobs
		for _, job := range jobsToProcess {
			jobs <- job
		}
		close(jobs)

		//Esperar a que terminen los workers
		waitGroup.Wait()

		elapsed := time.Since(startTime)

		fmt.Println("\nTodos los trabajos completados")
		fmt.Printf("Errores: %d\n", failed)
		fmt.Printf("Total: %.1fs\n", elapsed.Seconds())

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

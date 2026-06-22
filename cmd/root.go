package cmd

import (
	"fmt"
	"os"
	"sync"
	"runtime"
	"time"

	"github.com/spf13/cobra"
	"media-converter/converter"
)

var (
	inputDir   string
	outputDir  string
	format     string
	numWorkers int
	quality    int
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

		// 4. Validar número de workers
		if numWorkers < 1 {
			fmt.Println("Error: --workers debe ser mayor a 0")
			return
		}

		// Validar calidad
		if quality < 1 || quality > 100 {
			fmt.Println("Error: --quality debe estar entre 1 y 100")
			return
		}

		// 5. Mostrar configuración de ejecución
		converter.ShowConfig(inputDir, outputDir, format, numWorkers)

		// 6. Obtener y listar trabajos (jobs) a procesar
		jobsToProcess, err := converter.GetJobs(inputDir, outputDir, format)
		if err != nil {
			fmt.Println("Failed to list input files")
			return
		}

		converter.Resume(jobsToProcess)

		// 7. Calcular workers efectivos (no más workers que jobs disponibles)
		effectiveWorkers := numWorkers
		if effectiveWorkers > len(jobsToProcess) {
			effectiveWorkers = len(jobsToProcess)
		}

		jobs:= make(chan converter.Job, len(jobsToProcess))
		var waitGroup sync.WaitGroup
		var completed int32
		var failed int32
		totalJobs := len(jobsToProcess)

		startTime := time.Now()

		// Lanzar workers
		fmt.Printf("Lanzando %d workers...\n\n", effectiveWorkers)
		for i := 0; i < effectiveWorkers; i++ {
			waitGroup.Add(1)
			go converter.Worker(i+1, jobs, &waitGroup, &completed, &failed, totalJobs, quality)
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

package cmd

import (
	"fmt"
	"os"
	"sync"
	"runtime"

	"github.com/spf13/cobra"
	"media-converter/converter"
)

var (
	inputDir  string
	outputDir string
	format    string
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

		// 4. Mostrar configuración de ejecución
		converter.ShowConfig(inputDir, outputDir, format)

		// 5. Obtener y listar trabajos (jobs) a procesar
		jobsToProcess, err := converter.GetJobs(inputDir, outputDir, format)
		if err != nil {
			fmt.Println("Failed to list input files")
			return
		}

		//6. Crear el pool de workers, segun la cantidad de procesadores
		numProcessors := runtime.NumCPU()

		jobs:= make(chan converter.Job, len(jobsToProcess))
		var waitGroup sync.WaitGroup

		//Lanzar workers
		for i:= 0; i < numProcessors; i++ {
			waitGroup.Add(1)
			go converter.Worker(jobs, &waitGroup)
		}

		//Llenar el canal con los jobs
		for _, job := range jobsToProcess {
			jobs <- job  
		}
		close(jobs)

		//Esperar a que terminen los workers
		waitGroup.Wait()

		fmt.Println("\nTodos los trabajos completados")


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

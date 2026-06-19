package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	inputDir  string
	outputDir string
	format    string
)

var supportedFormats = map[string]bool{
	"webp": true,
}

var rootCmd = &cobra.Command{
	Use:   "media-converter",
	Short: "Convert images and videos",
	Long: `Media Converter is a CLI tool built with Go
to process files concurrently using worker pools.`,

	Run: func(cmd *cobra.Command, args []string) {

		// -------------------------
		// Validar input
		// -------------------------

		inputInfo, err := os.Stat(inputDir)

		if err != nil {
			fmt.Printf(
				"Input directory '%s' does not exist\n",
				inputDir,
			)
			return
		}

		if !inputInfo.IsDir() {
			fmt.Printf(
				"'%s' is not a directory\n",
				inputDir,
			)
			return
		}

		// -------------------------
		// Crear output si no existe
		// -------------------------

		if _, err := os.Stat(outputDir); os.IsNotExist(err) {

			err := os.MkdirAll(outputDir, 0755)

			if err != nil {
				fmt.Printf(
					"Failed to create output directory '%s'\n",
					outputDir,
				)
				return
			}

			fmt.Printf(
				"Output directory created: %s\n",
				outputDir,
			)
		}

		// -------------------------
		// Validar output
		// -------------------------

		outputInfo, err := os.Stat(outputDir)

		if err != nil {
			fmt.Printf(
				"Unable to access output directory '%s'\n",
				outputDir,
			)
			return
		}

		if !outputInfo.IsDir() {
			fmt.Printf(
				"'%s' is not a directory\n",
				outputDir,
			)
			return
		}

		// -------------------------
		// Validar permisos escritura
		// -------------------------

		testFile := filepath.Join(
			outputDir,
			".write_test",
		)

		file, err := os.Create(testFile)

		if err != nil {
			fmt.Printf(
				"No write permission in '%s'\n",
				outputDir,
			)
			return
		}

		file.Close()

		_ = os.Remove(testFile)

		// -------------------------
		// Validar formato
		// -------------------------

		if !supportedFormats[format] {

			fmt.Printf(
				"Unsupported format '%s'\n",
				format,
			)

			fmt.Println("Supported formats:")

			for f := range supportedFormats {
				fmt.Println("-", f)
			}

			return
		}

		// -------------------------
		// Mostrar configuración
		// -------------------------

		fmt.Println()
		fmt.Println("Configuration")
		fmt.Println("-------------")
		fmt.Println("Input :", inputDir)
		fmt.Println("Output:", outputDir)
		fmt.Println("Format:", format)
	},
}

func Execute() {
	err := rootCmd.Execute()

	if err != nil {
		os.Exit(1)
	}
}

func init() {

	rootCmd.Flags().StringVarP(
		&inputDir,
		"input",
		"i",
		"",
		"Input directory",
	)

	rootCmd.Flags().StringVarP(
		&outputDir,
		"output",
		"o",
		"",
		"Output directory",
	)

	rootCmd.Flags().StringVarP(
		&format,
		"format",
		"f",
		"",
		"Output format",
	)

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
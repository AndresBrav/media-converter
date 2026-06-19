package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var inputDir string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "media-converter",
	Short: "Convert images and videos",
	Long: `Media Converter is a CLI tool built with Go
to process files concurrently using worker pools.`,

	Run: func(cmd *cobra.Command, args []string) {

		info, err := os.Stat(inputDir)

		if err != nil {
			fmt.Printf(
				"Input directory '%s' does not exist\n",
				inputDir,
			)
			return
		}

		if !info.IsDir() {
			fmt.Printf(
				"'%s' is not a directory\n",
				inputDir,
			)
			return
		}

		fmt.Printf(
			"Input directory found: %s\n",
			inputDir,
		)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
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

	err := rootCmd.MarkFlagRequired("input")

	if err != nil {
		panic(err)
	}
}
package cmd

import (
	"fmt"
	"media-converter/converter"

	"github.com/spf13/cobra"
)

var port int

var uiCmd = &cobra.Command{
	Use:   "ui",
	Short: "Abre la interfaz gráfica en el navegador",
	Long:  `Inicia un servidor local y abre automáticamente Media Converter en el navegador web.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Iniciando Media Converter GUI en el puerto %d...\n", port)
		if err := converter.StartUIServer(port); err != nil {
			fmt.Printf("Error al iniciar el servidor: %v\n", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(uiCmd)
	uiCmd.Flags().IntVarP(&port, "port", "p", 8080, "Puerto del servidor local")
}

package converter

import "fmt"

// ShowConfig imprime la configuración inicial en la consola.
func ShowConfig(input, output, format string) {
	fmt.Println()
	fmt.Println("Configuration")
	fmt.Println("-------------")
	fmt.Println("Input :", input)
	fmt.Println("Output:", output)
	fmt.Println("Format:", format)
}

// Resume imprime la lista de archivos que se van a procesar.
func Resume(files []File) {
	fmt.Println("Resume")
	for _, f := range files {
		fmt.Println(f.InputPath, "->", f.OutputPath)
	}
	fmt.Printf("Archivos encontrados: %d\n", len(files))
}

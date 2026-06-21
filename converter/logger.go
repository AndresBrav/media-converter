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

// Resume imprime la lista de trabajos (jobs) que se van a procesar.
func Resume(jobs []Job) {
	fmt.Println("Resume")
	for _, job := range jobs {
		fmt.Println(job.InputPath, "->", job.OutputPath)
	}
	fmt.Printf("Archivos encontrados: %d\n", len(jobs))
}

package converter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Job almacena las rutas del archivo a procesar.
type Job struct {
	InputPath  string
	OutputPath string
}

// GetJobs lee el directorio de entrada, filtra por extensiones soportadas y genera automáticamente la ruta de destino final para cada Job.
func GetJobs(inputDir string, outputDir string, format string) ([]Job, error) {
	var jobs []Job
	inputFile, err := os.ReadDir(inputDir)

	if err != nil {
		fmt.Printf("Error reading directory: '%s'\n", inputDir)
		return nil, err
	}

	for _, f := range inputFile {
		if f.IsDir() {
			continue
		}
		inputPath := filepath.Join(inputDir, f.Name())
		ext := filepath.Ext(f.Name())
		if IsSupportedFormat(ext) {
			// Calcular la ruta de salida específica (ej. resultado/foto1.webp)
			baseName := strings.TrimSuffix(f.Name(), ext)
			newFileName := baseName + format
			outputPath := filepath.Join(outputDir, newFileName)

			jobs = append(jobs,
				Job{
					InputPath:  inputPath,
					OutputPath: outputPath,
				})
		}
	}

	return jobs, nil
}

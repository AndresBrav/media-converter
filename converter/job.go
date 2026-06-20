package converter

import (
	"fmt"
	"os"
	"path/filepath"
)

// Job almacena las rutas del archivo a procesar.
type Job struct {
	InputPath  string
	OutputPath string
}

// GetJobs lee el directorio de entrada y filtra por extensiones soportadas, devolviendo una lista de Jobs.
func GetJobs(inputDir string, outputDir string) ([]Job, error) {
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
		if IsSupportedFormat(filepath.Ext(f.Name())) {
			jobs = append(jobs,
				Job{
					InputPath:  inputPath,
					OutputPath: outputDir,
				})
		}
	}

	return jobs, nil
}

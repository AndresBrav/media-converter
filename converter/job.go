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
func GetJobs(inputDir string, outputDir string, format string, recursive bool) ([]Job, error) {
	var jobs []Job

	//Si no es recursivo, solo listamos los archivos en el directorio de entrada
	if !recursive {
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
				baseName := strings.TrimSuffix(f.Name(), ext)
				newFileName := baseName + format
				outputPath := filepath.Join(outputDir, newFileName)

				jobs = append(jobs, Job{
					InputPath:  inputPath,
					OutputPath: outputPath,
				})
			}
		}

		return jobs, nil
	}

	// Si es recursivo
	err := filepath.Walk(inputDir, func(inputPath string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		ext := filepath.Ext(info.Name())
		if IsSupportedFormat(ext) {
			//Calcular la ruta relativa desde inputDir
			relPath, errRel := filepath.Rel(inputDir, inputPath)
			if errRel != nil {
				return errRel
			}
			//Reemplazamos la extensión del archivo con el formato de salida
			// Calcular la ruta de salida específica (ej. resultado/foto1.webp)
			baseName := strings.TrimSuffix(relPath, ext)
			newFileName := baseName + format
			outputPath := filepath.Join(outputDir, newFileName)

			//Creamos el directorio destino si no existe
			outputFilePath := filepath.Dir(outputPath)
			if err := os.MkdirAll(outputFilePath, 0755); err != nil {
				return fmt.Errorf("error creating output directory '%s' : %w", outputFilePath, err)
			}

			jobs = append(jobs,
				Job{
					InputPath:  inputPath,
					OutputPath: outputPath,
				})
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error reading directory: '%s'\n", inputDir)
		return nil, err
	}

	return jobs, nil
}

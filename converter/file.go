package converter

import (
	"fmt"
	"os"
	"path/filepath"
)

// File almacena las rutas del archivo a procesar.
type File struct {
	InputPath  string
	OutputPath string
}

// GetFiles lee el directorio de entrada y filtra por extensiones soportadas.
func GetFiles(inputDir string, outputDir string) ([]File, error) {
	var fileAddress []File
	inputFile, err := os.ReadDir(inputDir)

	if err != nil {
		fmt.Printf("Error reading directory: '%s'\n", inputDir)
		return nil, err
	}

	for _, f := range inputFile {
		inputPath := filepath.Join(inputDir, f.Name())
		if IsSupportedFormat(filepath.Ext(f.Name())) {
			fileAddress = append(fileAddress,
				File{
					InputPath:  inputPath,
					OutputPath: outputDir,
				})
		}
	}

	return fileAddress, nil
}

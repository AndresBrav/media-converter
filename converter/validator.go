package converter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var supportedFormats = map[string]bool{
	".webp": true,
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".mkv":  true,
	".avi":  true,
	".mov":  true,
	".mp4":  true,
}

// IsSupportedFormat comprueba si una extensión está en los formatos soportados.
func IsSupportedFormat(ext string) bool {
	return supportedFormats[strings.ToLower(ext)]
}

// ValidateInput comprueba que el directorio de entrada exista y sea válido.
func ValidateInput(dir string) bool {
	inputInfo, err := os.Stat(dir)
	if err != nil {
		fmt.Printf("Input directory '%s' does not exist\n", dir)
		return false
	}

	if !inputInfo.IsDir() {
		fmt.Printf("'%s' is not a directory\n", dir)
		return false
	}

	return true
}

// PrepareOutput crea el directorio de salida si no existe y verifica permisos de escritura.
func PrepareOutput(dir string) bool {
	// Crear si no existe
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			fmt.Printf("Failed to create output directory '%s'\n", dir)
			return false
		}
		fmt.Printf("Output directory created: %s\n", dir)
	}

	// Validar que sea un directorio
	outputInfo, err := os.Stat(dir)
	if err != nil {
		fmt.Printf("Unable to access output directory '%s'\n", dir)
		return false
	}

	if !outputInfo.IsDir() {
		fmt.Printf("'%s' is not a directory\n", dir)
		return false
	}

	// Validar permisos de escritura creando un archivo temporal
	testFile := filepath.Join(dir, ".write_test")
	file, err := os.Create(testFile)
	if err != nil {
		fmt.Printf("No write permission in '%s'\n", dir)
		return false
	}
	file.Close()
	_ = os.Remove(testFile)

	return true
}

// ValidateFormat verifica si el formato destino está soportado y lo devuelve normalizado.
func ValidateFormat(fmtStr string) (string, bool) {
	normalized := fmtStr
	if !strings.HasPrefix(normalized, ".") {
		normalized = "." + normalized
	}
	normalized = strings.ToLower(normalized)

	if !supportedFormats[normalized] {
		fmt.Printf("Unsupported format '%s'\n", fmtStr)
		fmt.Println("Supported formats:")
		for f := range supportedFormats {
			fmt.Println("-", strings.TrimPrefix(f, "."))
		}
		return "", false
	}
	return normalized, true
}

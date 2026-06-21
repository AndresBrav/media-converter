package converter

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	gowebp "github.com/mayahiro/go-webp"
	"github.com/disintegration/imaging"
	"golang.org/x/image/webp"
)

func convert(job Job) {
	formatInput := strings.ToLower(filepath.Ext(job.InputPath))
	formatOutput := strings.ToLower(filepath.Ext(job.OutputPath))

	var img image.Image
	var err error

	// --- Decodificación ---
	if formatInput == ".webp" {
		// Usar el decodificador puro Go de golang.org/x/image/webp
		f, errOpen := os.Open(job.InputPath)
		if errOpen != nil {
			fmt.Printf("Error abriendo archivo %s: %v\n", job.InputPath, errOpen)
			return
		}
		defer f.Close()
		img, err = webp.Decode(f)
	} else {
		img, err = imaging.Open(job.InputPath)
	}

	if err != nil {
		fmt.Printf("Error leyendo imagen %s: %v\n", job.InputPath, err)
		return
	}

	// --- Codificación ---
	switch formatOutput {
	case ".jpg", ".jpeg", ".png":
		err = imaging.Save(img, job.OutputPath)
		if err != nil {
			fmt.Printf("Error guardando imagen %s: %v\n", job.OutputPath, err)
		} else {
			fmt.Printf("✓ %s -> %s\n", job.InputPath, job.OutputPath)
		}
	case ".webp":
		// Usar el codificador puro Go de mayahiro/go-webp
		f, errCreate := os.Create(job.OutputPath)
		if errCreate != nil {
			fmt.Printf("Error creando archivo %s: %v\n", job.OutputPath, errCreate)
			return
		}
		defer f.Close()
		err = gowebp.Encode(f, img, nil)
		if err != nil {
			fmt.Printf("Error codificando webp %s: %v\n", job.OutputPath, err)
		} else {
			fmt.Printf("✓ %s -> %s\n", job.InputPath, job.OutputPath)
		}
	default:
		fmt.Println("Formato no soportado:", formatOutput)
	}
}
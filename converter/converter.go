package converter

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	deepwebp "github.com/deepteams/webp"
	"github.com/disintegration/imaging"
	"golang.org/x/image/webp"
)

var videoInputFormats = map[string]bool{
	".mkv": true,
	".avi": true,
	".mov": true,
	".mp4": true,
}

// IsVideoInput comprueba si una extensión corresponde a un formato de video.
func IsVideoInput(ext string) bool {
	return videoInputFormats[strings.ToLower(ext)]
}

func convert(job Job, quality int) error {
	formatInput := strings.ToLower(filepath.Ext(job.InputPath))
	formatOutput := strings.ToLower(filepath.Ext(job.OutputPath))

	// convertir video co ffmpeg
	if IsVideoInput(formatInput) && formatOutput == ".mp4" {
		cmd := exec.Command("ffmpeg",
			"-i", job.InputPath,
			"-c:v", "libx264",
			"-c:a", "aac",
			"-y",
			job.OutputPath,
		)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("error convirtiendo video %s:\n%s%w", job.InputPath, string(output), err)
		}
		return nil
	}

	var img image.Image
	var err error

	// --- Decodificación ---
	if formatInput == ".webp" {
		// Usar el decodificador puro Go de golang.org/x/image/webp
		f, errOpen := os.Open(job.InputPath)
		if errOpen != nil {
			return fmt.Errorf("error abriendo archivo %s: %w", job.InputPath, errOpen)
		}
		defer f.Close()
		img, err = webp.Decode(f)
	} else {
		img, err = imaging.Open(job.InputPath)
	}

	if err != nil {
		return fmt.Errorf("error leyendo imagen %s: %w", job.InputPath, err)
	}

	// --- Codificación ---
	switch formatOutput {
	case ".jpg", ".jpeg":
		err = imaging.Save(img, job.OutputPath, imaging.JPEGQuality(quality))
		if err != nil {
			return fmt.Errorf("error guardando imagen %s: %w", job.OutputPath, err)
		}
	case ".png":
		err = imaging.Save(img, job.OutputPath)
		if err != nil {
			return fmt.Errorf("error guardando imagen %s: %w", job.OutputPath, err)
		}
	case ".webp":
		// Usar el codificador puro Go de github.com/deepteams/webp
		f, errCreate := os.Create(job.OutputPath)
		if errCreate != nil {
			return fmt.Errorf("error creando archivo %s: %w", job.OutputPath, errCreate)
		}
		defer f.Close()
		err = deepwebp.Encode(f, img, &deepwebp.EncoderOptions{
			Quality: float32(quality),
		})
		if err != nil {
			return fmt.Errorf("error codificando webp %s: %w", job.OutputPath, err)
		}
	default:
		return fmt.Errorf("formato no soportado: %s", formatOutput)
	}

	return nil
}

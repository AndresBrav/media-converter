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

// Options contiene las opciones de procesamiento para cada imagen.
type Options struct {
	Quality   int
	Width     int
	Height    int
	Watermark string
	Thumbnail bool
}

// IsVideoInput comprueba si una extensión corresponde a un formato de video.
func IsVideoInput(ext string) bool {
	return videoInputFormats[strings.ToLower(ext)]
}

func saveImage(img image.Image, outputPath string, formatOutput string, quality int) error {
	switch formatOutput {
	case ".jpg", ".jpeg":
		return imaging.Save(img, outputPath, imaging.JPEGQuality(quality))
	case ".png":
		return imaging.Save(img, outputPath)
	case ".webp":
		f, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("error creando archivo %s: %w", outputPath, err)
		}
		defer f.Close()
		return deepwebp.Encode(f, img, &deepwebp.EncoderOptions{
			Quality: float32(quality),
		})
	default:
		return fmt.Errorf("formato no soportado: %s", formatOutput)
	}
}

func convert(job Job, opts Options) error {
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

	// Decodificación
	if formatInput == ".webp" {
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

	// Resize
	if opts.Width > 0 || opts.Height > 0 {
		if opts.Width > 0 && opts.Height > 0 {
			img = imaging.Fill(img, opts.Width, opts.Height, imaging.Center, imaging.Lanczos)
		} else {
			img = imaging.Resize(img, opts.Width, opts.Height, imaging.Lanczos)
		}
	}

	// Watermark
	if opts.Watermark != "" {
		wmImg, errWm := imaging.Open(opts.Watermark)
		if errWm != nil {
			return fmt.Errorf("error abriendo watermark %s: %w", opts.Watermark, errWm)
		}
		wmWidth := img.Bounds().Dx() / 10
		wmImg = imaging.Resize(wmImg, wmWidth, 0, imaging.Lanczos)
		posX := img.Bounds().Dx() - wmImg.Bounds().Dx() - 10
		posY := img.Bounds().Dy() - wmImg.Bounds().Dy() - 10
		img = imaging.Overlay(img, wmImg, image.Pt(posX, posY), 0.5)
	}

	// Codificación
	if err := saveImage(img, job.OutputPath, formatOutput, opts.Quality); err != nil {
		return fmt.Errorf("error guardando imagen %s: %w", job.OutputPath, err)
	}

	// Thumbnail
	if opts.Thumbnail {
		thumbExt := filepath.Ext(job.OutputPath)
		thumbPath := strings.TrimSuffix(job.OutputPath, thumbExt) + "_thumb" + thumbExt
		thumbImg := imaging.Thumbnail(img, 150, 150, imaging.Lanczos)
		if err := saveImage(thumbImg, thumbPath, formatOutput, opts.Quality); err != nil {
			return fmt.Errorf("error guardando thumbnail %s: %w", thumbPath, err)
		}
	}

	return nil
}

package converter

import (
    "fmt"
    "image"
    "os"
    "path/filepath"

    "github.com/chai2010/webp"
    "github.com/disintegration/imaging"
)

func convert(job Job) {
	formatInput := filepath.Ext(job.InputPath)
	formatOutput := filepath.Ext(job.OutputPath)

	var img image.Image
	var err error

	if formatInput == ".webp" {
		f,_ := os.Open(job.InputPath)
		defer f.Close()
		img, err = webp.Decode(f)
	}else{
		img, err = imaging.Open(job.InputPath)
	}

	if err != nil {
		fmt.Println("Error leyendo imagen:", err)
		return
	}

	switch formatOutput {
	case ".jpg", ".jpeg", ".png":
		imaging.Save(img, job.OutputPath)
	case ".webp":
		f, _ := os.Create(job.OutputPath)
		defer f.Close()
		webp.Encode(f, img, nil)
	default:
		fmt.Println("Formato no soportado", formatOutput)
	}
}
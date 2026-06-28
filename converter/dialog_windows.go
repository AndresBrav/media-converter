//go:build windows

package converter

import (
	"fmt"
	"os/exec"
	"strings"
	"syscall"
)

func browseDialog(dialogType string) (string, error) {
	var script string
	if dialogType == "directory" {
		script = `
		Add-Type -AssemblyName System.Windows.Forms
		$f = New-Object System.Windows.Forms.FolderBrowserDialog
		$f.Description = "Selecciona una carpeta"
		$f.ShowNewFolderButton = $true
		$result = $f.ShowDialog()
		if ($result -eq [System.Windows.Forms.DialogResult]::OK) {
			$f.SelectedPath
		} else {
			"CANCELLED"
		}
		`
	} else {
		script = `
		Add-Type -AssemblyName System.Windows.Forms
		$f = New-Object System.Windows.Forms.OpenFileDialog
		$f.Filter = "Imágenes (*.png;*.jpg;*.jpeg;*.webp)|*.png;*.jpg;*.jpeg;*.webp"
		$f.Title = "Selecciona una imagen de marca de agua"
		$result = $f.ShowDialog()
		if ($result -eq [System.Windows.Forms.DialogResult]::OK) {
			$f.FileName
		} else {
			"CANCELLED"
		}
		`
	}

	cmd := exec.Command("powershell", "-NoProfile", "-Command", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	result := strings.TrimSpace(string(output))
	if result == "CANCELLED" || result == "" {
		return "", fmt.Errorf("selección cancelada")
	}
	return result, nil
}

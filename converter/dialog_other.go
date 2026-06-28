//go:build !windows

package converter

import "fmt"

func browseDialog(dialogType string) (string, error) {
	return "", fmt.Errorf("diálogo de selección nativo solo disponible en Windows")
}

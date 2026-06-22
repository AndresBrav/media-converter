package converter

import (
	"fmt"
	"path/filepath"
	"sync"
	"sync/atomic"
)

// Worker escucha el canal de Jobs y ejecuta la conversión de cada uno.
func Worker(id int, jobs chan Job, waitGroup *sync.WaitGroup, completed *int32, failed *int32, total int) {
	fmt.Printf("Worker %d started\n", id)
	for job := range jobs {
		err := convert(job)
		current := atomic.AddInt32(completed, 1)
		inputName := filepath.Base(job.InputPath)
		outputName := filepath.Base(job.OutputPath)
		if err != nil {
			atomic.AddInt32(failed, 1)
			fmt.Printf("[%d/%d] ✗ Error convirtiendo %s: %v\n", current, total, inputName, err)
		} else {
			fmt.Printf("[%d/%d] ✓ %s -> %s\n", current, total, inputName, outputName)
		}
	}
	waitGroup.Done()
}
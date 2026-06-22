package converter

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// Worker escucha el canal de Jobs y ejecuta la conversión de cada uno.
func Worker(jobs chan Job, waitGroup *sync.WaitGroup, completed *int32, total int) {
	for job := range jobs {
		err := convert(job)
		current := atomic.AddInt32(completed, 1)
		if err != nil {
			fmt.Printf("[%d/%d] ✗ %v\n", current, total, err)
		} else {
			fmt.Printf("[%d/%d] ✓ %s -> %s\n", current, total, job.InputPath, job.OutputPath)
		}
	}
	waitGroup.Done()
}
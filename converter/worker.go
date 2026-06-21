package converter

import "sync"

// Worker escucha el canal de Jobs y ejecuta la conversión de cada uno.
func Worker(jobs chan Job, waitGroup *sync.WaitGroup) {
	for job := range jobs {
		convert(job)
	}
	waitGroup.Done()
}
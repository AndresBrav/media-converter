package converter

import "sync"

//Falta la estrestura de Job o se puede usar el File
func Worker(jobs chan Job, waitGroup *sync.WaitGroup) {

	for job := range jobs {
		convert(job)//asi se llamara la funcion de conversor 
	}
	waitGroup.Done()
}
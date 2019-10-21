package ravelry

import (
	"fmt"
	"sync"
)

// QueryRunner launches workers to run *nWorkers concurrent queries
type QueryRunner struct {
	nWorkers int
	name     string
	wg       sync.WaitGroup
	work     func() bool // worker return true when all work is complete
	close    func()      // close output channel of indeterminate type
}

// work should read input, determine if there is more work to do,
// and then return a function to do the work
func NewQueryRunner(nWorkers int, work func() (bool, func()), close func()) *QueryRunner {
	var wg sync.WaitGroup
	return &QueryRunner{
		nWorkers: nWorkers,
		wg:       wg,
		work: func() bool {
			doneReading, doWork := work()
			if doneReading {
				return true
			}
			wg.Add(1)
			defer wg.Done()
			doWork()
			return false
		},
		close: close,
	}
}

func (qr *QueryRunner) Run() {
	allJobsStarted := make(chan struct{})

	for i := 0; i < qr.nWorkers; i++ {
		go func() {
			for {
				workDone := qr.work()
				if workDone {
					fmt.Println("work done")
					allJobsStarted <- struct{}{}
					return
				}
			}
		}()
	}
	go func() {
		for i := 0; i < qr.nWorkers; i++ {
			<-allJobsStarted
		}
		qr.wg.Wait()
		fmt.Printf("Closing %s channel\n", qr.name)
		qr.close()
	}()
}

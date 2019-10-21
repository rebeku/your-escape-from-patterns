package ravelry

import (
	"fmt"
	"sort"
	"sync"
	"testing"
)

func TestQueryRunner(t *testing.T) {
	in := make(chan int)
	out := make(chan string, 5)

	work := func() (bool, func()) {
		n, more := <-in
		return !more, func() {
			out <- fmt.Sprintf("%d", n)
		}
	}

	closef := func() {
		fmt.Println("Closing out")
		close(out)
	}

	qr := NewQueryRunner(10, work, closef)
	qr.Run()

	nIn := 30000

	expected := make([]string, 0, nIn)
	var wg sync.WaitGroup
	wg.Add(nIn)

	go func() {
		for i := 0; i < nIn; i++ {
			in <- i
			expected = append(expected, fmt.Sprintf("%d", i))
		}
	}()

	outStrings := make([]string, 0, nIn)
	go func() {
		for {
			s := <-out
			outStrings = append(outStrings, s)
			wg.Done()
		}
	}()

	wg.Wait()

	sort.StringSlice(expected).Sort()
	sort.StringSlice(outStrings).Sort()
	for i, s := range outStrings {
		if s != expected[i] {
			t.Errorf("Unexpected result %d: Expected %s but found %s", i, expected[i], s)
		}
	}
}

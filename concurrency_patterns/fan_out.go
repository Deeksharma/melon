package concurrency_patterns

import (
	"fmt"
	"sync"
)

func Split(source <-chan int, n int) []<-chan int {
	dests := make([]<-chan int, 0) // Create the dests slice

	for i := 0; i < n; i++ { // Create n destination channels

		ch := make(chan int)
		dests = append(dests, ch)

		// Each channel gets a dedicated
		// goroutine that competes for reads
		go func() {
			defer close(ch)
			for val := range source {
				ch <- val
			}
		}()
	}
	return dests
}

func TestingFanOut() {
	source := make(chan int)  // The input channel
	dests := Split(source, 5) // Retrieve 5 output channels
	go func() {
		for i := 1; i <= 10; i++ { // Send the number 1..10 to source
			source <- i
		}
		close(source) // and close it when we're done
	}()

	//dest := Funnel(dests...)
	var wg sync.WaitGroup
	wg.Add(len(dests)) // Use WaitGroup to wait until the output channels all close
	for i, ch := range dests {
		go func(i int, d <-chan int) {
			defer wg.Done()
			for val := range d {
				fmt.Printf("#%d got %d\n", i, val) // i - which destination is printing the res
			}
		}(i, ch)
	}
	wg.Wait()
}

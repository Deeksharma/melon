package concurrency_patterns

import (
	"fmt"
	"sync"
	"time"
)

func Funnel(sources ...<-chan int) <-chan int {
	dest := make(chan int) // The shared output channel

	// Used to automatically close dest
	// when all sources are closed
	var wg sync.WaitGroup

	// Set size of the WaitGroup
	wg.Add(len(sources))

	// Start a goroutine for each source
	for _, ch := range sources {
		go func(c <-chan int) {
			defer wg.Done() // Notify WaitGroup when c closes

			for n := range c {
				dest <- n
			}
		}(ch)
	}
	go func() {
		wg.Wait()   // Start a goroutine to close dest
		close(dest) // after all sources close
	}()
	return dest
}

func TestingFanIn() {
	sources := make([]<-chan int, 0) // Create an empty channel slice

	for i := 0; i < 3; i++ {
		ch := make(chan int)
		sources = append(sources, ch) // Create an empty channel slice

		go func() { // Run a toy goroutine for each
			defer close(ch) // Close ch when the routine ends
			for j := 1; j <= 5; j++ {
				ch <- j
				time.Sleep(time.Second)
			}
		}()
	}
	dest := Funnel(sources...)
	for d := range dest {
		fmt.Println(d)
	}
}

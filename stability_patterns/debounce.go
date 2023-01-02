package stability_patterns

import (
	"context"
	"sync"
	"time"
)

func DebounceFirst(circuit Circuit, d time.Duration) Circuit {
	var threshold time.Time
	var result string
	var err error
	var m sync.Mutex
	return func(ctx context.Context) (string, error) {
		m.Lock()
		defer func() {
			threshold = time.Now().Add(d) // the threshold will keep increasing until no request comes within the threshold
			m.Unlock()
		}()
		if time.Now().Before(threshold) {
			return result, err
		}
		result, err = circuit(ctx)
		return result, err
	}
}

func DebounceLast(circuit Circuit, d time.Duration) Circuit {
	var threshold time.Time = time.Now()
	var ticker *time.Ticker
	var result string
	var err error
	var once sync.Once
	var m sync.Mutex
	return func(ctx context.Context) (string, error) {
		m.Lock()
		defer m.Unlock()              // this defer is for the function
		threshold = time.Now().Add(d) // this duration keeps getting added to the last request, after the last request
		once.Do(func() {              // for all the functions this will be run once
			ticker = time.NewTicker(time.Millisecond * 100)
			go func() {
				defer func() { // this defer is for the goroutine
					m.Lock()
					ticker.Stop()
					once = sync.Once{}
					m.Unlock()
				}()
				for {
					select {
					case <-ticker.C: // channel where the ticks are delivered, ticker.C
						m.Lock()
						if time.Now().After(threshold) { // if the threshold is not passed then, before line 67 return, this goroutine will be closed and the new set os sync.Once
							result, err = circuit(ctx)
							m.Unlock()
							return // this will kill goroutine
						}
						m.Unlock()
					case <-ctx.Done(): // when this will be done - if the context is canceled ot the whole work is completed
						m.Lock()
						result, err = "", ctx.Err()
						m.Unlock()
						return
					}
				}
			}()
		})
		return result, err
	}
}

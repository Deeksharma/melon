package stability_patterns

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func Throttle(e Effector, max uint, refill uint, d time.Duration) Effector {
	var tokens = max
	var once sync.Once
	return func(ctx context.Context) (string, error) {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		once.Do(func() {
			ticker := time.NewTicker(d)
			go func() {
				defer ticker.Stop()
				for {
					select {
					case <-ctx.Done():
						return
					case <-ticker.C:
						t := tokens + refill // refill tokens after some number of times
						if t > max {
							t = max
						}
						tokens = t
					}
				}
			}()
		})
		if tokens <= 0 {
			return "", fmt.Errorf("too many calls")
		}
		tokens--
		return e(ctx)
	}
}

package stability_patterns

import (
	"context"
	"errors"
	"sync"
	"time"
)

type Circuit func(context.Context) (string, error)

func Breaker(circuit Circuit, failureThreshold uint) Circuit {
	var consecutiveFailures int = 0
	var lastAttempt = time.Now()
	var m sync.RWMutex
	return func(ctx context.Context) (string, error) {
		m.RLock() // Establish a "read lock"
		d := consecutiveFailures - int(failureThreshold)
		if d >= 0 {
			shouldRetryAt := lastAttempt.Add(time.Second * 2 << d) // exponential backoff
			if !time.Now().After(shouldRetryAt) {
				m.RUnlock()
				return "", errors.New("service unreachable")
			}
		}
		m.RUnlock()                   // Release read lock
		response, err := circuit(ctx) // Issue request proper
		m.Lock()                      // Lock around shared resources
		defer m.Unlock()
		lastAttempt = time.Now() // Record time of attempt
		// Circuit returned an error,
		// so we count the failure
		// and return
		if err != nil {
			consecutiveFailures++
			return response, err
		}
		consecutiveFailures = 0 // Reset failures counter
		return response, nil
	}
}

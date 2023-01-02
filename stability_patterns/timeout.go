package stability_patterns

import (
	"context"
	"fmt"
	"time"
)

type SlowFunction func(string) (string, error)

type WithContext func(context.Context, string) (string, error)

func Timeout(f SlowFunction) WithContext {
	return func(ctx context.Context, arg string) (string, error) {
		chres := make(chan string)
		cherr := make(chan error)
		go func() {
			res, err := f(arg)
			chres <- res
			cherr <- err
		}()
		select {
		case res := <-chres:
			return res, <-cherr
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
}

func Slow(string string) (string, error) {
	return "", nil
}
func Testing() {
	ctx := context.Background()
	ctxt, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	timeout := Timeout(Slow) // this is very important to get the closur function first, other the slow will try to execute itself first
	res, err := timeout(ctxt, "some input")
	fmt.Println(res, err)
}

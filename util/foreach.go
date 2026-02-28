package util

import (
	"errors"
	"sync"
)

func ForEach[S ~[]E, E any](s S, f func(E) error) error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errs []error

	for _, e := range s {
		wg.Go(func() {
			err := f(e)
			mu.Lock()
			defer mu.Unlock()
			errs = append(errs, err)
		})
	}

	wg.Wait()
	return errors.Join(errs...)
}

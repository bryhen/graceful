package graceful

import "context"

// Provides convenience wrapper to run multiple Funcs concurrently by Start().
//
// Useful for speeding up startup or shutdown, where each function does not depend on any preceding step.
// For example, you may want to concurrently initialize connections to a db, cache, config service, etc at the same time.
//
// Order is not guaranteed for these functions. All provided functions must return before the returned function returns.
//
// Only the first error (if any) received will be reported.
func Multi(fns ...Func) Func {
	return func(ctx context.Context) error {
		errCh := make(chan error, len(fns))

		// Run functions concurrently
		for _, fn := range fns {
			go func(f Func) {
				errCh <- f(ctx)
			}(fn)
		}

		// Read their errors and store the first (if any)
		var err error
		for range len(fns) {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case e := <-errCh:
				if err == nil && e != nil {
					err = e
				}
			}
		}

		return err
	}
}

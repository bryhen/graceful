package graceful

// Signals that the application should exit. Passes the provided error, which can be nil, to unblock Run().
//
// Only the first error passed to Shutdown() will be propogated. It is safe to call concurrently.
//
// This function should be called by scripts that have completed successfully (with nil) or applications that have an encountered an error requiring shutdown (with a non-nil error).
func Shutdown(err error) {
	select {
	case rte <- err:
	default:
	}
}

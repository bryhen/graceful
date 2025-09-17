package graceful

import (
	"fmt"
	"os"
	"time"
)

func parseOptions(config *config, opts []*option) error {
	for _, opt := range opts {
		switch opt.code {
		case optionStartupTimeout:
			if v, ok := opt.value.(time.Duration); ok {
				if v < 1 {
					return fmt.Errorf("startup timeout must be positive")
				}
				config.startupTimeout = v
			} else {
				return fmt.Errorf("failed to cast StartupTimeout to time.Duration")
			}

		case optionShutdownTimeout:
			if v, ok := opt.value.(time.Duration); ok {
				if v < 1 {
					return fmt.Errorf("shutdown timeout must be positive")
				}
				config.shutdownTimeout = v
			} else {
				return fmt.Errorf("failed to cast StartupTimeout to time.Duration")
			}

		case optionSignals:
			if sigs, ok := opt.value.([]os.Signal); ok {
				config.signals = append(config.signals, sigs...)
			} else {
				return fmt.Errorf("failed to cast signals")
			}
		}
	}

	return nil
}

// Maximum amount of time to wait for startup functions to complete before calling a timeout err. Default: unlimited.
func WithStartupTimeout(d time.Duration) *option {
	return &option{
		code:  optionStartupTimeout,
		value: d,
	}
}

// Maximum amount of time to wait for shutdown functions to complete before calling a timeout err. Default: unlimited.
func WithShutdownTimeout(d time.Duration) *option {
	return &option{
		code:  optionShutdownTimeout,
		value: d,
	}
}

// These signals will trigger a shutdown. Default: os.Interrupt, syscall.SIGINT, syscall.SIGTERM.
func WithSignals(sigs []os.Signal) *option {
	return &option{
		code:  optionSignals,
		value: sigs,
	}
}

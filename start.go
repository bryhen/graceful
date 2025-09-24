package graceful

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Contains information about why the program exited.
type ExitReason struct {
	OsSignal     os.Signal
	ErrStartup   error
	ErrRuntime   error
	ErrsShutdown []error
}

type ExitReasonPrintable struct {
	OsSignal     string   `json:"osSignal"`
	ErrStartup   string   `json:"errStartup"`
	ErrRuntime   string   `json:"errRuntime"`
	ErrsShutdown []string `json:"errsShutdown"`
}

func (er *ExitReason) ToPrintable() *ExitReasonPrintable {
	erp := &ExitReasonPrintable{}

	if er.OsSignal != nil {
		erp.OsSignal = er.OsSignal.String()
	}

	if er.ErrStartup != nil {
		erp.ErrStartup = er.ErrStartup.Error()
	}

	if er.ErrRuntime != nil {
		erp.ErrRuntime = er.ErrRuntime.Error()
	}

	for _, e := range er.ErrsShutdown {
		erp.ErrsShutdown = append(erp.ErrsShutdown, e.Error())
	}

	return erp
}

// Marshals the struct.
func (er *ExitReason) MarshalStr() string {
	bs, _ := json.Marshal(er.ToPrintable())
	return string(bs)
}

// Marshals the struct with indents.
// prefix is usually "" and indent is usually "\t"
func (er *ExitReason) MarshalIndentStr(prefix string, indent string) string {
	bs, _ := json.MarshalIndent(er.ToPrintable(), prefix, indent)
	return string(bs)
}

type Func func(ctx context.Context) error

type option struct {
	code  int
	value any
}

type config struct {
	shutdownTimeout time.Duration
	startupTimeout  time.Duration
	signals         []os.Signal
}

const (
	optionStartupTimeout  = 1
	optionShutdownTimeout = 2
	optionSignals         = 10
)

var (
	rte = make(chan error, 1)
)

// Helps run an application by handling graceful startup and shutdown.
//
// Returns the guaranteed non-nil ExitReason struct which contains information about why the program exited.
//
// This function will:
//
// 1. Run startup functions sequentially.
//   - If any of these functions returns an error, this function will return immediately.
//
// 2. Blocks until a runtime signal (from Shutdown()) or specified OS signal (ie ctrl+c) is received.
//   - Only the first runtime error received (if any) will be returned. All others are discarded.
//   - Default signals monitored are os.Interrupt, syscall.SIGINT, and syscall.SIGTERM.
//
// 3. Run the shutdown functions sequentially.
func Start(startupFns []Func, shutdownFns []Func, opts ...*option) *ExitReason {
	er := &ExitReason{}
	fnErrs := make(chan error, 1)

	config := &config{
		signals: []os.Signal{os.Interrupt, syscall.SIGINT, syscall.SIGTERM},
	}
	if err := parseOptions(config, opts); err != nil {
		er.ErrStartup = err
		return er
	}

	// Start the application and exit early if any errors occur.
	stCtx, stCancel := context.Background(), nop
	if config.startupTimeout > 0 {
		stCtx, stCancel = context.WithTimeout(stCtx, config.startupTimeout)
	}

	go func() {

		for _, fn := range startupFns {
			if err := fn(stCtx); err != nil {
				fnErrs <- err
				stCancel()
				return
			}
		}

		fnErrs <- nil
	}()

	select {
	case er.ErrStartup = <-fnErrs:
	case <-stCtx.Done():
		er.ErrStartup = stCtx.Err()
	}

	stCancel()

	if er.ErrStartup != nil {
		return er
	}

	// Monitor the application/OS and document why we're shutting down.
	osSig := make(chan os.Signal, 1)
	signal.Notify(osSig, config.signals...)

	select {
	case er.ErrRuntime = <-rte:
	case er.OsSignal = <-osSig:
	}

	// Shutdown the application and collect all the errors that occurred during shutdown.
	sdCtx, sdCancel := context.Background(), nop
	if config.shutdownTimeout > 0 {
		sdCtx, sdCancel = context.WithTimeout(sdCtx, config.shutdownTimeout)
	}
	defer sdCancel()

	go func() {
		for _, fn := range shutdownFns {
			fnErrs <- fn(sdCtx)
		}
	}()

Shutdown:
	for range shutdownFns {
		select {
		case e := <-fnErrs:
			if e != nil {
				er.ErrsShutdown = append(er.ErrsShutdown, e)
			}
		case <-sdCtx.Done():
			er.ErrsShutdown = append(er.ErrsShutdown, sdCtx.Err())
			break Shutdown
		}
	}

	return er
}

func nop() {}

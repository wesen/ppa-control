package debouncer

import (
	"sync"
	"time"
)

type Debouncer struct {
	interval      time.Duration
	timer         *time.Timer
	mu            sync.Mutex
	fn            func()
	lastExecution time.Time
}

func NewDebouncer(interval time.Duration) *Debouncer {
	return &Debouncer{
		interval: interval,
	}
}

func (d *Debouncer) Run(f func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// If a timer already exists, stop it
	if d.timer != nil {
		d.timer.Stop()
	}

	// Set the function to be executed
	d.fn = f

	// Calculate the time since the last execution
	timeSinceLastExec := time.Since(d.lastExecution)

	// Determine the delay for the next execution
	delay := d.interval - timeSinceLastExec
	if delay < 0 {
		delay = 0
	}

	// Schedule the function to be executed after the calculated delay
	d.timer = time.AfterFunc(delay, func() {
		d.mu.Lock()
		defer d.mu.Unlock()

		if d.fn != nil {
			d.fn()
			// Update the last execution time
			d.lastExecution = time.Now()
			// Reset the function after execution
			d.fn = nil
		}
	})
}

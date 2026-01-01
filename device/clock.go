package device

/*
#cgo LDFLAGS: -L../rust/target/release -lbeefdown_clock
#include "../rust/beefdown_clock.h"
#include <stdlib.h>

extern void clockTickCallback(void* userData);
*/
import "C"
import (
	"fmt"
	"sync"
	"unsafe"
)

// Global registry for Clock instances (avoids passing Go pointers to C)
var (
	clockRegistry   = make(map[uintptr]*Clock)
	clockRegistryMu sync.RWMutex
	nextClockID     uintptr = 1
)

// Clock wraps the Rust high-precision clock
type Clock struct {
	clock    *C.Clock
	callback func()
	mu       sync.Mutex
	running  bool
	id       uintptr
}

// NewClock creates a new Rust-backed clock
// bpm: beats per minute
func NewClock(bpm float64) (*Clock, error) {
	clock := C.clock_new(C.double(bpm))
	if clock == nil {
		return nil, fmt.Errorf("failed to create Rust clock")
	}

	clockRegistryMu.Lock()
	id := nextClockID
	nextClockID++
	clockRegistryMu.Unlock()

	c := &Clock{
		clock:   clock,
		running: false,
		id:      id,
	}

	clockRegistryMu.Lock()
	clockRegistry[id] = c
	clockRegistryMu.Unlock()

	return c, nil
}

// Start the clock with a callback that fires on each tick (24ppq)
func (c *Clock) Start(callback func()) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running {
		return fmt.Errorf("clock already running")
	}

	c.callback = callback

	// Pass the clock ID as user data (not a Go pointer!)
	//
	// Note: go vet warns about this unsafe.Pointer conversion, but it's safe because:
	// 1. c.id is an integer ID (not a real pointer)
	// 2. C side treats it as opaque user data (void*)
	// 3. Callback converts it back to uintptr and looks up the Clock in clockRegistry
	// 4. No actual Go pointers are passed to C (avoiding Go's CGo pointer rules)
	result := C.clock_start(c.clock, C.tick_callback(C.clockTickCallback), unsafe.Pointer(c.id))

	if result != 0 {
		return fmt.Errorf("failed to start clock")
	}

	c.running = true
	return nil
}

// Stop the clock
func (c *Clock) Stop() error {
	c.mu.Lock()
	if !c.running {
		c.mu.Unlock()
		return nil // Already stopped
	}
	c.running = false
	c.mu.Unlock()

	// Don't hold the mutex while stopping - the Rust thread needs to acquire
	// the mutex in the callback, so holding it here would cause a deadlock
	result := C.clock_stop(c.clock)
	if result != 0 {
		return fmt.Errorf("failed to stop clock")
	}

	return nil
}

// SetBPM updates the clock tempo (can be called while running)
func (c *Clock) SetBPM(bpm float64) error {
	result := C.clock_set_bpm(c.clock, C.double(bpm))
	if result != 0 {
		return fmt.Errorf("failed to set BPM")
	}
	return nil
}

// Close frees the clock resources
func (c *Clock) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.clock != nil {
		// Stop first if running
		if c.running {
			C.clock_stop(c.clock)
			c.running = false
		}
		C.clock_free(c.clock)
		c.clock = nil

		// Remove from registry
		clockRegistryMu.Lock()
		delete(clockRegistry, c.id)
		clockRegistryMu.Unlock()
	}
}

//export clockTickCallback
func clockTickCallback(userData unsafe.Pointer) {
	// Convert the pointer back to the clock ID
	id := uintptr(userData)

	// Look up the clock in the registry
	clockRegistryMu.RLock()
	c, ok := clockRegistry[id]
	clockRegistryMu.RUnlock()

	if !ok {
		return // Clock was closed/removed
	}

	c.mu.Lock()
	callback := c.callback
	c.mu.Unlock()

	if callback != nil {
		callback()
	}
}

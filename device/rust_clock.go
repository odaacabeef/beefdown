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

// Global registry for RustClock instances (avoids passing Go pointers to C)
var (
	clockRegistry   = make(map[uintptr]*RustClock)
	clockRegistryMu sync.RWMutex
	nextClockID     uintptr = 1
)

// RustClock wraps the Rust high-precision clock
type RustClock struct {
	clock    *C.Clock
	callback func()
	mu       sync.Mutex
	running  bool
	id       uintptr
}

// NewRustClock creates a new Rust-backed clock
// bpm: beats per minute
func NewRustClock(bpm float64) (*RustClock, error) {
	clock := C.clock_new(C.double(bpm))
	if clock == nil {
		return nil, fmt.Errorf("failed to create Rust clock")
	}

	clockRegistryMu.Lock()
	id := nextClockID
	nextClockID++
	clockRegistryMu.Unlock()

	rc := &RustClock{
		clock:   clock,
		running: false,
		id:      id,
	}

	clockRegistryMu.Lock()
	clockRegistry[id] = rc
	clockRegistryMu.Unlock()

	return rc, nil
}

// Start the clock with a callback that fires on each tick (24ppq)
func (rc *RustClock) Start(callback func()) error {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if rc.running {
		return fmt.Errorf("clock already running")
	}

	rc.callback = callback

	// Pass the clock ID as user data (not a Go pointer!)
	userData := unsafe.Pointer(rc.id)
	result := C.clock_start(rc.clock, C.tick_callback(C.clockTickCallback), userData)

	if result != 0 {
		return fmt.Errorf("failed to start clock")
	}

	rc.running = true
	return nil
}

// Stop the clock
func (rc *RustClock) Stop() error {
	rc.mu.Lock()
	if !rc.running {
		rc.mu.Unlock()
		return nil // Already stopped
	}
	rc.running = false
	rc.mu.Unlock()

	// Don't hold the mutex while stopping - the Rust thread needs to acquire
	// the mutex in the callback, so holding it here would cause a deadlock
	result := C.clock_stop(rc.clock)
	if result != 0 {
		return fmt.Errorf("failed to stop clock")
	}

	return nil
}

// SetBPM updates the clock tempo (can be called while running)
func (rc *RustClock) SetBPM(bpm float64) error {
	result := C.clock_set_bpm(rc.clock, C.double(bpm))
	if result != 0 {
		return fmt.Errorf("failed to set BPM")
	}
	return nil
}

// Close frees the clock resources
func (rc *RustClock) Close() {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if rc.clock != nil {
		// Stop first if running
		if rc.running {
			C.clock_stop(rc.clock)
			rc.running = false
		}
		C.clock_free(rc.clock)
		rc.clock = nil

		// Remove from registry
		clockRegistryMu.Lock()
		delete(clockRegistry, rc.id)
		clockRegistryMu.Unlock()
	}
}

//export clockTickCallback
func clockTickCallback(userData unsafe.Pointer) {
	// Convert the pointer back to the clock ID
	id := uintptr(userData)

	// Look up the clock in the registry
	clockRegistryMu.RLock()
	rc, ok := clockRegistry[id]
	clockRegistryMu.RUnlock()

	if !ok {
		return // Clock was closed/removed
	}

	rc.mu.Lock()
	callback := rc.callback
	rc.mu.Unlock()

	if callback != nil {
		callback()
	}
}

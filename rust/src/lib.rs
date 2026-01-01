mod timing;
mod clock;
mod midi;

pub use clock::Clock;
use std::ffi::{c_void, CStr};
use std::os::raw::c_char;

/// FFI-safe callback type for clock ticks
pub type TickCallback = extern "C" fn(*mut c_void);

/// Wrapper to make the user_data pointer Send
/// Safety: The caller must ensure the pointer remains valid
struct SendPtr(*mut c_void);
unsafe impl Send for SendPtr {}

impl SendPtr {
    fn as_ptr(&self) -> *mut c_void {
        self.0
    }
}

/// Create a new clock with the given BPM
/// Returns an opaque pointer to the clock
#[no_mangle]
pub extern "C" fn clock_new(bpm: f64) -> *mut Clock {
    let clock = Box::new(Clock::new(bpm));
    Box::into_raw(clock)
}

/// Start the clock with a callback that fires on each tick (24ppq)
/// The callback receives the user_data pointer
/// Returns 0 on success, -1 on error
#[no_mangle]
pub extern "C" fn clock_start(
    clock: *mut Clock,
    callback: TickCallback,
    user_data: *mut c_void,
) -> i32 {
    if clock.is_null() {
        return -1;
    }

    let clock = unsafe { &mut *clock };
    let user_data = SendPtr(user_data);

    match clock.start(move || {
        callback(user_data.as_ptr());
    }) {
        Ok(_) => 0,
        Err(_) => -1,
    }
}

/// Stop the clock
/// Returns 0 on success, -1 on error
#[no_mangle]
pub extern "C" fn clock_stop(clock: *mut Clock) -> i32 {
    if clock.is_null() {
        return -1;
    }

    let clock = unsafe { &mut *clock };

    match clock.stop() {
        Ok(_) => 0,
        Err(_) => -1,
    }
}

/// Set the BPM of the clock (can be called while running)
/// Returns 0 on success, -1 on error
#[no_mangle]
pub extern "C" fn clock_set_bpm(clock: *mut Clock, bpm: f64) -> i32 {
    if clock.is_null() {
        return -1;
    }

    let clock = unsafe { &mut *clock };
    clock.set_bpm(bpm);
    0
}

/// Free the clock
#[no_mangle]
pub extern "C" fn clock_free(clock: *mut Clock) {
    if !clock.is_null() {
        unsafe {
            let _ = Box::from_raw(clock);
        }
    }
}

// ============================================================================
// MIDI FFI Functions
// ============================================================================

/// FFI-safe callback type for MIDI input
pub type MidiInputCallback = extern "C" fn(*mut c_void, *const u8, usize, i64);

/// Create a virtual MIDI output port
/// Returns port ID on success, -1 on error
#[no_mangle]
pub extern "C" fn midi_create_virtual_output(name: *const c_char) -> i32 {
    if name.is_null() {
        return -1;
    }

    let name = unsafe {
        match CStr::from_ptr(name).to_str() {
            Ok(s) => s,
            Err(_) => return -1,
        }
    };

    match midi::create_virtual_output(name) {
        Ok(id) => id,
        Err(_) => -1,
    }
}

/// Connect to an existing MIDI output port by name
/// Returns port ID on success, -1 on error
#[no_mangle]
pub extern "C" fn midi_connect_output(name: *const c_char) -> i32 {
    if name.is_null() {
        return -1;
    }

    let name = unsafe {
        match CStr::from_ptr(name).to_str() {
            Ok(s) => s,
            Err(_) => return -1,
        }
    };

    match midi::connect_output(name) {
        Ok(id) => id,
        Err(_) => -1,
    }
}

/// Send MIDI message bytes to an output port
/// Returns 0 on success, -1 on error
#[no_mangle]
pub extern "C" fn midi_send(port_id: i32, bytes: *const u8, len: usize) -> i32 {
    if bytes.is_null() || len == 0 || len > 3 {
        return -1;
    }

    let bytes = unsafe { std::slice::from_raw_parts(bytes, len) };

    match midi::send(port_id, bytes) {
        Ok(_) => 0,
        Err(_) => -1,
    }
}

/// Close an output port
#[no_mangle]
pub extern "C" fn midi_close_output(port_id: i32) {
    midi::close_output(port_id);
}

/// Create a virtual MIDI input port
/// Returns port ID on success, -1 on error
#[no_mangle]
pub extern "C" fn midi_create_virtual_input(name: *const c_char) -> i32 {
    if name.is_null() {
        return -1;
    }

    let name = unsafe {
        match CStr::from_ptr(name).to_str() {
            Ok(s) => s,
            Err(_) => return -1,
        }
    };

    match midi::create_virtual_input(name) {
        Ok(id) => id,
        Err(_) => -1,
    }
}

/// Connect to an existing MIDI input port by name
/// Returns port ID on success, -1 on error
#[no_mangle]
pub extern "C" fn midi_connect_input(name: *const c_char) -> i32 {
    if name.is_null() {
        return -1;
    }

    let name = unsafe {
        match CStr::from_ptr(name).to_str() {
            Ok(s) => s,
            Err(_) => return -1,
        }
    };

    match midi::connect_input(name) {
        Ok(id) => id,
        Err(_) => -1,
    }
}

/// Start listening for MIDI messages on an input port
/// Returns 0 on success, -1 on error
#[no_mangle]
pub extern "C" fn midi_start_listening(
    port_id: i32,
    callback: MidiInputCallback,
    user_data: *mut c_void,
) -> i32 {
    match midi::start_listening(port_id, callback, user_data) {
        Ok(_) => 0,
        Err(_) => -1,
    }
}

/// Stop listening on an input port
#[no_mangle]
pub extern "C" fn midi_stop_listening(port_id: i32) {
    midi::stop_listening(port_id);
}

/// Close an input port
#[no_mangle]
pub extern "C" fn midi_close_input(port_id: i32) {
    midi::close_input(port_id);
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_clock_lifecycle() {
        let clock = clock_new(120.0);
        assert!(!clock.is_null());

        assert_eq!(clock_set_bpm(clock, 140.0), 0);

        clock_free(clock);
    }
}

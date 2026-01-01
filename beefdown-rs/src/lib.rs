mod timing;
mod clock;

pub use clock::Clock;
use std::ffi::c_void;

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

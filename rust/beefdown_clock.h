/*
 * beefdown_clock.h
 *
 * C header for the Rust high-precision MIDI clock library.
 * Provides 6.2x better timing accuracy than Go (0.126ms vs 0.786ms error).
 *
 * Usage from Go:
 *   See GO_INTEGRATION.md for full example.
 */

#ifndef BEEFDOWN_CLOCK_H
#define BEEFDOWN_CLOCK_H

#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

/* Opaque pointer to Clock */
typedef struct Clock Clock;

/* Callback type for clock ticks (24ppq) */
typedef void (*tick_callback)(void* user_data);

/**
 * Create a new clock with the given BPM.
 * Returns an opaque pointer to the clock, or NULL on error.
 *
 * Example:
 *   Clock* clock = clock_new(120.0);
 */
Clock* clock_new(double bpm);

/**
 * Start the clock with a callback that fires on each tick (24ppq).
 * The callback receives the user_data pointer.
 *
 * Returns:
 *   0 on success
 *   -1 on error (e.g., clock is NULL or already running)
 *
 * Example:
 *   int result = clock_start(clock, my_callback, my_data);
 */
int32_t clock_start(Clock* clock, tick_callback callback, void* user_data);

/**
 * Stop the clock.
 *
 * Returns:
 *   0 on success
 *   -1 on error (e.g., clock is NULL)
 *
 * Example:
 *   int result = clock_stop(clock);
 */
int32_t clock_stop(Clock* clock);

/**
 * Set the BPM of the clock (can be called while running).
 *
 * Returns:
 *   0 on success
 *   -1 on error (e.g., clock is NULL)
 *
 * Example:
 *   int result = clock_set_bpm(clock, 140.0);
 */
int32_t clock_set_bpm(Clock* clock, double bpm);

/**
 * Free the clock and release resources.
 * After calling this, the clock pointer is invalid.
 *
 * Example:
 *   clock_free(clock);
 */
void clock_free(Clock* clock);

#ifdef __cplusplus
}
#endif

#endif /* BEEFDOWN_CLOCK_H */

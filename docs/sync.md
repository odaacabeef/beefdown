# Sync

MIDI sync messages are sent to a dedicated virtual MIDI output called
`beefdown-sync`. Supposedly this improves sync stability [^1].

The `sync` setting supports two options: `none` and `leader`.

When set to `leader`, MIDI clock, start, and stop messages are sent. They can be
used to sync other MIDI devices.

## Ableton Live

This has been tested with Ableton Live for recording MIDI.

So far the outcome hasn't been perfect. Recorded MIDI doesn't quite land where
it's intended. Reducing the sync delay can help, but this is also imprecise.
It's close enough though that it can be fixed by quantizing.

[^1]: https://help.ableton.com/hc/en-us/articles/209071149-Synchronizing-Live-via-MIDI

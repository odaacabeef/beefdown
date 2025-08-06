# Sync

MIDI sync messages are sent to a dedicated virtual MIDI output called
`beefdown-sync`. Supposedly this improves sync stability [^1].

The `sync` setting supports three options: `none`, `leader`, `follower`.

When set to `leader`, MIDI clock, start, and stop messages are sent. They can be
used to sync other MIDI devices.

When set to `follower`, it listens for those messages. Currently, this only
works with other leader instances of `beefdown` because the source name is
hardcoded. This will change soon.

## Ableton Live

So far only `sync;leader` has been tested with Ableton Live for recording MIDI.

The outcome hasn't been perfect. Recorded MIDI doesn't quite land where it's
intended. Reducing the sync delay can help, but this is also imprecise. It's
close enough though that it can be fixed by quantizing.

[^1]: https://help.ableton.com/hc/en-us/articles/209071149-Synchronizing-Live-via-MIDI

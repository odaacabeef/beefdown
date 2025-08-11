# MIDI IO

## Track Output

By default, a virtual MIDI output is created called `beefdown` for sending note
and control messages.

You can configure this output to use an existing device instead:

````
```beef.sequence
output:'Crumar Seven'
```
````

## Sync

The `sync` setting supports three options: `none`, `leader`, `follower`.

### Sync Output

When `sync:leader` is set, MIDI sync messages (clock, start, stop) are sent to a
dedicated virtual output called `beefdown-sync`. Separating sync from track
messages is intended to improve sync stability.

### Sync Input

When `sync:follower` is set, it's expected to find an output called
`beefdown-sync` to listen for sync messages. You can configure the input to be
something else:

````
```beef.sequence
sync:follower
input:'IAC Driver Bus 1'
```
````

### Ableton Live

When synchronizing with a DAW like Live [^1], it's recommended to use the DAW as
the leader.

You can have beefdown lead, but recorded MIDI doesn't quite land where it's
intended. Reducing the sync delay can help, but this is also imprecise. It's
close enough though that it can usually be fixed by quantizing.

[^1]: https://help.ableton.com/hc/en-us/articles/209071149-Synchronizing-Live-via-MIDI

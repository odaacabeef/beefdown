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

When `sync:follower` is set, a virtual input named `beefdown-sync` is created.
This will show up in a DAW to be able to send sync messages to.

You can configure the sequence `input` to instead listen to an existing source
of sync messages. For example, if you wanted to listen to a leader instance of
beefdown:

````
```beef.sequence
sync:follower
input:beefdown-sync
```
````

### Ableton Live

When synchronizing with a DAW like Live [^1], it's recommended to use the DAW as
the leader. You can have beefdown lead, but it's less reliable in the context of
recording. In either case, you'll likely need to adjust the sync delay [^2].

[^1]: https://help.ableton.com/hc/en-us/articles/209071149-Synchronizing-Live-via-MIDI
[^2]: https://www.ableton.com/en/manual/synchronizing-with-link-tempo-follower-and-midi/#sync-delay

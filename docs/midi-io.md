# MIDI IO

## Output

By default, a virtual MIDI output is created called `beefdown` for sending note
and control messages. This is our "track" output.

When `sync:leader` is set, a virual output called `beefdown-sync` is also
created for sending sync messages. This is our "sync" output.

You can configure the track output to use an existing device:

````
```beef.sequence
output:Crumar Seven
```
````

The sync output cannot be configured.

## Input

When `sync:follower` is set, by default beefdown expects to find an output
called `beefdown-sync` to listen for sync messages. You can configure the input
to be something else:

````
```beef.sequence
sync:follower
input:Ableton Live
```
````

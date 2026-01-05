# MIDI IO

## Voice Output

By default, a virtual MIDI output is created called `beefdown` for sending voice
messages.

You can configure this output to use an existing device instead:

````
```beef.sequence
voiceout:'Crumar Seven'
```
````

## Sync

The `sync` setting supports three options: `none`, `leader`, `follower`.

### Sync Output

When `sync:leader` is set, MIDI sync messages (clock, start, stop) are sent to a
dedicated virtual output called `beefdown-sync`. Separating sync from voice
messages is intended to improve sync stability.

You can also send sync messages to an existing midi port instead of creating a
virtual output:

````
```beef.sequence
sync:leader
syncout:mc-dest-b
```
````

### Sync Input

When `sync:follower` is set, a virtual input named `beefdown-sync` is created.
This will show up in a DAW to be able to send sync messages to.

You can configure the sequence `input` to instead listen to an existing source
of sync messages. For example, if you wanted to listen to a leader instance of
beefdown:

````
```beef.sequence
sync:follower
syncin:beefdown-sync
```
````

### Ableton Live

https://help.ableton.com/hc/en-us/articles/209071149-Synchronizing-Live-via-MIDI

## midi-cable

[odaacabeef/midi-cable](https://github.com/odaacabeef/midi-cable) is a project
you could use to offload management of virtual ports.

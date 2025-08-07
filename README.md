# beefdown

A markdown-driven MIDI sequencer.

## Sequences

This file is a sequence. Run it with:

```
go run . README.md
```

![screenshot](/docs/screenshot.png)

Code blocks with `beef` prefixed language identifiers are used to specify
musical information.

See [docs/controls.md](docs/controls.md) for application controls.

### Sequence

Sequence blocks contain global configuration.

````
```beef.sequence
bpm:150
loop:true
sync:none
```
````

See [docs/sync.md](docs/sync.md) for more info on the `sync` setting.

### Parts

Parts are collections of notes.

````
```beef.part name:a
c4:8

a2:6

d3:4



```
````

Each line represents a beat.

`c4:1` would play c4 for 1 beat.

```
c4:1
|| |
|| +--- beats
|+--- octave
+--- note
```

Horizontal space is irrelevant. Do whatever makes the most sense to you
visually.

````
```beef.part name:a'
c4:8      d3:8

     a2:6


               d5:3


```
````

By default parts are sent on channel 1. The next part will be sent on channel 2.

````
```beef.part name:b ch:2
c4:2
          c6:2


     c5:2
               c7:2


```
````

To omit off messages, don't include a beat value. You might want this for
samplers or percussion instruments.

````
```beef.part name:hh-1 group:drums ch:16
gb1
gb1
gb1
gb1
gb1
gb1
gb1
gb1
```
````

You can change the length of each beat with `div`. Recognized values are `4th`,
`4th-triplet`, `8th`, `8th-triplet`, `16th`, and `32nd`. See
[docs/division.md](docs/division.md) for more details on this.

````
```beef.part name:hh-2 group:drums ch:16 div:8th
gb1
gb1
gb1
gb1
    bb1
gb1
gb1
gb1
gb1
gb1
gb1
gb1
    bb1
gb1
gb1
gb1
```
````

For parts with repeated lines, you can multiply them. The next part repeats
`gb1` 24 times.

````
```beef.part name:hh-3 group:drums ch:16 div:8th-triplet
gb1 *24
```
````

You can also use the modulo operator (%) with multiplication. Syntax is `*N%M`
where:

* `N` is the multiplication factor (how many times to repeat)
* `M` is the modulo factor (every Mth step will contain the step)

The next part repeats `gb1` 12 times, then for 12 additional steps every other
step will contain `gb1` (13, 15, 17... `gb1`, 14, 16, 18... blank).

````
```beef.part name:hh-4 group:drums ch:16 div:8th-triplet
gb1 *12
gb1 *12%2
```
````

````
```beef.part name:hh-5 group:drums ch:16 div:16th
gb1     *4
    bb1
gb1     *7
    bb1
gb1     *7
    bb1
gb1     *7
    bb1
gb1     *3
```
````

````
```beef.part name:ks-1 group:drums ch:16 div:8th
      *10
c1
c1
   d1
c1
```
````

````
```beef.part name:ks-2 group:drums ch:16
c1

   d1

c1

   d1

```
````

All notes are sent with a velocity of 100.

Parts also have basic chord support. Run
[examples/chords-ii-v-i.md](examples/chords-ii-v-i.md) to see how that works.

### Arrangements

Arrangements are collections of parts.

````
```beef.arrangement name:kick-snare-hi-hat group:drums
ks-2 hh-1
ks-2 hh-2
ks-2 hh-3
ks-2 hh-4
ks-2 hh-5
```
````

You can also multiply arrangement steps!

````
```beef.arrangement name:all-the-parts group:last
ks-1 hh-1 a
ks-2 hh-1 a'   *2
ks-2 hh-2 a' b *2
ks-2 hh-3 a' b
ks-2 hh-4 a' b
ks-2 hh-5 a' b *2
```
````

## MIDI Output Configuration

By default, a virtual MIDI output called `beefdown` is created to send note and
control messages to. This is usually ideal if you're working with virtual
instruments in a DAW.

You also have the ability to send to a different output instead:

```bash
# List available MIDI outputs
beefdown -list-outputs

# Use a specific MIDI output
beefdown -output 'Crumar Seven' sequence.md
```

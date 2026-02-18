# beefdown

A MIDI sequencer that uses markdown syntax to define sequences.

## Sequences

This file is a sequence. Run it with:

```
beefdown README.md
```

_See [docs/build.md](docs/build.md) for info on building `beefdown`._

![screenshot](docs/screenshot.png)

Code blocks with `beef` prefixed language identifiers are used to specify
musical information.

_See [docs/controls.md](docs/controls.md) for application controls._

### Sequence

Sequence blocks contain global configuration.

````
```beef.sequence
bpm:150
loop:true
sync:none
```
````

_See [docs/midi-io.md](docs/midi-io.md) for more info on the `sync` setting._

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
`4th-triplet`, `8th`, `8th-triplet`, `16th`, and `32nd`. _See
[docs/division.md](docs/division.md) for more details on this._

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
step will contain `gb1` (13, 15, 17... `gb1`; 14, 16, 18... blank).

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

Parts also have basic chord support. _Run `beefdown examples/chords-ii-v-i.md`
to see to see an example._

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

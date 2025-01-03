# seq

A markdown-driven MIDI seqencer.

## Sequences

This file is a sequence. Run it with:

```
go run . README.md
```

Code blocks with `seq` prefixed language identifiers are used to specify musical
information.

Before reading on, switch to code view: [README.md?plain=1](README.md?plain=1).

### Metadata

Metadata blocks contain global configuration.

`seq.metadata`

```seq.metadata
bpm:150
loop:true
# sync:leader (wip)
```

### Parts

Parts are collections of notes.

`seq.part name:<name> group:<group> ch:<channel> div:<division>`

```seq.part name:a
c4:8

a2:6

d3:8



```

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

```seq.part name:a-
c4:8      d3:8

     a2:6


               d5:3


```

The next part will be sent on channel 2.

```seq.part name:b ch:2
c4:2
          c6:2


     c5:2
               c7:2


```

To omit off messages, don't include a beat value. You might want this for
samplers or percussion instruments.

```seq.part name:hh-1 group:drums ch:16
gb1
gb1
gb1
gb1
gb1
gb1
gb1
gb1
```

You can change the length of each beat with `div`. Recognized values are `8th`,
`8th-triplet`, `16th`, and `32nd`.

```seq.part name:hh-2 group:drums ch:16 div:8th
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

```seq.part name:hh-3 group:drums ch:16 div:8th-triplet
gb1
gb1
gb1
gb1
gb1
gb1
gb1
gb1
gb1
gb1
gb1
gb1
gb1
gb1
gb1
gb1
gb1
gb1
gb1
gb1
gb1
gb1
gb1
gb1
```

```seq.part name:hh-4 group:drums ch:16 div:16th
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

```seq.part name:ks-1 group:drums ch:16 div:8th










c1
c1
   d1
c1
```

```seq.part name:ks-2 group:drums ch:16
c1

   d1

c1

   d1

```

### Arrangmements

Arrangements are collections of parts.

`seq.arrangement name:<name> group:<group>`

```seq.arrangement name:kick-snare-hi-hat group:drums
ks-2 hh-1
ks-2 hh-2
ks-2 hh-3
ks-2 hh-4
```

```seq.arrangement name:all-the-parts group:last
ks-1 hh-1 a
ks-2 hh-1 a-
ks-2 hh-1 a-
ks-2 hh-2 a- b
ks-2 hh-2 a- b
ks-2 hh-3 a- b
ks-2 hh-3 a- b
ks-2 hh-4 a- b
ks-2 hh-4 a- b
```

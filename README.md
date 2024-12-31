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
```

### Parts

Parts are collections of notes.

`seq.part name:<name> group:<group> ch:<channel>`

```seq.part name:a ch:1
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

```seq.part name:a- ch:1
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

```seq.part name:hh group:drums ch:16
gb1
gb1
gb1
gb1
gb1
gb1
gb1
gb1
```

```seq.part name:ks-1 group:drums ch:16





c1
   d1

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

```seq.arrangement name:kshh group:drums
ks-2 hh
ks-2 hh
ks-2 hh
ks-2 hh
```

```seq.arrangement name:all group:last
ks-1 hh a
ks-2 hh a-
ks-2 hh a-
ks-2 hh a- b
ks-2 hh a- b
```

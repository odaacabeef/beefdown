# Example Song

This is a simple beefdown sequence to demonstrate the complete pipeline.

## Sequence Metadata

```beef.sequence
.sequence bpm:120 sync:leader output:Beefdown
```

## Bass Part

```beef.part
.part name:bass ch:2 div:24 group:rhythm
c2:4
e2:4
g2:4
c3:4
*2
```

## Melody Part

```beef.part
.part name:melody ch:1 div:24 group:lead
c4:2
d4:2
e4:4
CM7:4
*2
```

## Chord Part

```beef.part
.part name:chords ch:3 div:24 group:harmony
CM7:8
FM7:8
GM7:8
CM7:8
```

## Verse Arrangement

```beef.arrangement
.arrangement name:verse group:main
part:bass
part:melody
part:chords
```

# Euclidean Rhythm Generator

````
```beef.sequence
loop: true
```
````

Generators create parts procedurally.

They accept all the same configuration options as parts along with some
additional options to specify behavior.

The euclidean generator distributes pulses evenly across steps.

## How It Works

The generator uses a Bresenham-based approach (from computer graphics) to
distribute pulses. This is simpler and more predictable than the traditional
Bjorklund algorithm, while producing equivalent musical results.

**The algorithm:**
- Maintains a "bucket" that accumulates pulses
- Each step adds the pulse count to the bucket
- When the bucket overflows (≥ steps), output a pulse and subtract 8
- Otherwise, output a rest

**Example: 5 pulses in 8 steps**

Starting with `bucket = 0`, at each step we add 5 (pulses).
When `bucket >= 8` (steps), output a pulse and subtract 8.
Otherwise, output a rest.

```
Step:    1    2    3    4    5    6    7    8
---------------------------------------------
Add:    +5   +5   +5   +5   +5   +5   +5   +5
Bucket:  5   10    7   12    9    6   11    8
         ↓    ↓    ↓    ↓    ↓    ↓    ↓    ↓
        <8   %8   <8   %8   %8   <8   %8   %8
              2         4    1         3    0

Output:  .    x    .    x    x    .    x    x
```

Result: `.x.xx.xx` - pulses distributed as evenly as possible.

## Classic Euclidean Patterns

### 3 pulses in 8 steps (tresillo rhythm)

````
```beef.gen.euclidean
name: tresillo
group: classic
ch: 1
pulses: 3
steps: 8
note: c4
```
````

### 5 pulses in 8 steps (cinquillo)

````
```beef.gen.euclidean
name: cinquillo
group: classic
ch: 2
pulses: 5
steps: 8
note: e4
```
````

### 5 pulses in 16 steps
````
```beef.gen.euclidean
name: sparse-kick
group: classic
ch: 3
pulses: 5
steps: 16
note: c3
div: 8th
```
````

### 7 pulses in 16 steps
````
```beef.gen.euclidean
name: busy-hats
group: classic
ch: 4
div: 16th
pulses: 7
steps: 16
note: f#5
div: 8th
```
````

````
```beef.arrangement name:euclidean-demo group:classic
tresillo cinquillo sparse-kick busy-hats
```
````

## Rotation

Rotation shifts where the pattern starts without changing the pulse distribution.

**Example: 5 pulses in 16 steps**
```
rotation: 0   x..x..x..x..x...  (starts with pulse on beat 1)
rotation: 2   ..x..x..x..x..x.  (starts with space, pulse on beat 3)
rotation: 4   ....x..x..x..x..  (pulse starts on beat 5)
```

**Musically:** Different rotations create different rhythmic feels and downbeat
emphasis. The same pattern can feel completely different depending on where it
lands in the measure.

### Same pattern, different rotations (5 pulses in 16 steps)

#### Rotation 0 (starts with pulse)
````
```beef.gen.euclidean
name: r0
group: rotation
ch: 1
pulses: 5
steps: 16
note: c4
rotation: 0
div: 8th
```
````

#### Rotation 2 (shifted right by 2)
````
```beef.gen.euclidean
name: r2
group: rotation
ch: 2
pulses: 5
steps: 16
note: e4
rotation: 2
div: 8th
```
````

#### Rotation 4 (shifted right by 4)
````
```beef.gen.euclidean
name: r4
group: rotation
ch: 3
pulses: 5
steps: 16
note: g4
rotation: 4
div: 8th
```
````

Using the same pattern with different rotations can create interesting
polyrhythms.

````
```beef.arrangement name: polyrhythm group: rotation
r0 r2 r4
```
````

## Note Pools

The euclidean generator can use note pools for melodic variation while
maintaining rhythmic structure. When the `notes` parameter contains commas, each
pulse randomly selects from the pool.

### Single Note vs Pool

Single note
````
```beef.gen.euclidean
name: single
group: pools
ch: 1
pulses: 5
steps: 16
note: c4
div: 8th
```
````

Note pool - random selection on each pulse:
````
```beef.gen.euclidean
name: pool
group: pools
ch: 2
pulses: 5
steps: 16
notes: c4,e4,g4,c5
div: 8th
```
````

### Deterministic Randomness

The `seed` parameter ensures patterns are reproducible. Different seeds produce
different melodic sequences with the same rhythm:

````
```beef.gen.euclidean
name: seed-100
group: pools
ch: 3
pulses: 5
steps: 16
notes: c4,e4,g4,c5
seed: 100
div: 8th
```
````

````
```beef.gen.euclidean
name: seed-200
group: pools
ch: 4
pulses: 5
steps: 16
notes: c4,e4,g4,c5
seed: 200
div: 8th
```
````

````
```beef.arrangement name: pool-demo group: pools
single pool seed-100 seed-200
```
````

**Choosing seed values:** don't overthink it - the seed just selects which
random sequence you get, not the quality of randomness. If unset, it defaults to
0. Simple incrementing (1, 2, 3) works fine. Different seeds produce different
patterns, same seed produces the same pattern.

## Combining Pools with Rotation

Rotation and note pools work together - rotation shifts the rhythm timing while
the pool determines note selection:

````
```beef.gen.euclidean
name: pool-rot-0
group: pool-rotation
ch: 5
pulses: 5
steps: 16
notes: c4,e4,g4
seed: 10
rotation: 0
div: 8th
```
````

````
```beef.gen.euclidean
name: pool-rot-3
group: pool-rotation
ch: 6
pulses: 5
steps: 16
notes: c4,e4,g4
seed: 10
rotation: 3
div: 8th
```
````

Same seed means same note sequence, but rotation shifts when those notes play.

````
```beef.arrangement name: pool-rotation-demo group: pool-rotation
pool-rot-0 pool-rot-3
```
````

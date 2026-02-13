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

The generator uses a **Bresenham-based approach** (from computer graphics) to
distribute pulses. This is simpler and more predictable than the traditional
Bjorklund algorithm, while producing equivalent musical results.

**The algorithm:**
- Maintains a "bucket" that accumulates pulses
- Each step adds the pulse count to the bucket
- When the bucket overflows (≥ steps), output a pulse and reset
- Otherwise, output a rest

**Example: 5 pulses in 8 steps**
```
Step:   1  2  3  4  5  6  7  8
Bucket: 5  10 15 20 25 30 35 40
        ↓  ↓  ↓  ↓  ↓  ↓  ↓  ↓
After:  5  2  7  4  1  6  3  0  (mod 8)
Output: x  .  x  .  x  .  x  x
```
Result: `x.xx.xx.` - pulses distributed as evenly as possible.

## Understanding Rotation

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

## Rotation Examples

Rotation changes where the pattern starts, creating different rhythmic feels.

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

Using the same pattern with different rotations creates interesting polyrhythms:

````
```beef.arrangement name: polyrhythm group: rotation
r0 r2 r4
```
````

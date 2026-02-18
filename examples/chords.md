# Chords Example

This example demonstrates chord functionality including triads (major, minor,
diminished, augmented, suspended), 7th chords, extended chords (9ths, 11ths,
13ths), and slash chord notation for inversions and polychords.

## Triads

````
```beef.part name:triads
CM:4
*3
Cdim:4
*3
Caug:4
*3
Csus2:4
*3
```
````

## 7th Chords

````
```beef.part name:sevenths
CM7:4
*3
Cm7:4
*3
C7:4
*3
Cdim7:4
*3
```
````

## Extended Chords (9ths)

````
```beef.part name:ninths
C9:4
*3
Cm9:4
*3
CM9:4
*3
Caug7:4
*3
```
````

## Extended Chords (11ths)

````
```beef.part name:elevenths
C11:4
*3
Cm11:4
*3
CM11:4
*3
Csus4:4
*3
```
````

## Extended Chords (13ths)

````
```beef.part name:thirteenths
C13:4
*3
Cm13:4
*3
CM13:4
*3
Csus2:4
*3
```
````

## Inversions

Slash notation with chord tones in the bass creates inversions.

````
```beef.part name:inversions
CM7:4
*3
CM7/E:4
*3
CM7/G:4
*3
CM7/B:4
*3
```
````

## Polychords

Slash notation with non-chord tones in the bass creates polychords.

````
```beef.part name:polychords
CM7:4
*3
CM7/D:4
*3
CM7/F:4
*3
CM7/A:4
*3
```
````

## Common Jazz Voicings

````
```beef.part name:jazz-voicings
Dm7/G:4
*3
G7/B:4
*3
CM7/E:4
*3
Am7/C:4
*3
```
````

## ii-V-I with Bass Movement

````
```beef.part name:ii-v-i-bass
Dm9:2

G13:2

CM9:4
*3
Dm7:2

Dm7/C:2

G7/B:2

G7/F:2

CM7/E:2

CM7/G:2

CM7:2

```
````

````
```beef.arrangement name:all
triads
sevenths
ninths
elevenths
thirteenths
inversions
polychords
jazz-voicings
ii-v-i-bass
```
````

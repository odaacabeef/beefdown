# Chords Example

This example demonstrates chord functionality including triads (major, minor,
diminished, augmented, suspended), 7th chords (including half-diminished m7b5),
extended chords (9ths, 11ths, 13ths), altered dominants (7b9, 7#9, 7alt, etc.),
and slash chord notation for inversions and polychords.

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

## Altered Dominants

Essential for bebop and modern jazz harmony.

````
```beef.part name:basic-altered
G7:4
*3
G7b9:4
*3
G7#9:4
*3
G7b5:4
*3
```
````

````
```beef.part name:extended-altered
G7#11:4
*3
G7b13:4
*3
G7alt:4
*3
G7:4
*3
```
````

## Minor ii-V-i with Alterations

Classic minor key progression with altered dominants.

````
```beef.part name:minor-ii-v-i
Dm7b5:2

G7b9:2

Cm7:4
*3
Dm7b5:2

G7alt:2

Cm7:4
*3
```
````

## Tritone Substitution

Using altered dominants with tritone substitution (Db7 substitutes for G7).

````
```beef.part name:tritone-sub
Dm7:2

G7b9:2

CM7:4
*3
Dm7:2

Db7#11:2

CM7:4
*3
```
````

## Bebop Line

Chromatic approach with altered dominants.

````
```beef.part name:bebop
Dm7:2

Db7#9:2

CM7:2

Cm7:2

Bm7b5:2

E7b9:2

Am7:2

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
basic-altered
extended-altered
minor-ii-v-i
tritone-sub
bebop
inversions
polychords
jazz-voicings
ii-v-i-bass
```
````

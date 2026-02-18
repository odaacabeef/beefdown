# Slash Chords Example

This example demonstrates slash chord notation for inversions and polychords,
following standard jazz notation conventions.

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
inversions
polychords
jazz-voicings
ii-v-i-bass
```
````

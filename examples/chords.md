# Chords Example

This example demonstrates chord functionality including triads (major, minor,
diminished, augmented, suspended), 7th chords, extended chords (9ths, 11ths,
13ths), and a jazz progression.

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

## Jazz Progression with Extended Chords

````
```beef.part name:jazz-progression
Dm9:2

G13:2

CM9:4
*3
Am11:2

D7:2

```
````

````
```beef.arrangement name:all
triads
sevenths
ninths
elevenths
thirteenths
jazz-progression
```
````

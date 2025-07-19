# Functions

````
```beef.sequence
loop: true
```
````

Functions generate parts.

They accept all the same configuration options as parts along with some
additional options to specify behavior.

## Arpeggiate

````
```beef.func.arpeggiate
name:arp-1
group:arp
notes:c4,e4,g4,c5
length:16
```
````

````
```beef.func.arpeggiate
name:arp-2
group:arp
div:8th
ch:2
notes:c5,g4,c4,e4
length:32
```
````

````
```beef.arrangement name:arps group:arp
arp-1 arp-2
```
````

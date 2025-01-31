# Division

Each quarter note is comprised of 24 clock messages. It can be divided:

| note        | divisor |
| ----------- | ------- |
| 4th         | 24      |
| 4th-triplet | 16      |
| 8th         | 12      |
| 8th-triplet | 8       |
| 16th        | 6       |
| 32nd        | 3       |

Combining parts of different divisions require different step counts if you
intend to match duration. The sum of clock messages needs to be the same for
each part.

```
{steps} * {divisor} = {clock messages}

{clock messages} / {divisor} = {steps}
```

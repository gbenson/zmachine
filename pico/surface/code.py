import board
import time

from digitalio import DigitalInOut, Direction, Pull
from rotaryio import IncrementalEncoder

led = DigitalInOut(board.LED)
led.direction = Direction.OUTPUT


class Encoder:
    def __init__(self, n, *pins):
        self.number = n
        a, b, s = [getattr(board, f"GP{n}") for n in pins]
        self.encoder = IncrementalEncoder(a, b)
        self.switch = DigitalInOut(s)
        self.switch.pull = Pull.UP
        self.lastpos = self.encoder.position
        self.lastsw = self.switch.value

    def update(self):
        anychange = False
        pos = self.encoder.position
        if pos != self.lastpos:
            print(f"encoder[{self.number}].position = {pos}")
            self.lastpos = pos
            anychange = True
        switch = self.switch.value
        if switch != self.lastsw:
            print(f"encoder[{self.number}].switch = {switch}")
            self.lastsw = switch
            anychange = True
        return anychange


def main():
    encoders = [
        Encoder(n, *pins) for n, pins in enumerate((
            (5, 6, 7),    # => 0  Locations:
            (8, 9, 10),   # => A
            (11, 12, 13), # => B    +-----------+
            (14, 15, 28), # => C    | 0         |
            (26, 27, 22), # => D    |  ABCD     |
            (2, 3, 4),    # => X    | X Y Z     |
            (19, 20, 21), # => Y    +-----------+
            (16, 17, 18), # => Z
        ))
    ]

    while True:
        anychange = False
        for e in encoders:
            if e.update():
                anychange = True
        led.value = anychange
        time.sleep(0.001)  # XXX maybe?


if __name__ == "__main__":
    main()

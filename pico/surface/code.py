import board
import time

from adafruit_midi import MIDI
from adafruit_midi.control_change import ControlChange
from digitalio import DigitalInOut, Direction, Pull
from rotaryio import IncrementalEncoder
from usb_midi import PortIn, PortOut, ports

led = DigitalInOut(board.LED)
led.direction = Direction.OUTPUT

midi_ins = [
    MIDI(midi_in=port)
    for port in ports
    if isinstance(port, PortIn)
]
midi_outs = [
    MIDI(midi_out=port, out_channel=15)
    for port in ports
    if isinstance(port, PortOut)
]

def send_midi(msg):
    for port in midi_outs:
        port.send(msg)


class Encoder:
    def __init__(self, n, *pins):
        if n < 0 or n >= 16:
            # cc number calculations assume 0 <= n < 16
            raise NotImplementedError(n)
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
        if (delta := pos - self.lastpos):
            print(f"encoder[{self.number}].position = {pos}")
            control = self.number + 16  # 16..31 inclusive
            value = max(0, min(127, delta+64))
            send_midi(ControlChange(control, value))
            self.lastpos = pos
            anychange = True

        switch = self.switch.value
        if switch != self.lastsw:
            print(f"encoder[{self.number}].switch = {switch}")
            control = self.number + 104  # 104..119 inclusive
            value = 127 if switch else 0
            send_midi(ControlChange(control, value))
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

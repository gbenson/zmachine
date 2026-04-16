import board
import time

from adafruit_midi import MIDI
from adafruit_midi.control_change import ControlChange
from adafruit_midi.control_change_values import VOLUME
from adafruit_midi.midi_reset import Reset
from analogio import AnalogIn
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
            value = 0 if switch else 127
            send_midi(ControlChange(control, value))
            self.lastsw = switch
            anychange = True

        return anychange


class Potentiometer:
    def __init__(self, pin, cc_msb, cc_lsb=-1, *, squelch=64):
        if cc_lsb < 0:
            cc_lsb = cc_msb + 32
        self.cc_msb = cc_msb
        self.cc_lsb = cc_lsb
        self.pot = AnalogIn(pin)
        self.lastval = self.pot.value
        self.squelch = squelch

    def update(self):
        # Pico has 12-bit ADC, so 4096 possible values, but the
        # library scales values up to uint16 (so 0..65535), then
        # we then scale them down to the 14 bits we can fit in a
        # regular (paired) MIDI CC.
        val = self.pot.value >> 2
        if abs(val - self.lastval) < self.squelch:
            return False
        print(f"potentiometer[{self.cc_msb}].value = {val}")
        self.send_value(val)
        self.lastval = val
        return True

    def send_value(self, value):
        send_midi(ControlChange(self.cc_msb, (value >> 7) & 127))
        if (cc_lsb := self.cc_lsb) is None:
            return
        send_midi(ControlChange(cc_lsb, value & 127))


def main():
    pots = [Potentiometer(board.A2, VOLUME)]  # ADC2 == GP28 == physical pin 34
    encoders = [
        Encoder(n, *pins) for n, pins in enumerate((
            (8, 9, 10),   # => A  Locations:
            (11, 12, 13), # => B
            (14, 15, 0),  # => C    +-----------+
            (26, 27, 22), # => D    | 0         |
            (5, 6, 7),    # => 0    |  ABCD     |
            (2, 3, 4),    # => X    | X Y Z     |
            (19, 20, 21), # => Y    +-----------+
            (16, 17, 18), # => Z
        ))
    ]

    blink_ticks = 0
    while True:
        for port in midi_ins:
            while (msg := port.receive()):
                print("reset!")
                if isinstance(msg, Reset):
                    for p in pots:
                        p.send_value(p.lastval)

        anychange = False
        for p in pots:
            if p.update():
                anychange = True
        for e in encoders:
            if e.update():
                anychange = True

        if anychange:
            blink_ticks = 20
            led.value = False
        elif blink_ticks > 0:
            blink_ticks -= 1
        else:
            led.value = True

        time.sleep(0.001)


if __name__ == "__main__":
    main()

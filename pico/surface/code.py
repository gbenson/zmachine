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

        switch = not self.switch.value
        if switch != self.lastsw:
            print(f"encoder[{self.number}].switch = {switch}")
            control = self.number + 104  # 104..119 inclusive
            value = 127 if switch else 0
            send_midi(ControlChange(control, value))
            self.lastsw = switch
            anychange = True

        return anychange


class Potentiometer:
    def __init__(self, pin, cc_msb, cc_lsb=-1, *, smooth=16):
        if cc_lsb < 0:
            cc_lsb = cc_msb + 32
        self.cc_msb = cc_msb
        self.cc_lsb = cc_lsb
        self.pot = AnalogIn(pin)

        # smooth must be a power of two
        self.smoothmask = (smooth - 1)
        self.smoothshift = self.smoothmask.bit_length()
        if (1 << self.smoothshift) != smooth:
            raise ValueError("must be a power of 2")

        val = self.pot.value
        self.buf = [val] * smooth
        self.lastsum = sum(self.buf)
        self.bufpos = 0
        self.lastval = val

    def update(self):
        val = self.pot.value

        buf = self.buf
        pos = self.bufpos
        lastsum = self.lastsum
        nextsum = lastsum - buf[pos] + val
        #print(f"\x1B[34m{val:4} {pos:3} {lastsum:x} {nextsum:x}\x1B[0m")
        buf[pos] = val
        self.bufpos = (pos + 1) & self.smoothmask

        if lastsum == nextsum:
            return False
        self.lastsum = nextsum

        # AnalogIn.value returns uint16 (so 0..65535), but the Pico
        # has a 12-bit ADC.  We're shifting to divide by len(self.buf)
        # anyway, so we add 4 to that shift to remove the extra 4 bits
        # at the same time.
        val = nextsum >> (self.smoothshift + 4)

        if abs(val - self.lastval) < 2:
            return False
        self.lastval = val

        # Now upscale the 12-bit value into the 14 bits that fit in a
        # regular (paired) MIDI CC.
        val <<= 2

        print(f"potentiometer[{self.cc_msb}].value = {val}")
        self.send_value(val)
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
            (0, 1, 16),   # => A  Locations:
            (2, 3, 17),   # => B
            (4, 5, 18),   # => C    +-----------+
            (6, 7, 19),   # => D    | 0         |
            (8, 9, 20),   # => 0    |  ABCD     |
            (10, 11, 21), # => X    | X Y Z     |
            (12, 13, 22), # => Y    +-----------+
            (14, 15, 26), # => Z
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

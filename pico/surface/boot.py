import board
import storage
import supervisor
import usb_hid
import usb_midi

from digitalio import DigitalInOut, Pull

# Define our USB Device Descriptor vendor and product ID.
supervisor.set_usb_identification(
    manufacturer="Zmachine",    # default: "Raspberry Pi"
    product="Control Surface",  # default: "Pico"
)

# Don't present as a USB drive unless the encoder whose
# switch is on GP18 is pressed (i.e. GP18 is shorted to
# GND).  ALSA doesn't see the I2S DAC for some reason if
# the CIRCUITPY drive is present when the Pi boots (and
# it's then *minutes* before anything shows on the OLED
# or SSH comes up...)
switch = DigitalInOut(board.GP18)
switch.pull = Pull.UP
if switch.value:
    storage.disable_usb_drive()

# Don't present as a USB keyboard and mouse to the host.
# We show up as "Zmachine Control Surface Keyboard" and
# "Zmachine Control Surface Mouse" otherwise.
usb_hid.disable()

# Define our USB MIDI interface names.
usb_midi.set_names(
    in_jack_name="ZMCS",   # default: "CircuitPython usb_midi.ports[0]"
    out_jack_name="ZMCS",  # default: "CircuitPython usb_midi.ports[0]"
)

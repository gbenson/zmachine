import supervisor
import usb_hid
import usb_midi

# Define our USB Device Descriptor vendor and product ID.
supervisor.set_usb_identification(
    manufacturer="Zmachine",    # default: "Raspberry Pi"
    product="Control Surface",  # default: "Pico"
)

# Don't present as a USB keyboard and mouse to the host.
# We show up as "Zmachine Control Surface Keyboard" and
# "Zmachine Control Surface Mouse" otherwise.
usb_hid.disable()

# Define our USB MIDI interface names.
usb_midi.set_names(
    in_jack_name="ZMCS",   # default: "CircuitPython usb_midi.ports[0]"
    out_jack_name="ZMCS",  # default: "CircuitPython usb_midi.ports[0]"
)

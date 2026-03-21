package ssd1305emu

import (
	"periph.io/x/conn/v3"
	"periph.io/x/conn/v3/spi"
)

type spiconn struct {
	port *Port
}

// Duplex implements [conn.Conn].
func (c *spiconn) Duplex() conn.Duplex {
	panic("ssd1305emu.spiconn.Duplex")
}

// String implements [fmt.Stringer].
func (c *spiconn) String() string {
	return "ssd1305emu:SPI"
}

// Tx implements [conn.Conn].
func (c *spiconn) Tx(w, r []byte) error {
	if r != nil {
		panic("ssd1305emu.spiconn.Tx")
	}

	return c.port.recv(w)
}

// TxPackets implements [spi.Conn].
func (c *spiconn) TxPackets(p []spi.Packet) error {
	panic("ssd1305emu.spiconn.TxPackets")
}

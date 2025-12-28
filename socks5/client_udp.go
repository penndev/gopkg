package socks5

import (
	"errors"
	"fmt"
	"net"
)

// UDP Read Write func
type UDPClient struct {
	net.Conn
	ATYP     ATYP
	DST_ADDR []byte
	DST_PORT uint16
}

func (c *UDPClient) Read(b []byte) (int, error) {
	buf := make([]byte, 1024)
	n, err := c.Conn.Read(buf)
	datagram := UDPDatagram{}
	if err != nil {
		return 0, err
	}
	err = datagram.Decode(buf[:n])
	if err != nil {
		return 0, err
	}
	bufLen := len(datagram.DATA)
	if len(b) < bufLen {
		return 0, fmt.Errorf("UDPClient Read buf si small[%d]", bufLen)
	}
	copy(b[:bufLen], datagram.DATA[:])
	return bufLen, err
}

func (c *UDPClient) Write(data []byte) (int, error) {
	datagram := UDPDatagram{
		ATYP:     c.ATYP,
		DST_ADDR: c.DST_ADDR,
		DST_PORT: c.DST_PORT,
		DATA:     data,
	}
	if d, err := datagram.Encode(); err == nil {
		n, err := c.Conn.Write(d)
		if err != nil {
			return 0, err
		}
		if n != len(d) {
			return 0, errors.New("write byte len error")
		}
		return len(data), nil
	} else {
		return 0, err
	}
}

func (c *UDPClient) Close() error {
	return c.Conn.Close()
}

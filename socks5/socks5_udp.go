package socks5

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
)

type UDPDatagram struct {
	ATYP     ATYP
	DST_ADDR []byte
	DST_PORT uint16
	DATA     []byte
}

func (r *UDPDatagram) Decode(buf []byte) error {
	if len(buf) < 7 {
		return errors.New("UDPDatagram Decode error buf size")
	}
	r.ATYP = ATYP(buf[3])
	portLen := 0
	switch r.ATYP {
	case ATYP_IPV4:
		if len(buf) < 10 {
			return errors.New("UDPDatagram Decode 0x01 err")
		}
		r.DST_ADDR = buf[4:8]
		portLen = 8
	case ATYP_DOMAIN:
		addrLen := int(buf[4])
		if len(buf) < (addrLen + 7) {
			return errors.New("UDPDatagram Decode 0x03 err")
		}
		r.DST_ADDR = buf[5 : 5+addrLen]
		portLen = 5 + addrLen
	case ATYP_IPV6:
		if len(buf) < 22 {
			return errors.New("UDPDatagram Decode 0x04 err")
		}
		r.DST_ADDR = buf[4:21]
		portLen = 21
	default:
		log.Println(buf)
		return fmt.Errorf("UDPDatagram Decode atyp error %d", r.ATYP)
	}
	r.DST_PORT = binary.BigEndian.Uint16([]byte{buf[portLen], buf[portLen+1]})
	r.DATA = buf[portLen+2:]
	return nil
}

func (r *UDPDatagram) Encode() ([]byte, error) {
	buf := []byte{0x00, 0x00, 0x00, byte(r.ATYP)}
	if r.ATYP == ATYP_DOMAIN {
		buf = append(buf, byte(len(r.DST_ADDR)))
	}
	buf = append(buf, r.DST_ADDR...)
	bufPort := make([]byte, 2)
	binary.BigEndian.PutUint16(bufPort, r.DST_PORT)
	buf = append(buf, bufPort...)
	buf = append(buf, r.DATA...)
	return buf, nil
}

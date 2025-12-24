package socks5

import (
	"encoding/binary"
	"errors"
)

type Requests struct {
	CMD      CMD
	ATYP     ATYP
	DST_ADDR []byte
	DST_PORT uint16
}

func (r *Requests) Encode() ([]byte, error) {
	buf := []byte{Version, byte(r.CMD), 0x00, byte(r.ATYP)}
	switch r.ATYP {
	case ATYP_IPV4:
		if len(r.DST_ADDR) != 4 {
			return nil, errors.New("ATYP_IPV4 len not 4")
		}
		buf = append(buf, r.DST_ADDR...)
	case ATYP_DOMAIN:
		if len(r.DST_ADDR) < 1 && len(r.DST_ADDR) > 255 {
			return nil, errors.New("ATYP_DOMAIN len error")
		}
		domainBuf := append([]byte{byte(len(r.DST_ADDR))}, r.DST_ADDR...)
		buf = append(buf, domainBuf...)
	case ATYP_IPV6:
		if len(r.DST_ADDR) != 16 {
			return nil, errors.New("ATYP_IPV6 len not 4")
		}
		buf = append(buf, r.DST_ADDR...)
	default:
		return nil, errors.New("ATYP error")
	}
	bufPort := make([]byte, 2)
	binary.BigEndian.PutUint16(bufPort, r.DST_PORT)
	buf = append(buf, bufPort...)
	return buf, nil
}

func (r *Requests) Decode(buf []byte) error {
	if buf[0] != Version {
		return errors.New("socks5 version error")
	}
	cmd := CMD(buf[1])
	switch cmd {
	case CMD_BIND, CMD_CONNECT, CMD_UDP_ASSOCIATE:
		r.CMD = cmd
	default:
		return errors.New("cmd error")
	}
	if buf[2] != 0 {
		return errors.New("RSV error")
	}
	switch ATYP(buf[3]) {
	case ATYP_IPV4:
		if len(buf) != 10 {
			return errors.New("Replies Decode ipv4 len error")
		}
		r.ATYP = ATYP_IPV4
		r.DST_ADDR = buf[4:8]
		r.DST_PORT = binary.BigEndian.Uint16(buf[8:10])
	case ATYP_DOMAIN:
		r.ATYP = ATYP_DOMAIN
		domainLen := int(buf[4])
		if len(buf) != (domainLen + 7) {
			return errors.New("Replies Decode domain len err")
		}
		r.DST_ADDR = buf[5 : 5+domainLen]
		r.DST_PORT = binary.BigEndian.Uint16(buf[5+domainLen : 7+domainLen])
	case ATYP_IPV6:
		r.ATYP = ATYP_IPV6
		if len(buf) != 22 {
			return errors.New("Replies Decode ipv6 len err")
		}
		r.DST_ADDR = buf[4:20]
		r.DST_PORT = binary.BigEndian.Uint16(buf[20:22])
	default:
		return errors.New("atyp error")
	}
	return nil
}

package socks5

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strconv"
)

// o  X'00' succeeded
// o  X'01' general SOCKS server failure
// o  X'02' connection not allowed by ruleset
// o  X'03' Network unreachable
// o  X'04' Host unreachable
// o  X'05' Connection refused
// o  X'06' TTL expired
// o  X'07' Command not supported
// o  X'08' Address type not supported
// o  X'09' to X'FF' unassigned
type REP byte

const (
	REP_SUCCEEDED              REP = 0x00
	REP_GENERAL_FAILURE        REP = 0x01
	REP_CONNECTION_NOT_ALLOWED REP = 0x02
	REP_NETWORK_UNREACHABLE    REP = 0x03
	REP_HOST_UNREACHABLE       REP = 0x04
	REP_CONNECTION_REFUSED     REP = 0x05
	REP_TTL_EXPIRED            REP = 0x06
	REP_COMMAND_NOT_SUPPORTED  REP = 0x07
	REP_ADDRESS_NOT_SUPPORTED  REP = 0x08
)

type Replies struct {
	REP      REP
	ATYP     ATYP
	BND_ADDR []byte
	BND_PORT uint16
}

// 从字节进行结构体序列化
func (r *Replies) Decode(buf []byte) error {
	if len(buf) < 7 {
		return errors.New("Replies Decode error byte")
	}
	if buf[0] != Version {
		return fmt.Errorf("error version %d", buf[0])
	}
	r.REP = REP(buf[1])

	switch ATYP(buf[3]) {
	case ATYP_IPV4:
		if len(buf) != 10 {
			return errors.New("Replies Decode ipv4 len error")
		}
		r.ATYP = ATYP_IPV4
		r.BND_ADDR = buf[4:8]
		r.BND_PORT = binary.BigEndian.Uint16(buf[8:10])
	case ATYP_DOMAIN:
		r.ATYP = ATYP_DOMAIN
		domainLen := int(buf[4])
		if len(buf) != (domainLen + 7) {
			return errors.New("Replies Decode domain len err")
		}
		r.BND_ADDR = buf[5 : 5+domainLen]
		r.BND_PORT = binary.BigEndian.Uint16(buf[5+domainLen : 7+domainLen])
	case ATYP_IPV6:
		r.ATYP = ATYP_IPV6
		if len(buf) != 22 {
			return errors.New("Replies Decode ipv6 len err")
		}
		r.BND_ADDR = buf[4:20]
		r.BND_PORT = binary.BigEndian.Uint16(buf[20:22])
	default:
		return errors.New("atyp error")
	}
	return nil
}

func (r *Replies) Encode() ([]byte, error) {
	buf := []byte{Version, byte(r.REP), 0x00, byte(r.ATYP)}
	switch r.ATYP {
	case ATYP_IPV4:
		if len(r.BND_ADDR) != 4 {
			return nil, errors.New("ATYP_IPV4 len not 4")
		}
		buf = append(buf, r.BND_ADDR...)
	case ATYP_DOMAIN:
		if len(r.BND_ADDR) < 1 && len(r.BND_ADDR) > 255 {
			return nil, errors.New("ATYP_DOMAIN len error")
		}
		domainBuf := append([]byte{byte(len(r.BND_ADDR))}, r.BND_ADDR...)
		buf = append(buf, domainBuf...)
	case ATYP_IPV6:
		if len(r.BND_ADDR) != 16 {
			return nil, errors.New("ATYP_IPV6 len not 4")
		}
		buf = append(buf, r.BND_ADDR...)
	default:
		return nil, errors.New("ATYP error")
	}
	bufPort := make([]byte, 2)
	binary.BigEndian.PutUint16(bufPort, r.BND_PORT)
	buf = append(buf, bufPort...)
	return buf, nil
}

func (r *Replies) Addr() string {
	var host string
	switch r.ATYP {
	case ATYP_IPV4, ATYP_IPV6:
		host = net.IP(r.BND_ADDR).String()
	case ATYP_DOMAIN:
		host = string(r.BND_ADDR)
	}
	return net.JoinHostPort(host, strconv.Itoa(int(r.BND_PORT)))
}

// https://datatracker.ietf.org/doc/html/rfc1928
// https://datatracker.ietf.org/doc/html/rfc1929
package socks5

const Version = 0x05

// The values currently defined for METHOD are:
//
//	o  X'00' NO AUTHENTICATION REQUIRED
//	o  X'01' GSSAPI
//	o  X'02' USERNAME/PASSWORD
//	o  X'03' to X'7F' IANA ASSIGNED
//	o  X'80' to X'FE' RESERVED FOR PRIVATE METHODS
//	o  X'FF' NO ACCEPTABLE METHODS
type METHOD byte

const (
	METHOD_NO_AUTH           METHOD = 0x00
	METHOD_USERNAME_PASSWORD METHOD = 0x02
	METHOD_NO_ACCEPTABLE     METHOD = 0xFF
)

// o  CMD
//
//	o  CONNECT X'01'
//	o  BIND X'02'
//	o  UDP ASSOCIATE X'03'
type CMD byte

const (
	CMD_CONNECT       CMD = 0x01
	CMD_BIND          CMD = 0x02
	CMD_UDP_ASSOCIATE CMD = 0x03
)

// o  ATYP   address type of following address
//
//	o  IP V4 address: X'01'
//	o  DOMAINNAME: X'03'
//	o  IP V6 address: X'04'
type ATYP byte

const (
	ATYP_IPV4   ATYP = 0x01
	ATYP_DOMAIN ATYP = 0x03
	ATYP_IPV6   ATYP = 0x04
)

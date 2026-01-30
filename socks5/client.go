package socks5

import (
	"errors"
	"fmt"
	"net"
	"strconv"
)

// Socks5 Clint Conn
type Client struct {
	Username string
	Password string

	// tcp客户端
	Conn net.Conn

	// udp客户端
	rwUDP *UDPClient
}

// Parse SOCKS5 Requests struct
// https://datatracker.ietf.org/doc/html/rfc1928#section-4
func (c *Client) requests(network, address string) (Requests, error) {
	req := Requests{}

	switch network {
	case "tcp":
		req.CMD = CMD_CONNECT
	case "udp":
		req.CMD = CMD_UDP_ASSOCIATE
	default:
		return req, errors.New("not support " + network)
	}
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return req, err
	}
	ip := net.ParseIP(host)
	if ip != nil {
		if ip.To4() != nil {
			req.ATYP = ATYP_IPV4
			req.DST_ADDR = []byte(ip.To4())
		} else {
			req.ATYP = ATYP_IPV6
			req.DST_ADDR = []byte(ip.To16())
		}
	} else {
		req.ATYP = ATYP_DOMAIN
		req.DST_ADDR = []byte(host)
	}

	portInt, err := strconv.Atoi(port)
	if err != nil {
		return req, err
	}
	req.DST_PORT = uint16(portInt)
	return req, nil
}

func (c *Client) Dial(network, address string) (net.Conn, error) {
	req, err := c.requests(network, address)
	if err != nil {
		return nil, err
	}
	if b, err := req.Encode(); err == nil {
		c.Conn.Write(b)
	} else {
		return nil, err
	}
	buf := make([]byte, 231)
	n, err := c.Conn.Read(buf)
	if err != nil {
		return nil, err
	}
	rep := Replies{}
	rep.Decode(buf[:n])
	if rep.REP == 0x00 {
		if req.CMD == CMD_UDP_ASSOCIATE {
			UDPrw, err := net.Dial("udp", rep.Addr())
			c.rwUDP = &UDPClient{
				Conn:     UDPrw,
				ATYP:     req.ATYP,
				DST_ADDR: req.DST_ADDR,
				DST_PORT: req.DST_PORT,
			}
			return c.rwUDP, err
		} else {
			// how about the bind?
			return c.Conn, nil
		}
	} else {
		return nil, fmt.Errorf("error replies REP [%d]", rep.REP)
	}
}

func (c *Client) Negotiation() error {
	var err error
	if c.Username == "" {
		_, err = c.Conn.Write([]byte{Version, 0x1, byte(METHOD_NO_AUTH)})
	} else {
		_, err = c.Conn.Write([]byte{Version, 0x2, byte(METHOD_NO_AUTH), byte(METHOD_USERNAME_PASSWORD)})
	}
	if err != nil {
		return err
	}
	buf := make([]byte, 2)
	rn, err := c.Conn.Read(buf)
	if err != nil {
		return err
	}
	if rn != 2 || buf[0] != Version {
		return errors.New("error socks5 service Version")
	}
	switch METHOD(buf[1]) {
	case METHOD_NO_AUTH:
		return nil
	case METHOD_USERNAME_PASSWORD:
		buf := []byte{0x01, byte(len(c.Username))}
		buf = append(buf, []byte(c.Username)...)
		buf = append(buf, byte(len(c.Password)))
		buf = append(buf, []byte(c.Password)...)
		if _, err := c.Conn.Write(buf); err != nil {
			return err
		}
		resBuf := make([]byte, 2)
		rn, err := c.Conn.Read(resBuf)
		if err != nil {
			return err
		}
		if rn != 2 || resBuf[0] != 0x01 {
			return errors.New("error socks5 username/password Version")
		}
		if resBuf[1] != 0x00 {
			return errors.New("error socks5 username/password")
		}
	default:
		return errors.New("error socks method not allow")
	}
	return nil
}

func (c *Client) Close() error {
	if c.rwUDP != nil {
		c.rwUDP.Close()
	}
	return c.Conn.Close()
}

func NewClient(address, user, pass string) (*Client, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	c := &Client{
		Username: user,
		Password: pass,
		Conn:     conn,
	}
	err = c.Negotiation()
	if err != nil {
		conn.Close()
		return nil, err
	}
	return c, nil
}

package ipstack

import (
	"errors"
	"net"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv6"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/icmp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/udp"
	"gvisor.dev/gvisor/pkg/waiter"
)

type ForwarderTCPRequest struct {
	*tcp.ForwarderRequest
}

func (r *ForwarderTCPRequest) TCPConn() (net.Conn, error) {
	var (
		waiterQueue waiter.Queue
	)
	endPoint, err := r.CreateEndpoint(&waiterQueue)
	if err != nil {
		r.Complete(true)
		return nil, errors.New(err.String())
	}
	defer r.Complete(false)
	localConn := gonet.NewTCPConn(&waiterQueue, endPoint)
	return localConn, nil
}

type ForwarderUDPRequest struct {
	*udp.ForwarderRequest
}

func (r *ForwarderUDPRequest) UDPConn() (net.Conn, error) {
	var (
		waiterQueue waiter.Queue
	)
	endPoint, err := r.CreateEndpoint(&waiterQueue)
	if err != nil {
		return nil, errors.New(err.String())
	}
	localConn := gonet.NewUDPConn(&waiterQueue, endPoint)
	return localConn, nil
}

type Option struct {
	Mtu        int
	Device     string
	HandleTCP  func(*ForwarderTCPRequest)
	handlerUDP func(*ForwarderUDPRequest)
	EndPoint   stack.LinkEndpoint
}

func Start(option Option) {
	s := stack.New(stack.Options{
		NetworkProtocols: []stack.NetworkProtocolFactory{
			ipv4.NewProtocol,
			ipv6.NewProtocol,
		},
		TransportProtocols: []stack.TransportProtocolFactory{
			tcp.NewProtocol,
			udp.NewProtocol,
			icmp.NewProtocol4,
			icmp.NewProtocol6,
		},
	})

	tcpForwarder := tcp.NewForwarder(s, 0, 2048, func(fr *tcp.ForwarderRequest) { option.HandleTCP(&ForwarderTCPRequest{fr}) })
	s.SetTransportProtocolHandler(tcp.ProtocolNumber, tcpForwarder.HandlePacket)

	udpForwarder := udp.NewForwarder(s, func(fr *udp.ForwarderRequest) { option.handlerUDP(&ForwarderUDPRequest{fr}) })
	s.SetTransportProtocolHandler(udp.ProtocolNumber, udpForwarder.HandlePacket)

	nicID := tcpip.NICID(s.UniqueID())
	s.CreateNICWithOptions(nicID, option.EndPoint, stack.NICOptions{
		Disabled: false,
	})
	s.SetPromiscuousMode(nicID, true)
	s.SetSpoofing(nicID, true)
	s.SetRouteTable([]tcpip.Route{
		{
			Destination: header.IPv4EmptySubnet,
			NIC:         nicID,
		},
		{
			Destination: header.IPv6EmptySubnet,
			NIC:         nicID,
		},
	})
}

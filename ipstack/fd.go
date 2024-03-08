//go:build linux
// +build linux

package ipstack

import (
	"gvisor.dev/gvisor/pkg/tcpip/link/fdbased"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

func NewEndpoint(fd int) (stack.LinkEndpoint, error) {
	ep, err := fdbased.New(&fdbased.Options{
		FDs: []int{fd},
		// MTU: mtu,
		// TUN only, ignore ethernet header.
		EthernetHeader: false,
	})
	if err != nil {
		return nil, err
	}
	return ep, nil
}

package netconn

import (
	"context"
	"net"
	"strconv"
)

var _ net.Addr = Addr("")

type Addr string

func (a Addr) Network() string {
	return ""
}

func (a Addr) String() string {
	return string(a)
}

func (a Addr) Split() (Host, uint16) {
	host, port, err := net.SplitHostPort(string(a))
	if err != nil {
		return Host(a), 0
	}
	p, _ := strconv.Atoi(port)
	return Host(host), uint16(p)
}

func (a Addr) Port() uint16 {
	_, port := a.Split()
	return port
}

func (a Addr) Host() Host {
	host, _ := a.Split()
	return host
}

type Host []byte

func (h Host) String() string {
	i := h.asIP()
	if i == nil {
		return string(h)
	}
	return i.String()
}

func (h Host) AsIP() net.IP {
	i := h.asIP()
	if i != nil {
		return i
	}

	addrs, _ := net.DefaultResolver.LookupIPAddr(context.Background(), string(h))
	if len(addrs) == 0 {
		return nil
	}

	return addrs[len(addrs)-1].IP.To4()
}

func (h Host) asIP() net.IP {
	if string(h) == "::" {
		return net.IP{127, 0, 0, 1}
	}
	i := net.IP(h)
	v4 := i.To4()
	if v4 != nil {
		return v4
	}
	return i.To16()
}

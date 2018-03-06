package main

import (
	"fmt"
	"net"
)

// Transfer ...
type Transfer interface {
	Open() (net.Conn, error)
	Close() error
	GetPort() int
	GetIP() net.IP
}

// PassiveTransfer ...
type PassiveTransfer struct {
	tcpListener *net.TCPListener
	port        int
	ip          net.IP
	conn        net.Conn
}

// CreatePassiveTransfer ...
func CreatePassiveTransfer() (*PassiveTransfer, error) {
	transfer := &(PassiveTransfer{tcpListener: nil, port: 0, ip: net.ParseIP("0.0.0.0")})
	var err error
	for port := minPort; port <= maxPort; port++ {
		laddr, err := net.ResolveTCPAddr("tcp", ":"+fmt.Sprintf("%v", port))
		if err != nil {
			continue
		}
		transfer.tcpListener, err = net.ListenTCP("tcp", laddr)
		if err == nil {
			break
		}
	}
	// addr, _ := net.ResolveTCPAddr("tcp", ":0")
	// var err error
	// transfer.tcpListener, err = net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, err
	}
	transfer.port = transfer.tcpListener.Addr().(*net.TCPAddr).Port
	transfer.ip = transfer.tcpListener.Addr().(*net.TCPAddr).IP
	return transfer, nil
}

// Open ...
func (p *PassiveTransfer) Open() (net.Conn, error) {
	if p.conn == nil {
		var err error
		p.conn, err = p.tcpListener.Accept()
		if err != nil {
			return nil, err
		}
	}
	return p.conn, nil
}

// Close ...
func (p *PassiveTransfer) Close() error {
	if p.tcpListener != nil {
		p.tcpListener.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
	return nil
}

// GetPort ...
func (p *PassiveTransfer) GetPort() int {
	return p.port
}

// GetIP ...
func (p *PassiveTransfer) GetIP() net.IP {
	return p.ip
}

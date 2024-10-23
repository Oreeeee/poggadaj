package structs

import "net"

type ClientInfo struct {
	Conn          net.Conn
	UIN           uint32
	Status        uint32
	Authenticated bool
}

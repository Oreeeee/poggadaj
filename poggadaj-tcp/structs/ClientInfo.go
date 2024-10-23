package structs

import (
	"net"
	uv "poggadaj-tcp/universal"
)

type ClientInfo struct {
	Conn          net.Conn
	UIN           uint32
	Status        uint32
	Authenticated bool
	NotifyList    []uv.GG_NotifyContact
}

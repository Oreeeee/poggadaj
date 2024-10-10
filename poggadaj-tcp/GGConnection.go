package main

import (
	"net"
	"poggadaj-tcp/universal"
)

type GGConnection struct {
	// Backend/frontend info
	UIN        uint32
	Status     uint32
	NotifyList []universal.GG_NotifyContact

	// Backend-only info
	GGVer         int
	Authenticated bool
	Conn          net.Conn
}

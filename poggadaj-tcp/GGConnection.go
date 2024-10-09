package main

import "net"

type GGConnection struct {
	// Backend/frontend info
	UIN    uint32
	Status uint32

	// Backend-only info
	GGVer         int
	Authenticated bool
	Conn          net.Conn
}

package main

import (
	"fmt"
	"poggadaj-tcp/gg60"
	"poggadaj-tcp/universal"
)

func Handle_GG60(currConn GGConnection, pRecv universal.GG_Packet) {
	// Handle the initial log in
	p := gg60.GG_Login60{}
	p.Deserialize(pRecv.Data)
	fmt.Println("Decoded data: ", p)

	currConn.UIN = p.UIN

	fmt.Println("Sending login response")
	if p.Hash == 4105424095 { // Password: 123
		currConn.Authenticated = true
		fmt.Println("Sending GG_LOGIN_OK")
		pOut := universal.InitGG_Packet(universal.GG_LOGIN_OK, []byte{})
		_, err := pOut.Send(currConn.Conn)
		if err != nil {
			fmt.Println("Error: ", err)
		}
	} else {
		fmt.Println("Sending GG_LOGIN_FAILED")
		pOut := universal.InitGG_Packet(universal.GG_LOGIN_FAILED, []byte{})
		_, err := pOut.Send(currConn.Conn)
		if err != nil {
			fmt.Println("Error: ", err)
		}
		return
	}

	// Connection loop
	for {
		pRecv := universal.GG_Packet{}
		err := pRecv.Receive(currConn.Conn)
		if err != nil {
			fmt.Println("Error receiving data, dropping connection!: ", err)
			return
		}

		switch pRecv.PacketType {
		case universal.GG_LIST_EMPTY:
			fmt.Println("Received GG_LIST_EMPTY")
		case universal.GG_NEW_STATUS:
			fmt.Println("Received GG_NEW_STATUS")
			p := universal.GG_New_Status{}
			p.Deserialize(pRecv.Data)
			fmt.Println("New status: ", p.Status)
		default:
			fmt.Println("Received unknown packet, ignoring: ", pRecv.PacketType)
		}
	}
}

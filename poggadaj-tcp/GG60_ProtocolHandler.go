package main

import (
	"fmt"
	"poggadaj-tcp/gg60"
	"poggadaj-tcp/universal"
	"time"
)

func MsgChannel_GG60(currConn GGConnection) {
	for {
		msg := <-currConn.MsgChan
		fmt.Println(currConn.UIN, " received a message!")
		pS := universal.GG_Recv_MSG{
			Sender:   msg.From,
			Seq:      0,
			Time:     uint32(time.Now().Unix()),
			MsgClass: 0x08,
			Content:  msg.Content,
		}
		pOut := universal.InitGG_Packet(universal.GG_RECV_MSG, pS.Serialize())
		_, err := pOut.Send(currConn.Conn)
		if err != nil {
			fmt.Println("Error: ", err)
		}
	}
}

func Handle_GG60(currConn GGConnection, pRecv universal.GG_Packet) {
	// Handle the initial log in
	p := gg60.GG_Login60{}
	p.Deserialize(pRecv.Data)
	fmt.Println("Decoded data: ", p)

	currConn.UIN = p.UIN

	fmt.Println("Sending login response")
	if p.Hash == 4105424095 { // Password: 123
		currConn.Authenticated = true
		currConn.Status = p.Status
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

	// Create a message channel for the current user
	currConn.MsgChan = make(chan Message)
	(*currConn.MsgChans)[currConn.UIN] = currConn.MsgChan

	// Start a message sending channel
	go MsgChannel_GG60(currConn)

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
			currConn.Status = p.Status
			fmt.Println("New status: ", p.Status)
		case universal.GG_SEND_MSG:
			fmt.Println("Client is sending a message...")
			p := universal.GG_Send_MSG{}
			p.Deserialize(pRecv.Data, pRecv.Length)
			fmt.Printf("Recipient: %d, Message: %s\n", p.Recipient, p.Content)
			(*currConn.MsgChans)[p.Recipient] <- Message{currConn.UIN, p.Content}
		default:
			fmt.Printf("Received unknown packet, ignoring: 0x00%x\n", pRecv.PacketType)
		}
	}
}

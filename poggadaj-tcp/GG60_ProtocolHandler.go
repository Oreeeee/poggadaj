package main

import (
	"fmt"
	"poggadaj-tcp/gg60"
	"poggadaj-tcp/universal"
	"time"
)

func MsgChannel_GG60(currConn GGConnection) {
	pubsub := GetMessageChannel(currConn.UIN)
	for {
		msg := RecvMessageChannel(pubsub)

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
	passHash, _ := GetGG32Hash(currConn.UIN)
	if p.Hash == passHash {
		currConn.Authenticated = true
		currConn.Status = p.Status

		fmt.Println("Sending GG_LOGIN_OK")
		pOut := universal.InitGG_Packet(universal.GG_LOGIN_OK, []byte{})
		_, err := pOut.Send(currConn.Conn)
		if err != nil {
			fmt.Println("Error: ", err)
		}

		// Set user's status
		SetUserStatus(currConn.UIN, p.Status)
	} else {
		fmt.Println("Sending GG_LOGIN_FAILED")
		pOut := universal.InitGG_Packet(universal.GG_LOGIN_FAILED, []byte{})
		_, err := pOut.Send(currConn.Conn)
		if err != nil {
			fmt.Println("Error: ", err)
		}
		return
	}

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
		case universal.GG_NOTIFY_FIRST:
			fmt.Println("Received GG_NOTIFY_FIRST")
		case universal.GG_NOTIFY_LAST:
			fmt.Println("Received GG_NOTIFY_LAST")
			universal.GG_NotifyContactDeserialize(pRecv.Data, pRecv.Length, &currConn.NotifyList)
			fmt.Println(currConn.NotifyList)
		case universal.GG_ADD_NOTIFY:
			fmt.Println("Received GG_ADD_NOTIFY")
			universal.GG_AddNotify(pRecv.Data, &currConn.NotifyList)
			fmt.Println(currConn.NotifyList)
		case universal.GG_REMOVE_NOTIFY:
			fmt.Println("Received GG_REMOVE_NOTIFY (unimplemented)")
		case universal.GG_LIST_EMPTY:
			fmt.Println("Received GG_LIST_EMPTY")
		case universal.GG_NEW_STATUS:
			fmt.Println("Received GG_NEW_STATUS")

			p := universal.GG_New_Status{}
			p.Deserialize(pRecv.Data)

			SetUserStatus(currConn.UIN, p.Status)

			fmt.Println("New status: ", p.Status)
		case universal.GG_SEND_MSG:
			fmt.Println("Client is sending a message...")

			p := universal.GG_Send_MSG{}
			p.Deserialize(pRecv.Data, pRecv.Length)
			fmt.Printf("Recipient: %d, Message: %s\n", p.Recipient, p.Content)

			PublishMessageChannel(p.Recipient, Message{currConn.UIN, p.Content})
		default:
			fmt.Printf("Received unknown packet, ignoring: 0x00%x\n", pRecv.PacketType)
		}
	}
}

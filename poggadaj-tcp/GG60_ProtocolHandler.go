package main

import (
	"fmt"
	"poggadaj-tcp/gg60"
	"poggadaj-tcp/universal"
	"time"
)

func MsgChannel_GG60(currConn *GGConnection) {
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

func StatusChannel_GG60(currConn *GGConnection) {
	pubsub := GetStatusChannel()
	for {
		statusChange := RecvStatusChannel(pubsub)
		fmt.Println(statusChange)

		// Check if the status change is applicable for this connection
		for _, e := range currConn.NotifyList {
			if e.UIN == statusChange.UIN {
				fmt.Printf("%d's status change is relevant for %d\n", statusChange.UIN, currConn.UIN)

				status := statusChange.Status
				if status == universal.GG_STATUS_INVISIBLE {
					fmt.Println("Got GG_STATUS_INVISIBLE, sending GG_STATUS_NOT_AVAIL")
					status = universal.GG_STATUS_NOT_AVAIL
				}

				p := gg60.GG_Status60{
					UIN:        statusChange.UIN,
					Status:     uint8(status),
					RemoteIP:   0,
					RemotePort: 0,
					Version:    0,
					ImageSize:  0,
					Unknown1:   0,
				}
				pOut := universal.InitGG_Packet(universal.GG_STATUS60, p.Serialize())
				_, err := pOut.Send(currConn.Conn)
				if err != nil {
					fmt.Println("Error: ", err)
				}
			}
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

		// Change user's status to not available after disconnect
		defer SetUserStatus(currConn.UIN, universal.GG_STATUS_NOT_AVAIL)
	} else {
		fmt.Println("Sending GG_LOGIN_FAILED")
		pOut := universal.InitGG_Packet(universal.GG_LOGIN_FAILED, []byte{})
		_, err := pOut.Send(currConn.Conn)
		if err != nil {
			fmt.Println("Error: ", err)
		}
		return
	}

	// Start send channels
	go MsgChannel_GG60(&currConn)
	go StatusChannel_GG60(&currConn)

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
			universal.GG_NotifyContactDeserialize(pRecv.Data, pRecv.Length, &currConn.NotifyList)
		case universal.GG_NOTIFY_LAST:
			fmt.Println("Received GG_NOTIFY_LAST")
			universal.GG_NotifyContactDeserialize(pRecv.Data, pRecv.Length, &currConn.NotifyList)
			fmt.Println(currConn.NotifyList)

			// Respond with GG_NOTIFY_REPLY
			// TODO: unfuck this shit
			//response := make([]byte, 0)
			//buf := bytes.NewBuffer(response)
			//for _, notifyContact := range currConn.NotifyList {
			//	statusChange := universal.StatusChangeMsg{
			//		UIN:    notifyContact.UIN,
			//		Status: FetchUserStatus(notifyContact.UIN),
			//	}
			//	binary.Write(buf, binary.LittleEndian, universal.GG_NotifyReplySerialize(statusChange))
			//}
			//
			//pOut := universal.InitGG_Packet(universal.GG_NOTIFY_REPLY60, buf.Bytes())
			//_, err := pOut.Send(currConn.Conn)
			//if err != nil {
			//	fmt.Println("Error: ", err)
			//}
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
		case universal.GG_PING:
			fmt.Println("Received GG_PING")

			pOut := universal.InitGG_Packet(universal.GG_PONG, []byte{})
			_, err := pOut.Send(currConn.Conn)
			if err != nil {
				fmt.Println("Error: ", err)
			}
		default:
			fmt.Printf("Received unknown packet, ignoring: 0x00%x\n", pRecv.PacketType)
		}
	}
}

package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"poggadaj-tcp/gg60"
	"poggadaj-tcp/universal"
	"time"
)

func MsgChannel_GG60(currConn *GGConnection, run *bool) {
	defer Logger.Debugf("Quitting message channel")
	pubsub := GetMessageChannel(currConn.UIN)
	for *run {
		msg := RecvMessageChannel(pubsub)
		if !*run {
			// Sanity check to not accidentally write to a closed socket
			continue
		}

		Logger.Debugf("%d received a message!", currConn.UIN)
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
			Logger.Errorf("Error: %s", err)
		}
	}
}

func StatusChannel_GG60(currConn *GGConnection, run *bool) {
	defer Logger.Debugf("Quitting status channel")
	pubsub := GetStatusChannel()
	for *run {
		statusChange := RecvStatusChannel(pubsub)
		if !*run {
			// Sanity check to not accidentally write to a closed socket
			continue
		}
		//fmt.Println(statusChange)

		// Check if the status change is applicable for this connection
		for _, e := range currConn.NotifyList {
			if e.UIN == statusChange.UIN {
				Logger.Debugf("%d's status change is relevant for %d", statusChange.UIN, currConn.UIN)

				status := statusChange.Status
				switch status {
				case universal.GG_STATUS_INVISIBLE:
					Logger.Debugf("Got GG_STATUS_INVISIBLE, sending GG_STATUS_NOT_AVAIL")
					status = universal.GG_STATUS_NOT_AVAIL
				case universal.GG_STATUS_INVISIBLE_DESCR:
					Logger.Debugf("Got GG_STATUS_INVISIBLE_DESCR, sending GG_STATUS_NOT_AVAIL_DESCR")
					status = universal.GG_STATUS_NOT_AVAIL_DESCR
				}

				p := gg60.GG_Status60{
					UIN:         statusChange.UIN,
					Status:      uint8(status),
					RemoteIP:    0,
					RemotePort:  0,
					Version:     0,
					ImageSize:   0,
					Unknown1:    0,
					Description: statusChange.Description,
				}
				pOut := universal.InitGG_Packet(universal.GG_STATUS60, p.Serialize())
				_, err := pOut.Send(currConn.Conn)
				if err != nil {
					Logger.Errorf("Error: %s", err)
				}
			}
		}
	}
}

func Handle_GG60(currConn GGConnection, pRecv universal.GG_Packet) {
	// Handle the initial log in
	p := gg60.GG_Login60{}
	p.Deserialize(pRecv.Data)

	currConn.UIN = p.UIN

	Logger.Debugf("Sending login response")
	passHash, _ := GetGG32Hash(currConn.UIN)
	if p.Hash == passHash {
		currConn.Authenticated = true
		currConn.Status = p.Status

		Logger.Debugf("Sending GG_LOGIN_OK")
		pOut := universal.InitGG_Packet(universal.GG_LOGIN_OK, []byte{})
		_, err := pOut.Send(currConn.Conn)
		if err != nil {
			fmt.Println("Error: ", err)
		}

		// Set user's status
		SetUserStatus(universal.StatusChangeMsg{
			UIN:    currConn.UIN,
			Status: p.Status,
		})

		// Change user's status to not available after disconnect
		defer SetUserStatus(universal.StatusChangeMsg{
			UIN:    currConn.UIN,
			Status: universal.GG_STATUS_NOT_AVAIL,
		})
	} else {
		Logger.Debugf("Sending GG_LOGIN_FAILED")
		pOut := universal.InitGG_Packet(universal.GG_LOGIN_FAILED, []byte{})
		_, err := pOut.Send(currConn.Conn)
		if err != nil {
			fmt.Println("Error: ", err)
		}
		return
	}

	// Start send channels
	runMsgChannel := true
	runStatusChannel := true
	go MsgChannel_GG60(&currConn, &runMsgChannel)
	go StatusChannel_GG60(&currConn, &runStatusChannel)
	defer func(r *bool) { *r = false }(&runMsgChannel)
	defer func(r *bool) { *r = false }(&runStatusChannel)

	// Connection loop
	for {
		pRecv := universal.GG_Packet{}
		err := pRecv.Receive(currConn.Conn)
		if err != nil {
			Logger.Errorf("Error receiving data, dropping connection!: %s", err)
			return
		}

		switch pRecv.PacketType {
		case universal.GG_NOTIFY_FIRST:
			Logger.Debugf("Received GG_NOTIFY_FIRST")
			universal.GG_NotifyContactDeserialize(pRecv.Data, pRecv.Length, &currConn.NotifyList)
		case universal.GG_NOTIFY_LAST:
			Logger.Debugf("Received GG_NOTIFY_LAST")
			universal.GG_NotifyContactDeserialize(pRecv.Data, pRecv.Length, &currConn.NotifyList)
			//fmt.Println(currConn.NotifyList)

			// Respond with GG_NOTIFY_REPLY
			response := make([]byte, 0)
			buf := bytes.NewBuffer(response)
			for _, notifyContact := range currConn.NotifyList {
				statusChange := FetchUserStatus(notifyContact.UIN)
				binary.Write(buf, binary.LittleEndian, universal.GG_NotifyReplySerialize(statusChange))
			}

			pOut := universal.InitGG_Packet(universal.GG_NOTIFY_REPLY60, buf.Bytes())
			_, err := pOut.Send(currConn.Conn)
			if err != nil {
				Logger.Debugf("Error: %s", err)
			}
		case universal.GG_ADD_NOTIFY:
			Logger.Debugf("Received GG_ADD_NOTIFY")
			universal.GG_AddNotify(pRecv.Data, &currConn.NotifyList)
			//fmt.Println(currConn.NotifyList)
		case universal.GG_REMOVE_NOTIFY:
			Logger.Debugf("Received GG_REMOVE_NOTIFY (unimplemented)")
		case universal.GG_LIST_EMPTY:
			Logger.Debugf("Received GG_LIST_EMPTY")
		case universal.GG_NEW_STATUS:
			Logger.Debugf("Received GG_NEW_STATUS")

			p := universal.GG_New_Status{}
			p.Deserialize(pRecv.Data, pRecv.Length)

			SetUserStatus(universal.StatusChangeMsg{
				UIN:         currConn.UIN,
				Status:      p.Status,
				Description: p.Description,
			})

			Logger.Debugf("New status: 0x00%x, Description: %s", p.Status, p.Description)
		case universal.GG_SEND_MSG:
			Logger.Debugf("Client is sending a message...")

			p := universal.GG_Send_MSG{}
			p.Deserialize(pRecv.Data, pRecv.Length)
			//fmt.Printf("Recipient: %d, Message: %s\n", p.Recipient, p.Content)

			PublishMessageChannel(p.Recipient, Message{currConn.UIN, p.Content})
		case universal.GG_PING:
			Logger.Debugf("Received GG_PING")

			pOut := universal.InitGG_Packet(universal.GG_PONG, []byte{})
			_, err := pOut.Send(currConn.Conn)
			if err != nil {
				Logger.Errorf("Error: %s", err)
			}
		default:
			Logger.Warnf("Received unknown packet, ignoring: 0x00%x\n", pRecv.PacketType)
		}
	}
}

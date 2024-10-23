package main

import (
	"poggadaj-tcp/database"
	"poggadaj-tcp/gg60"
	"poggadaj-tcp/logging"
	"poggadaj-tcp/structs"
	"poggadaj-tcp/universal"
)

func MsgChannel_GG60(c GGClient, cI *structs.ClientInfo, run *bool) {
	defer logging.L.Debugf("Quitting message channel")
	pubsub := database.GetMessageChannel(cI.UIN)
	for *run {
		msg := database.RecvMessageChannel(pubsub)
		if !*run {
			// Sanity check to not accidentally write to a closed socket
			continue
		}

		logging.L.Debugf("%d received a message!", cI.UIN)
		c.SendRecvMsg(msg)
	}
}

func StatusChannel_GG60(c GGClient, cI *structs.ClientInfo, run *bool) {
	defer logging.L.Debugf("Quitting status channel")
	pubsub := database.GetStatusChannel()
	for *run {
		statusChange := database.RecvStatusChannel(pubsub)
		if !*run {
			// Sanity check to not accidentally write to a closed socket
			continue
		}

		// Check if the status change is applicable for this connection
		for _, e := range cI.NotifyList {
			if e.UIN == statusChange.UIN {
				logging.L.Debugf("%d's status change is relevant for %d", statusChange.UIN, cI.UIN)

				switch statusChange.Status {
				case universal.GG_STATUS_INVISIBLE:
					logging.L.Debugf("Got GG_STATUS_INVISIBLE, sending GG_STATUS_NOT_AVAIL")
					statusChange.Status = universal.GG_STATUS_NOT_AVAIL
				case universal.GG_STATUS_INVISIBLE_DESCR:
					logging.L.Debugf("Got GG_STATUS_INVISIBLE_DESCR, sending GG_STATUS_NOT_AVAIL_DESCR")
					statusChange.Status = universal.GG_STATUS_NOT_AVAIL_DESCR
				}

				c.SendStatus(statusChange)
			}
		}
	}
}

func Handle_GG60(currConn GGConnection, pRecv universal.GG_Packet) {
	// Handle the initial log in
	client := gg60.GG60Client{}
	clientInfo := client.GetClientInfoPtr()
	clientInfo.Conn = currConn.Conn
	client.HandleLogin(pRecv)

	if !clientInfo.Authenticated {
		return
	}

	defer client.Clean()

	// Start send channels
	runMsgChannel := true
	runStatusChannel := true
	go MsgChannel_GG60(&client, clientInfo, &runMsgChannel)
	go StatusChannel_GG60(&client, clientInfo, &runStatusChannel)
	defer func(r *bool) { *r = false }(&runMsgChannel)
	defer func(r *bool) { *r = false }(&runStatusChannel)

	// Connection loop
	for {
		pRecv := universal.GG_Packet{}
		err := pRecv.Receive(clientInfo.Conn)
		if err != nil {
			logging.L.Errorf("Error receiving data, dropping connection!: %s", err)
			return
		}

		switch pRecv.PacketType {
		case universal.GG_NOTIFY_FIRST:
			logging.L.Debugf("Received GG_NOTIFY_FIRST")
			client.HandleNotifyFirst(pRecv)
		case universal.GG_NOTIFY_LAST:
			logging.L.Debugf("Received GG_NOTIFY_LAST")
			client.HandleNotifyLast(pRecv)
		case universal.GG_ADD_NOTIFY:
			logging.L.Debugf("Received GG_ADD_NOTIFY")
			client.HandleAddNotify(pRecv)
		case universal.GG_REMOVE_NOTIFY:
			logging.L.Debugf("Received GG_REMOVE_NOTIFY (unimplemented)")
			client.HandleRemoveNotify(pRecv)
		case universal.GG_LIST_EMPTY:
			logging.L.Debugf("Received GG_LIST_EMPTY")
		case universal.GG_NEW_STATUS:
			logging.L.Debugf("Received GG_NEW_STATUS")
			client.HandleNewStatus(pRecv)
		case universal.GG_SEND_MSG:
			logging.L.Debugf("Client is sending a message...")
			client.HandleSendMsg(pRecv)
		case universal.GG_PING:
			logging.L.Debugf("Received GG_PING")
			client.SendPong()
		default:
			logging.L.Warnf("Received unknown packet, ignoring: 0x00%x\n", pRecv.PacketType)
		}
	}
}

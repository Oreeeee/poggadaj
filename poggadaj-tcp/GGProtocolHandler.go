package main

import (
	"net"
	"poggadaj-tcp/gg60"
	"poggadaj-tcp/gg70"
	"poggadaj-tcp/logging"
	"poggadaj-tcp/universal"
	"poggadaj-tcp/utils"
)

func HandleConnection(conn net.Conn) {
	defer conn.Close()

	// Here we create a GG_WELCOME packet once the client connects to the server
	ggw := universal.InitGG_Welcome()
	ggwB := ggw.Serialize()
	packet := universal.InitGG_Packet(universal.GG_WELCOME, ggwB)

	_, err := packet.Send(conn)
	if err != nil {
		logging.L.Errorf("Error: %s", err)
	}

	// Wait for the next packet, which will tell us the protocol version handler we need
	pRecv := universal.GG_Packet{}
	if pRecv.Receive(conn) != nil {
		logging.L.Errorf("Error receiving data, dropping connection!: %s", err)
		return
	}

	var client GGClient
	switch pRecv.PacketType {
	case gg60.GG_LOGIN60:
		logging.L.Infof("Gadu-Gadu 6.0 protocol detected")
		client = &gg60.GG60Client{}
	case gg70.GG_LOGIN70:
		logging.L.Infof("Gadu-Gadu 7.0 protocol detected")
		client = &gg70.GG70Client{}
	default:
		logging.L.Infof("Unknown protocol version!")
	}

	clientInfo := client.GetClientInfoPtr()
	clientInfo.Conn = conn
	client.HandleLogin(pRecv)

	if !clientInfo.Authenticated {
		return
	}

	defer client.Clean()

	// Start send channels
	runMsgChannel := true
	runStatusChannel := true
	go MsgChannel(client, clientInfo, &runMsgChannel)
	go StatusChannel(client, clientInfo, &runStatusChannel)
	defer utils.CloseChannel(&runMsgChannel)
	defer utils.CloseChannel(&runStatusChannel)

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
			logging.L.Debugf("Received GG_REMOVE_NOTIFY")
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

package main

import (
	"net"
	"poggadaj-tcp/clients"
	"poggadaj-tcp/logging"
	"poggadaj-tcp/protocol/packets"
	"poggadaj-tcp/protocol/packets/c2s"
	"poggadaj-tcp/protocol/packets/s2c"
	"poggadaj-tcp/utils"
)

func HandleConnection(conn net.Conn) {
	defer conn.Close()

	// Here we create a GG_WELCOME packet once the client connects to the server
	ggw := s2c.InitGG_Welcome()
	ggwB := ggw.Serialize()
	packet := packets.InitGG_Packet(s2c.GG_WELCOME, ggwB)

	_, err := packet.Send(conn)
	if err != nil {
		logging.L.Errorf("Error: %s", err)
	}

	// Wait for the next packet, which will tell us the protocol version handler we need
	pRecv := packets.GG_Packet{}
	if pRecv.Receive(conn) != nil {
		logging.L.Errorf("Error receiving data, dropping connection!: %s", err)
		return
	}

	client := clients.GGClient{}
	switch pRecv.PacketType {
	case c2s.GG_LOGIN30:
		logging.L.Infof("Ancient Gadu-Gadu protocol detected")
	case c2s.GG_LOGIN:
		logging.L.Infof("Gadu-Gadu late 4.x - 6.0 protocol detected")
	case c2s.GG_LOGIN60:
		logging.L.Infof("Gadu-Gadu 6.0 protocol detected")
	case c2s.GG_LOGIN70:
		logging.L.Infof("Gadu-Gadu 7.0 protocol detected")
	default:
		logging.L.Infof("Unknown protocol version!")
	}

	client.Conn = conn
	client.HandleLogin(pRecv.PacketType, pRecv)

	if !client.Authenticated {
		return
	}

	defer client.Clean()

	// Start send channels
	runMsgChannel := true
	runStatusChannel := true
	go MsgChannel(&client, &runMsgChannel)
	go StatusChannel(&client, &runStatusChannel)
	defer utils.CloseChannel(&runMsgChannel)
	defer utils.CloseChannel(&runStatusChannel)

	// Connection loop
	for {
		pRecv := packets.GG_Packet{}
		err := pRecv.Receive(client.Conn)
		if err != nil {
			logging.L.Errorf("Error receiving data, dropping connection!: %s", err)
			return
		}

		switch pRecv.PacketType {
		case c2s.GG_NOTIFY_FIRST:
			logging.L.Debugf("Received GG_NOTIFY_FIRST")
			client.HandleNotifyFirst(pRecv)
		case c2s.GG_NOTIFY_LAST:
			logging.L.Debugf("Received GG_NOTIFY_LAST")
			client.HandleNotifyLast(pRecv)
		case c2s.GG_ADD_NOTIFY:
			logging.L.Debugf("Received GG_ADD_NOTIFY")
			client.HandleAddNotify(pRecv)
		case c2s.GG_REMOVE_NOTIFY:
			logging.L.Debugf("Received GG_REMOVE_NOTIFY")
			client.HandleRemoveNotify(pRecv)
		case c2s.GG_LIST_EMPTY:
			logging.L.Debugf("Received GG_LIST_EMPTY")
		case c2s.GG_NEW_STATUS:
			logging.L.Debugf("Received GG_NEW_STATUS")
			client.HandleNewStatus(pRecv)
		case c2s.GG_SEND_MSG:
			logging.L.Debugf("Client is sending a message...")
			client.HandleSendMsg(pRecv)
		case c2s.GG_PING:
			logging.L.Debugf("Received GG_PING")
			client.SendPong()
		default:
			logging.L.Warnf("Received unknown packet, ignoring: 0x00%x\n", pRecv.PacketType)
		}
	}
}

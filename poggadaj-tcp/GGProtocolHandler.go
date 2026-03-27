// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package main

import (
	"net"
	"poggadaj-shared/logging"
	"poggadaj-tcp/clients"
	"poggadaj-tcp/protocol"
	"poggadaj-tcp/utils"
)

func HandleConnection(conn net.Conn) {
	defer conn.Close()

	// Here we create a GG_WELCOME packet once the client connects to the server
	ggw := protocol.InitGG_Welcome()
	ggwB := ggw.Serialize()
	packet := protocol.InitGG_Packet(protocol.GG_WELCOME, ggwB)

	_, err := packet.Send(conn)
	if err != nil {
		logging.L.Errorf("Error: %s", err)
	}

	// Wait for the next packet, which will tell us the protocol version handler we need
	pRecv := protocol.GG_Packet{}
	if pRecv.Receive(conn) != nil {
		logging.L.Errorf("Error receiving data, dropping connection!: %s", err)
		return
	}

	client := clients.GGClient{}
	switch pRecv.PacketType {
	case protocol.GG_LOGIN30:
		logging.L.Infof("Ancient Gadu-Gadu protocol detected")
	case protocol.GG_LOGIN:
		logging.L.Infof("Gadu-Gadu late 4.x - 6.0 protocol detected")
	case protocol.GG_LOGIN60:
		logging.L.Infof("Gadu-Gadu 6.0 protocol detected")
	case protocol.GG_LOGIN70:
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
		pRecv := protocol.GG_Packet{}
		err := pRecv.Receive(client.Conn)
		if err != nil {
			logging.L.Errorf("Error receiving data, dropping connection!: %s", err)
			return
		}

		switch pRecv.PacketType {
		case protocol.GG_NOTIFY30:
			logging.L.Debugf("Received GG_NOTIFY30")
			client.HandleNotify30(pRecv)
		case protocol.GG_NOTIFY_FIRST:
			logging.L.Debugf("Received GG_NOTIFY_FIRST")
			client.HandleNotifyFirst(pRecv)
		case protocol.GG_NOTIFY_LAST:
			logging.L.Debugf("Received GG_NOTIFY_LAST")
			client.HandleNotifyLast(pRecv)
		case protocol.GG_ADD_NOTIFY:
			logging.L.Debugf("Received GG_ADD_NOTIFY")
			client.HandleAddNotify(pRecv)
		case protocol.GG_REMOVE_NOTIFY:
			logging.L.Debugf("Received GG_REMOVE_NOTIFY")
			client.HandleRemoveNotify(pRecv)
		case protocol.GG_LIST_EMPTY:
			logging.L.Debugf("Received GG_LIST_EMPTY")
		case protocol.GG_NEW_STATUS:
			logging.L.Debugf("Received GG_NEW_STATUS")
			client.HandleNewStatus(pRecv)
		case protocol.GG_SEND_MSG:
			logging.L.Debugf("Client is sending a message...")
			client.HandleSendMsg(pRecv)
		case protocol.GG_USERLIST_REQUEST:
			logging.L.Debugf("Received GG_USERLIST_REQUEST")
			client.HandleUserlistReq(pRecv)
		case protocol.GG_PUBDIR50_REQUEST:
			logging.L.Debugf("Received GG_PUBDIR50_REQUEST")
			client.HandlePubdirReq(pRecv)
		case protocol.GG_PING:
			logging.L.Debugf("Received GG_PING")
			client.SendPong()
		default:
			logging.L.Warnf("Received unknown packet, ignoring: 0x00%x\n", pRecv.PacketType)
		}
	}
}

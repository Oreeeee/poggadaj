// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"charm.land/log/v2"
	"codeberg.org/or3e/poggadaj/cmd/poggadaj-tcp/protocol"
	"codeberg.org/or3e/poggadaj/internal/constants"
	"codeberg.org/or3e/poggadaj/internal/logging"
	"codeberg.org/or3e/poggadaj/internal/statuses"
	"codeberg.org/or3e/poggadaj/internal/structs"
	"codeberg.org/or3e/poggadaj/internal/utils"
	"golang.org/x/text/encoding/charmap"
)

type Client struct {
	conn   net.Conn
	server *Server
	logger *log.Logger

	UIN           uint32
	Status        uint32
	Authenticated bool
	NotifyList    []structs.GG_NotifyContact
	Version       uint8
	VOIP          bool
	ProtocolLevel uint8
	UserListBuf   []string
}

func (client *Client) Run() {
	defer client.conn.Close()

	stream := utils.NewIOStream([]byte{}, binary.LittleEndian, charmap.Windows1250)

	// Here we create a GG_WELCOME packet once the client connects to the server
	ggw := protocol.InitGG_Welcome()
	packet := protocol.InitGG_Packet(protocol.GG_WELCOME, ggw)

	_, err := packet.Send(client.conn)
	if err != nil {
		client.logger.Errorf("Error: %s", err)
	}

	defer client.Clean()

	// Connection loop
	for {
		pRecv, err := protocol.ReceivePacket(client.conn)
		if err != nil {
			client.logger.Errorf("Error receiving data, dropping connection!: %s", err)
			return
		}

		stream.Reset(pRecv.Data)

		// Handle the authentication of the user
		if !client.Authenticated {
			hasHandler := false
			if handler, ok := LoginHandlers[pRecv.PacketType]; ok {
				switch pRecv.PacketType {
				case protocol.GG_LOGIN30:
					client.logger.Infof("Ancient Gadu-Gadu protocol detected")
				case protocol.GG_LOGIN:
					client.logger.Infof("Gadu-Gadu late 4.x - 6.0 protocol detected")
				case protocol.GG_LOGIN60:
					client.logger.Infof("Gadu-Gadu 6.0 protocol detected")
				case protocol.GG_LOGIN70:
					client.logger.Infof("Gadu-Gadu 7.0 protocol detected")
				default:
					client.logger.Infof("Unknown protocol version!")
				}

				client.logger.Infof("Received login packet %x\n", pRecv.PacketType)
				hasHandler = true
				err = handler(client, stream)
			}

			if err != nil || !hasHandler {
				client.logger.Warn("Client failed to authenticate! Disconnecting!")
				client.SendLoginFail()
				return
			}

			client.logger.Info("Authenticated!")
			client.logger.Debug("Sending login response")
			client.Authenticated = true
			client.SendLoginOK()
			client.server.cache.SetUserStatus(structs.StatusChangeMsg{
				UIN:    client.UIN,
				Status: client.Status,
			})
			// Start send channels
			runMsgChannel := true
			runStatusChannel := true
			go client.MsgChannel(&runMsgChannel)
			go client.StatusChannel(&runStatusChannel)
			defer utils.CloseChannel(&runMsgChannel)
			defer utils.CloseChannel(&runStatusChannel)

			continue
		}

		if handler, ok := Handlers[pRecv.PacketType]; ok {
			client.logger.Infof("Received packet %x\n", pRecv.PacketType)
			err = handler(client, stream)
			if err != nil {
				client.logger.Error("Failed to handle packet", "packetId", pRecv.PacketType, "err", err)
			}
		} else {
			client.logger.Warnf("Received unknown packet, ignoring: 0x00%x\n", pRecv.PacketType)
		}
	}
}

func (c *Client) PutUserList() {
	err := c.server.db.DeleteUserList(c.UIN) // Delete user's contact list, as we are writing to the list and not appending to it
	if err != nil {
		c.logger.Errorf("Failed to delete user list: %s", err)
	}

	userListStr := strings.Join(c.UserListBuf, "")                   // Combine the buffer into one string
	userListSeparated := strings.Split(userListStr, "\r\n")          // Separate the request lines
	userListSeparated = userListSeparated[:len(userListSeparated)-1] // Remove the last (empty) index

	// Convert all the strings to UserListRequest objects
	var userlist []structs.UserListRequest
	for _, str := range userListSeparated {
		c.logger.Debugf("Read userlist: %s", strconv.Quote(str))
		user := structs.UserListRequest{}
		err := user.Read(str)
		if err != nil {
			c.logger.Errorf("Error parsing userlist line: %v", err)
		}
		userlist = append(userlist, user)
	}

	c.logger.Debugf("Received userlist put: %v", userlist)
	c.logger.Debugf("Putting userlist into the database")
	c.server.db.PutUserList(userlist, c.UIN)

	// Send acknowledgement that the server received the list
	for i, _ := range c.UserListBuf {
		c.SendPutUserListAck(i)
	}

	// Clear the buffer
	c.UserListBuf = []string{}
}

func (c *Client) SendPubdirResp(Type uint8, seq uint32, contents []byte) {
	p := protocol.GG_Pubdir50_Reply{
		Type:  Type,
		Seq:   seq,
		Reply: contents,
	}
	logging.StructPPrint(c.logger, "GG_PUBDIR50_REPLY", p.PrettyPrint())
	pOut := protocol.InitGG_Packet(protocol.GG_PUBDIR50_REPLY, &p)
	_, err := pOut.Send(c.conn)
	if err != nil {
		c.logger.Errorf("Error: %s", err)
	}
}

func (c *Client) SendPutUserListAck(i int) {
	var reqType uint8
	if i == 0 {
		reqType = constants.GG_USERLIST_PUT_REPLY
	} else {
		reqType = constants.GG_USERLIST_PUT_MORE_REPLY
	}

	p := protocol.GG_Userlist_Reply{
		Type:    reqType,
		Request: []byte(c.UserListBuf[i]),
	}
	logging.StructPPrint(c.logger, "GG_USERLIST_REPLY", p.PrettyPrint())
	pOut := protocol.InitGG_Packet(protocol.GG_USERLIST_REPLY, &p)
	_, err := pOut.Send(c.conn)
	if err != nil {
		c.logger.Errorf("Error: %s", err)
	}
}

func (c *Client) SendGetUserListResp(userListBuf string) {
	chunkedList := utils.ChunkString(userListBuf, 2048)
	lastIndex := len(chunkedList) - 1
	for i, str := range chunkedList {
		replyType := constants.GG_USERLIST_GET_MORE_REPLY
		if i == lastIndex {
			// The last part of a list import is type GG_USERLIST_GET_REPLY
			replyType = constants.GG_USERLIST_GET_REPLY
		}
		p := protocol.GG_Userlist_Reply{
			Type:    uint8(replyType),
			Request: []byte(str),
		}
		logging.StructPPrint(c.logger, "GG_USERLIST_REPLY", p.PrettyPrint())
		pOut := protocol.InitGG_Packet(protocol.GG_USERLIST_REPLY, &p)
		_, err := pOut.Send(c.conn)
		if err != nil {
			c.logger.Errorf("Error: %s", err)
		}
	}
}

func (c *Client) SendLoginOK() {
	pOut := protocol.InitEmptyGG_Packet(protocol.GG_LOGIN_OK)
	_, err := pOut.Send(c.conn)
	if err != nil {
		fmt.Println("Error: ", err)
	}
}

func (c *Client) SendLoginFail() {
	pOut := protocol.InitEmptyGG_Packet(protocol.GG_LOGIN_FAILED)
	_, err := pOut.Send(c.conn)
	if err != nil {
		fmt.Println("Error: ", err)
	}
}

func (c *Client) SendStatus(statusChange structs.StatusChangeMsg) {
	if c.Version >= 0x2a {
		c.SendStatus77(statusChange)
	} else if c.Version >= 0x20 {
		c.SendStatus60(statusChange)
	} else if c.Version <= 0x18 {
		c.SendStatus50(statusChange)
	}
}

func (c *Client) SendStatus50(statusChange structs.StatusChangeMsg) {
	p := protocol.GG_Status{
		UIN:         statusChange.UIN,
		Status:      statusChange.Status,
		Description: statusChange.Description,
	}
	logging.StructPPrint(c.logger, "GG_STATUS", p.PrettyPrint())
	pOut := protocol.InitGG_Packet(protocol.GG_STATUS, &p)
	_, err := pOut.Send(c.conn)
	if err != nil {
		c.logger.Errorf("Error: %s", err)
	}
}

func (c *Client) SendStatus60(statusChange structs.StatusChangeMsg) {
	p := protocol.GG_Status60{
		UIN:         statusChange.UIN,
		Status:      uint8(statusChange.Status),
		RemoteIP:    0,
		RemotePort:  0,
		Version:     0,
		ImageSize:   0,
		Description: statusChange.Description,
	}
	logging.StructPPrint(c.logger, "GG_STATUS60", p.PrettyPrint())
	pOut := protocol.InitGG_Packet(protocol.GG_STATUS60, &p)
	_, err := pOut.Send(c.conn)
	if err != nil {
		c.logger.Errorf("Error: %s", err)
	}
}

func (c *Client) SendStatus77(statusChange structs.StatusChangeMsg) {
	p := protocol.GG_Status77{
		UIN:         statusChange.UIN,
		Status:      uint8(statusChange.Status),
		RemoteIP:    0,
		RemotePort:  0,
		Version:     0,
		ImageSize:   0,
		Description: statusChange.Description,
	}
	logging.StructPPrint(c.logger, "GG_STATUS77", p.PrettyPrint())
	pOut := protocol.InitGG_Packet(protocol.GG_STATUS77, &p)
	_, err := pOut.Send(c.conn)
	if err != nil {
		c.logger.Errorf("Error: %s", err)
	}
}

func (c *Client) SendRecvMsg(msg structs.Message) {
	var messagePacket protocol.GG_Packet_Iface
	var packetId uint32

	if c.ProtocolLevel == 30 {
		messagePacket = &protocol.GG_Recv_MSG30{
			Sender:  msg.From,
			Content: string(msg.Content),
		}
		packetId = protocol.GG_RECV_MSG30
	} else {
		messagePacket = &protocol.GG_Recv_MSG{
			Sender:   msg.From,
			Seq:      0,
			Time:     uint32(time.Now().Unix()),
			MsgClass: msg.MsgClass,
			Content:  string(msg.Content),
		}
		packetId = protocol.GG_RECV_MSG
	}

	//logging.StructPPrint(c.logger, "GG_RECV_MSG", messagePacket.PrettyPrint())
	pOut := protocol.InitGG_Packet(packetId, messagePacket)
	_, err := pOut.Send(c.conn)
	if err != nil {
		c.logger.Errorf("Error: %s", err)
	}
}

func (c *Client) SendNotifyReply(data protocol.GG_Packet_Iface) {
	var pOut *protocol.GG_Packet
	if c.Version >= 0x2a {
		pOut = protocol.InitGG_Packet(protocol.GG_NOTIFY_REPLY77, data)
	} else {
		pOut = protocol.InitGG_Packet(protocol.GG_NOTIFY_REPLY60, data)
	}
	_, err := pOut.Send(c.conn)
	if err != nil {
		c.logger.Debugf("Error: %s", err)
	}
}

func (c *Client) SendPong() {
	pOut := protocol.InitEmptyGG_Packet(protocol.GG_PONG)
	_, err := pOut.Send(c.conn)
	if err != nil {
		c.logger.Errorf("Error: %s", err)
	}
}

func (c *Client) Clean() {
	// Change user's status to not available
	c.server.cache.SetUserStatus(structs.StatusChangeMsg{
		UIN:    c.UIN,
		Status: statuses.GG_STATUS_NOT_AVAIL,
	})
}

func NewClient(conn net.Conn, server *Server, logger *log.Logger) (*Client, error) {
	return &Client{
		conn:   conn,
		server: server,
		logger: logger,
	}, nil
}

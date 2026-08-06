package main

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"charm.land/log/v2"
	"codeberg.org/or3e/poggadaj/cmd/poggadaj-tcp/constants"
	"codeberg.org/or3e/poggadaj/cmd/poggadaj-tcp/protocol"
	"codeberg.org/or3e/poggadaj/cmd/poggadaj-tcp/pubdir"
	"codeberg.org/or3e/poggadaj/internal/cache"
	"codeberg.org/or3e/poggadaj/internal/statuses"
	"codeberg.org/or3e/poggadaj/internal/structs"
	"codeberg.org/or3e/poggadaj/internal/utils"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
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

	// Wait for the next packet, which will tell us the protocol version handler we need
	pRecv, err := protocol.ReceivePacket(client.conn)
	if err != nil {
		client.logger.Errorf("Error receiving data, dropping connection!: %s", err)
		return
	}
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

	client.HandleLogin(pRecv.PacketType, utils.NewIOStream(pRecv.Data, binary.LittleEndian, charmap.Windows1250))

	if !client.Authenticated {
		return
	}

	//defer client.Clean()

	// Start send channels
	runMsgChannel := true
	runStatusChannel := true
	go client.MsgChannel(&runMsgChannel)
	go client.StatusChannel(&runStatusChannel)
	defer utils.CloseChannel(&runMsgChannel)
	defer utils.CloseChannel(&runStatusChannel)

	// Connection loop
	for {
		pRecv, err := protocol.ReceivePacket(client.conn)
		if err != nil {
			client.logger.Errorf("Error receiving data, dropping connection!: %s", err)
			return
		}

		stream.Reset(pRecv.Data)

		switch pRecv.PacketType {
		case protocol.GG_NOTIFY30:
			client.logger.Debugf("Received GG_NOTIFY30")
			client.HandleNotify30(stream)
		case protocol.GG_NOTIFY_FIRST:
			client.logger.Debugf("Received GG_NOTIFY_FIRST")
			client.HandleNotifyFirst(stream)
		case protocol.GG_NOTIFY_LAST:
			client.logger.Debugf("Received GG_NOTIFY_LAST")
			client.HandleNotifyLast(stream)
		case protocol.GG_ADD_NOTIFY:
			client.logger.Debugf("Received GG_ADD_NOTIFY")
			client.HandleAddNotify(stream)
		case protocol.GG_REMOVE_NOTIFY:
			client.logger.Debugf("Received GG_REMOVE_NOTIFY")
			client.HandleRemoveNotify(stream)
		case protocol.GG_LIST_EMPTY:
			client.logger.Debugf("Received GG_LIST_EMPTY")
		case protocol.GG_NEW_STATUS:
			client.logger.Debugf("Received GG_NEW_STATUS")
			client.HandleNewStatus(stream)
		case protocol.GG_SEND_MSG:
			client.logger.Debugf("Client is sending a message...")
			client.HandleSendMsg(stream)
		case protocol.GG_USERLIST_REQUEST:
			client.logger.Debugf("Received GG_USERLIST_REQUEST")
			client.HandleUserlistReq(stream)
		case protocol.GG_PUBDIR50_REQUEST:
			client.logger.Debugf("Received GG_PUBDIR50_REQUEST")
			client.HandlePubdirReq(stream)
		case protocol.GG_PING:
			client.logger.Debugf("Received GG_PING")
			client.SendPong()
		default:
			client.logger.Warnf("Received unknown packet, ignoring: 0x00%x\n", pRecv.PacketType)
		}
	}
}

func (c *Client) HandleLogin(packetType uint32, data *utils.IOStream) bool {
	switch packetType {
	case protocol.GG_LOGIN30:
		c.ProtocolLevel = 30
		p := protocol.GG_Login30{}
		p.Deserialize(data)
		log.StructPPrint("GG_LOGIN30", p.PrettyPrint())

		c.UIN = p.UIN

		c.logger.Debugf("Sending login response")
		passHash, _ := c.server.db.GetAncientHash(c.UIN)
		if p.Hash == passHash {
			c.Authenticated = true
			c.Status = p.Status

			c.logger.Debugf("Sending GG_LOGIN_OK")
			c.SendLoginOK()

			cache.SetUserStatus(structs.StatusChangeMsg{
				UIN:    c.UIN,
				Status: p.Status,
			})

			return true
		}
		return false
	case protocol.GG_LOGIN:
		p := protocol.GG_Login{}
		p.Deserialize(data)
		log.StructPPrint("GG_LOGIN", p.PrettyPrint())

		c.UIN = p.UIN

		c.logger.Debugf("Sending login response")
		passHash, _ := c.server.db.GetGG32Hash(c.UIN)
		if p.Hash == passHash {
			c.Authenticated = true
			c.Status = p.Status

			c.logger.Debugf("Sending GG_LOGIN_OK")
			c.SendLoginOK()

			// Set user's status
			cache.SetUserStatus(structs.StatusChangeMsg{
				UIN:    c.UIN,
				Status: p.Status,
			})

			c.Version, c.VOIP = utils.GetVersionAndVOIP(p.Version)

			return true
		} else {
			c.logger.Debugf("Sending GG_LOGIN_FAILED")
			c.SendLoginFail()
			return false
		}
		return false
	case protocol.GG_LOGIN60:
		p := protocol.GG_Login60{}
		p.Deserialize(data)
		log.StructPPrint("GG_LOGIN60", p.PrettyPrint())

		c.UIN = p.UIN

		c.logger.Debugf("Sending login response")
		passHash, _ := c.server.db.GetGG32Hash(c.UIN)
		if p.Hash == passHash {
			c.Authenticated = true
			c.Status = p.Status

			c.logger.Debugf("Sending GG_LOGIN_OK")
			c.SendLoginOK()

			// Set user's status
			cache.SetUserStatus(structs.StatusChangeMsg{
				UIN:    c.UIN,
				Status: p.Status,
			})

			c.Version, c.VOIP = utils.GetVersionAndVOIP(p.Version)

			return true
		} else {
			c.logger.Debugf("Sending GG_LOGIN_FAILED")
			c.SendLoginFail()
			return false
		}
		return false
	case protocol.GG_LOGIN70:
		p := protocol.GG_Login70{}
		p.Deserialize(data)
		log.StructPPrint("GG_LOGIN70", p.PrettyPrint())

		c.UIN = p.UIN

		c.logger.Debugf("Sending login response")
		passHash, _ := c.server.db.GetSHA1Hash(c.UIN)
		if utils.StringifySHA1(p.Hash) == passHash {
			c.Authenticated = true
			c.Status = p.Status

			c.logger.Debugf("Sending GG_LOGIN_OK")
			c.SendLoginOK()

			// Set user's status
			cache.SetUserStatus(structs.StatusChangeMsg{
				UIN:    c.UIN,
				Status: p.Status,
			})

			c.Version, c.VOIP = utils.GetVersionAndVOIP(p.Version)

			return true
		} else {
			c.logger.Debugf("Sending GG_LOGIN_FAILED")
			c.SendLoginFail()
			return false
		}
		return false
	default:
		c.logger.Errorf("HandleLogin received unknown packetType: 0x%x", packetType)
		return false
	}
}

func (c *Client) HandleNotify30(pRecv *utils.IOStream) {
	p := protocol.GG_Notify30{}
	p.Deserialize(pRecv)
	log.StructPPrint("GG_NOTIFY30", p.PrettyPrint())
	for _, uin := range p.UINs {
		contact := structs.GG_NotifyContact{
			UIN:  uin,
			Type: 0x03,
		}
		c.NotifyList = append(c.NotifyList, contact)
	}
}

func (c *Client) HandleNotifyFirst(pRecv *utils.IOStream) {
	structs.GG_NotifyContactDeserialize(pRecv, &c.NotifyList)
}

func (c *Client) HandleNotifyLast(pRecv *utils.IOStream) {
	structs.GG_NotifyContactDeserialize(pRecv, &c.NotifyList)

	var packet protocol.GG_Packet_Iface

	// Respond with GG_NOTIFY_REPLY
	for _, notifyContact := range c.NotifyList {
		statusChange := c.server.cache.FetchUserStatus(notifyContact.UIN)
		if c.Version >= 0x2a {
			notifyReply := protocol.GG_Notify_Reply77{
				UIN:         statusChange.UIN,
				Status:      uint8(statusChange.Status),
				Description: statusChange.Description,
			}
			log.StructPPrint("GG_NOTIFY_REPLY77", notifyReply.PrettyPrint())
			packet = &notifyReply
		} else {
			notifyReply := protocol.GG_Notify_Reply60{
				UIN:         statusChange.UIN,
				Status:      uint8(statusChange.Status),
				Description: statusChange.Description,
			}
			log.StructPPrint("GG_NOTIFY_REPLY60", notifyReply.PrettyPrint())
			packet = &notifyReply
		}
	}

	c.SendNotifyReply(packet)
}

func (c *Client) HandleAddNotify(pRecv *utils.IOStream) {
	contact := structs.GG_AddNotify(pRecv, &c.NotifyList)
	c.SendStatus(c.server.cache.FetchUserStatus(contact.UIN))
}

func (c *Client) HandleRemoveNotify(pRecv *utils.IOStream) {
	p := protocol.GG_Remove_Notify{}
	p.Deserialize(pRecv)

	// Look for the contact that matches
	for i, notify := range c.NotifyList {
		if notify.UIN == p.UIN {
			c.logger.Debugf("Removed UIN: %d", notify.UIN)
			c.NotifyList[i] = structs.GG_NotifyContact{}
			return // We don't need to look further
		}
	}
}

func (c *Client) HandleNewStatus(pRecv *utils.IOStream) {
	p := protocol.GG_New_Status{}
	p.Deserialize(pRecv)

	c.server.cache.SetUserStatus(structs.StatusChangeMsg{
		UIN:         c.UIN,
		Status:      p.Status,
		Description: p.Description,
	})

	c.logger.Debugf("New status: 0x00%x, Description: %s", p.Status, p.Description)
}

func (c *Client) HandleSendMsg(pRecv *utils.IOStream) {
	p := protocol.GG_Send_MSG{}
	p.Deserialize(pRecv)
	log.StructPPrint("GG_SEND_MSG", p.PrettyPrint())
	c.server.cache.PublishMessageChannel(p.Recipient, structs.Message{c.UIN, p.MsgClass, []byte(p.Content)})
}

func (c *Client) HandleUserlistReq(pRecv *utils.IOStream) {
	packetLength := pRecv.Len()
	p := protocol.GG_Userlist_Request{}
	p.Deserialize(pRecv)
	log.StructPPrint("GG_USERLIST_REQUEST", p.PrettyPrint())

	switch p.Type {
	case constants.GG_USERLIST_PUT, constants.GG_USERLIST_PUT_MORE:
		if packetLength == 1 {
			// Client sends 1-sized userlist on userlist delete
			err := c.server.db.DeleteUserList(c.UIN)
			if err != nil {
				c.logger.Errorf("Failed to delete userlist: %s", err)
				return
			}

			p := protocol.GG_Userlist_Reply{
				Type:    constants.GG_USERLIST_PUT_REPLY,
				Request: p.Request,
			}
			log.StructPPrint("GG_USERLIST_REPLY", p.PrettyPrint())
			pOut := protocol.InitGG_Packet(protocol.GG_USERLIST_REPLY, &p)
			_, err = pOut.Send(c.conn)
			if err != nil {
				c.logger.Errorf("Error: %s", err)
			}
		}
		c.UserListBuf = append(c.UserListBuf, string(p.Request))
		if packetLength == 2048 {
			// We've got a multipart list, we need to add it to the buf and wait until
			// the client sends the final GG_USERLIST_PUT_MORE request
			return
		}
		// The client has sent the final part of the request, we can now process this
		c.PutUserList()
	case constants.GG_USERLIST_GET:
		c.logger.Debugf("Fetching contact list for UIN %d", c.UIN)
		userList := c.server.db.GetUserList(c.UIN)
		var userListBuf string
		for _, user := range userList {
			userListBuf += user.Write() + "\r\n"
		}
		c.logger.Debugf("Generated userlist: %s", strconv.Quote(userListBuf))
		c.logger.Debugf("Sending userlist back to the client...")

		c.SendGetUserListResp(userListBuf)
	}
}

func (c *Client) HandlePubdirReq(pRecv *utils.IOStream) {
	p := protocol.GG_Pubdir50_Request{}
	p.Deserialize(pRecv)
	log.StructPPrint("GG_PUBDIR50_REQUEST", p.PrettyPrint())

	switch p.Type {
	case constants.GG_PUBDIR50_SEARCH:
		req := pubdir.PubdirEntry{}
		err := req.Read(p.Request)
		if err != nil {
			c.logger.Errorf("Failed to read pubdir entry: %s", err)
			return
		}
		c.logger.Debugf("Received pubdir query: %+v", req)

		resp, nextStart, err := c.server.db.SearchInPubdir(&req)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			c.logger.Errorf("Failed to search through the pubdir: %v", err)
			c.SendPubdirResp(
				constants.GG_PUBDIR50_ERROR,
				p.Seq,
				nil,
			)
			return
		}
		c.logger.Debugf("Pubdir lookup returned %d rows", len(resp))

		if len(resp) == 0 {
			c.SendPubdirResp(
				constants.GG_PUBDIR50_SEARCH_REPLY,
				p.Seq,
				nil,
			)
			return
		}

		var respBuilder bytes.Buffer

		writer := transform.NewWriter(&respBuilder, charmap.Windows1250.NewEncoder())
		defer writer.Close()

		respBuilder.Write(pubdir.PubdirWriteRange(resp)) // TODO: make that use stream directly?
		pubdir.WriteSingleParam(writer, "\x00nextstart", nextStart)

		c.SendPubdirResp(
			constants.GG_PUBDIR50_SEARCH_REPLY,
			p.Seq,
			respBuilder.Bytes(),
		)
	case constants.GG_PUBDIR50_READ:
		resp, err := c.server.db.GetPubdirDataByUin(c.UIN)
		if errors.Is(err, sql.ErrNoRows) {
			c.logger.Infof("Creating empty pubdir entry for UIN %d", c.UIN)
			resp = &pubdir.PubdirEntry{}
			c.server.db.WritePubdirData(c.UIN, resp)
		} else if err != nil {
			c.logger.Errorf("Failed to retreive pubdir data for UIN %d: %v", c.UIN, err)
			c.SendPubdirResp(
				constants.GG_PUBDIR50_ERROR,
				p.Seq,
				nil,
			)
			return
		}

		// Swap the gender
		switch resp.Gender {
		case 1:
			resp.Gender = 2
		case 2:
			resp.Gender = 1
		}

		c.SendPubdirResp(
			constants.GG_PUBDIR50_SEARCH,
			p.Seq,
			resp.Write(),
		)
	case constants.GG_PUBDIR50_WRITE:
		req := pubdir.PubdirEntry{}
		err := req.Read(p.Request)
		if err != nil {
			c.logger.Errorf("Failed to read pubdir entry: %s", err)
			return
		}
		c.logger.Debugf("Received pubdir entry: %+v", req)

		// Swap the gender
		switch req.Gender {
		case 1:
			req.Gender = 2
		case 2:
			req.Gender = 1
		}

		err = c.server.db.WritePubdirData(c.UIN, &req)
		if err != nil {
			c.logger.Errorf("Failed to update pubdir data for %d: %v", c.UIN, err)
			c.SendPubdirResp(
				constants.GG_PUBDIR50_ERROR,
				p.Seq,
				nil,
			)
			return
		}

		// Acknowledge that the server received the data.
		// The fact that the correct type is 0x01 was not documented by libgadu.
		// Not sure if the data loopback is required though.
		// For future reference, the function that shows the correct message is FUN_00429cc1 in GG 6.1
		c.SendPubdirResp(
			constants.GG_PUBDIR50_WRITE,
			p.Seq,
			req.Write(),
		)
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
	log.StructPPrint("GG_PUBDIR50_REPLY", p.PrettyPrint())
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
	log.StructPPrint("GG_USERLIST_REPLY", p.PrettyPrint())
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
		log.StructPPrint("GG_USERLIST_REPLY", p.PrettyPrint())
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
	log.StructPPrint("GG_STATUS", p.PrettyPrint())
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
	log.StructPPrint("GG_STATUS60", p.PrettyPrint())
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
	log.StructPPrint("GG_STATUS77", p.PrettyPrint())
	pOut := protocol.InitGG_Packet(protocol.GG_STATUS77, &p)
	_, err := pOut.Send(c.conn)
	if err != nil {
		c.logger.Errorf("Error: %s", err)
	}
}

func (c *Client) SendRecvMsg(msg structs.Message) {
	pS := protocol.GG_Recv_MSG{
		Sender:   msg.From,
		Seq:      0,
		Time:     uint32(time.Now().Unix()),
		MsgClass: msg.MsgClass,
		Content:  string(msg.Content),
	}
	log.StructPPrint("GG_RECV_MSG", pS.PrettyPrint())
	pOut := protocol.InitGG_Packet(protocol.GG_RECV_MSG, &pS)
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

// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2025 Oreeeee

package clients

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"poggadaj-shared/cache"
	log "poggadaj-shared/logging"
	"poggadaj-shared/statuses"
	sharedstructs "poggadaj-shared/structs"
	"poggadaj-tcp/constants"
	db "poggadaj-tcp/database"
	"poggadaj-tcp/protocol/packets"
	"poggadaj-tcp/protocol/packets/c2s"
	"poggadaj-tcp/protocol/packets/s2c"
	"poggadaj-tcp/structs"
	uv "poggadaj-tcp/universal"
	"poggadaj-tcp/utils"
	"strconv"
	"strings"
	"time"
)

type GGClient struct {
	Conn          net.Conn
	UIN           uint32
	Status        uint32
	Authenticated bool
	NotifyList    []uv.GG_NotifyContact
	Version       uint8
	VOIP          bool
	ProtocolLevel uint8
	UserListBuf   []string
}

func (c *GGClient) HandleLogin(packetType uint32, pRecv packets.GG_Packet) bool {
	switch packetType {
	case c2s.GG_LOGIN30:
		c.ProtocolLevel = 30
		p := c2s.GG_Login30{}
		p.Deserialize(pRecv.Data)
		log.StructPPrint("GG_LOGIN30", p.PrettyPrint())

		c.UIN = p.UIN

		log.L.Debugf("Sending login response")
		passHash, _ := db.GetAncientHash(c.UIN)
		if p.Hash == passHash {
			c.Authenticated = true
			c.Status = p.Status

			log.L.Debugf("Sending GG_LOGIN_OK")
			c.SendLoginOK()

			cache.SetUserStatus(sharedstructs.StatusChangeMsg{
				UIN:    c.UIN,
				Status: p.Status,
			})

			return true
		}
		return false
	case c2s.GG_LOGIN:
		p := c2s.GG_Login{}
		p.Deserialize(pRecv.Data)
		log.StructPPrint("GG_LOGIN", p.PrettyPrint())

		c.UIN = p.UIN

		log.L.Debugf("Sending login response")
		passHash, _ := db.GetGG32Hash(c.UIN)
		if p.Hash == passHash {
			c.Authenticated = true
			c.Status = p.Status

			log.L.Debugf("Sending GG_LOGIN_OK")
			c.SendLoginOK()

			// Set user's status
			cache.SetUserStatus(sharedstructs.StatusChangeMsg{
				UIN:    c.UIN,
				Status: p.Status,
			})

			c.Version, c.VOIP = utils.GetVersionAndVOIP(p.Version)

			return true
		} else {
			log.L.Debugf("Sending GG_LOGIN_FAILED")
			c.SendLoginFail()
			return false
		}
		return false
	case c2s.GG_LOGIN60:
		p := c2s.GG_Login60{}
		p.Deserialize(pRecv.Data)
		log.StructPPrint("GG_LOGIN60", p.PrettyPrint())

		c.UIN = p.UIN

		log.L.Debugf("Sending login response")
		passHash, _ := db.GetGG32Hash(c.UIN)
		if p.Hash == passHash {
			c.Authenticated = true
			c.Status = p.Status

			log.L.Debugf("Sending GG_LOGIN_OK")
			c.SendLoginOK()

			// Set user's status
			cache.SetUserStatus(sharedstructs.StatusChangeMsg{
				UIN:    c.UIN,
				Status: p.Status,
			})

			c.Version, c.VOIP = utils.GetVersionAndVOIP(p.Version)

			return true
		} else {
			log.L.Debugf("Sending GG_LOGIN_FAILED")
			c.SendLoginFail()
			return false
		}
		return false
	case c2s.GG_LOGIN70:
		p := c2s.GG_Login70{}
		p.Deserialize(pRecv.Data)
		log.StructPPrint("GG_LOGIN70", p.PrettyPrint())

		c.UIN = p.UIN

		log.L.Debugf("Sending login response")
		passHash, _ := db.GetSHA1Hash(c.UIN)
		if utils.StringifySHA1(p.Hash) == passHash {
			c.Authenticated = true
			c.Status = p.Status

			log.L.Debugf("Sending GG_LOGIN_OK")
			c.SendLoginOK()

			// Set user's status
			cache.SetUserStatus(sharedstructs.StatusChangeMsg{
				UIN:    c.UIN,
				Status: p.Status,
			})

			c.Version, c.VOIP = utils.GetVersionAndVOIP(p.Version)

			return true
		} else {
			log.L.Debugf("Sending GG_LOGIN_FAILED")
			c.SendLoginFail()
			return false
		}
		return false
	default:
		log.L.Errorf("HandleLogin received unknown packetType: 0x%x", packetType)
		return false
	}
}

func (c *GGClient) HandleNotify30(pRecv packets.GG_Packet) {
	p := c2s.GG_Notify30{}
	p.Deserialize(pRecv.Data, pRecv.Length)
	log.StructPPrint("GG_NOTIFY30", p.PrettyPrint())
	for _, uin := range p.UINs {
		contact := uv.GG_NotifyContact{
			UIN:  uin,
			Type: 0x03,
		}
		c.NotifyList = append(c.NotifyList, contact)
	}
}

func (c *GGClient) HandleNotifyFirst(pRecv packets.GG_Packet) {
	uv.GG_NotifyContactDeserialize(pRecv.Data, pRecv.Length, &c.NotifyList)
}

func (c *GGClient) HandleNotifyLast(pRecv packets.GG_Packet) {
	uv.GG_NotifyContactDeserialize(pRecv.Data, pRecv.Length, &c.NotifyList)

	// Respond with GG_NOTIFY_REPLY
	response := make([]byte, 0)
	buf := bytes.NewBuffer(response)
	for _, notifyContact := range c.NotifyList {
		statusChange := cache.FetchUserStatus(notifyContact.UIN)
		if c.Version >= 0x2a {
			notifyReply := s2c.GG_Notify_Reply77{
				UIN:         statusChange.UIN,
				Status:      uint8(statusChange.Status),
				Description: statusChange.Description,
			}
			log.StructPPrint("GG_NOTIFY_REPLY77", notifyReply.PrettyPrint())
			binary.Write(buf, binary.LittleEndian, notifyReply.Serialize())
		} else {
			notifyReply := s2c.GG_Notify_Reply60{
				UIN:         statusChange.UIN,
				Status:      uint8(statusChange.Status),
				Description: statusChange.Description,
			}
			log.StructPPrint("GG_NOTIFY_REPLY60", notifyReply.PrettyPrint())
			binary.Write(buf, binary.LittleEndian, notifyReply.Serialize())
		}
	}

	c.SendNotifyReply(buf.Bytes())
}

func (c *GGClient) HandleAddNotify(pRecv packets.GG_Packet) {
	contact := uv.GG_AddNotify(pRecv.Data, &c.NotifyList)
	c.SendStatus(cache.FetchUserStatus(contact.UIN))
}

func (c *GGClient) HandleRemoveNotify(pRecv packets.GG_Packet) {
	p := c2s.GG_Remove_Notify{}
	p.Deserialize(pRecv.Data)

	// Look for the contact that matches
	for i, notify := range c.NotifyList {
		if notify.UIN == p.UIN {
			log.L.Debugf("Removed UIN: %d", notify.UIN)
			c.NotifyList[i] = uv.GG_NotifyContact{}
			return // We don't need to look further
		}
	}
}

func (c *GGClient) HandleNewStatus(pRecv packets.GG_Packet) {
	p := c2s.GG_New_Status{}
	p.Deserialize(pRecv.Data, pRecv.Length)

	cache.SetUserStatus(sharedstructs.StatusChangeMsg{
		UIN:         c.UIN,
		Status:      p.Status,
		Description: p.Description,
	})

	log.L.Debugf("New status: 0x00%x, Description: %s", p.Status, p.Description)
}

func (c *GGClient) HandleSendMsg(pRecv packets.GG_Packet) {
	p := c2s.GG_Send_MSG{}
	p.Deserialize(pRecv.Data, pRecv.Length)
	log.StructPPrint("GG_SEND_MSG", p.PrettyPrint())
	cache.PublishMessageChannel(p.Recipient, sharedstructs.Message{c.UIN, p.MsgClass, p.Content})
}

func (c *GGClient) HandleUserlistReq(pRecv packets.GG_Packet) {
	p := c2s.GG_Userlist_Request{}
	p.Deserialize(pRecv.Data, pRecv.Length)
	log.StructPPrint("GG_USERLIST_REQUEST", p.PrettyPrint())

	switch p.Type {
	case constants.GG_USERLIST_PUT, constants.GG_USERLIST_PUT_MORE:
		if pRecv.Length == 1 {
			// Client sends 1-sized userlist on userlist delete
			err := db.DeleteUserList(c.UIN)
			if err != nil {
				log.L.Errorf("Failed to delete userlist: %s", err)
				return
			}

			p := s2c.GG_Userlist_Reply{
				Type:    constants.GG_USERLIST_PUT_REPLY,
				Request: p.Request,
			}
			log.StructPPrint("GG_USERLIST_REPLY", p.PrettyPrint())
			pOut := packets.InitGG_Packet(s2c.GG_USERLIST_REPLY, p.Serialize())
			_, err = pOut.Send(c.Conn)
			if err != nil {
				log.L.Errorf("Error: %s", err)
			}
		}
		c.UserListBuf = append(c.UserListBuf, string(p.Request))
		if pRecv.Length == 2048 {
			// We've got a multipart list, we need to add it to the buf and wait until
			// the client sends the final GG_USERLIST_PUT_MORE request
			return
		}
		// The client has sent the final part of the request, we can now process this
		c.PutUserList()
	case constants.GG_USERLIST_GET:
		log.L.Debugf("Fetching contact list for UIN %d", c.UIN)
		userList := db.GetUserList(c.UIN)
		var userListBuf string
		for _, user := range userList {
			userListBuf += user.Write() + "\r\n"
		}
		log.L.Debugf("Generated userlist: %s", strconv.Quote(userListBuf))
		log.L.Debugf("Sending userlist back to the client...")

		c.SendGetUserListResp(userListBuf)
	}
}

func (c *GGClient) PutUserList() {
	err := db.DeleteUserList(c.UIN) // Delete user's contact list, as we are writing to the list and not appending to it
	if err != nil {
		log.L.Errorf("Failed to delete user list: %s", err)
	}

	userListStr := strings.Join(c.UserListBuf, "")                   // Combine the buffer into one string
	userListSeparated := strings.Split(userListStr, "\r\n")          // Separate the request lines
	userListSeparated = userListSeparated[:len(userListSeparated)-1] // Remove the last (empty) index

	// Convert all the strings to UserListRequest objects
	var userlist []structs.UserListRequest
	for _, str := range userListSeparated {
		log.L.Debugf("Read userlist: %s", strconv.Quote(str))
		user := structs.UserListRequest{}
		err := user.Read(str)
		if err != nil {
			log.L.Errorf("Error parsing userlist line: %v", err)
		}
		userlist = append(userlist, user)
	}

	log.L.Debugf("Received userlist put: %v", userlist)
	log.L.Debugf("Putting userlist into the database")
	db.PutUserList(userlist, c.UIN)

	// Send acknowledgement that the server received the list
	for i, _ := range c.UserListBuf {
		c.SendPutUserListAck(i)
	}

	// Clear the buffer
	c.UserListBuf = []string{}
}

func (c *GGClient) SendPutUserListAck(i int) {
	var reqType uint8
	if i == 0 {
		reqType = constants.GG_USERLIST_PUT_REPLY
	} else {
		reqType = constants.GG_USERLIST_PUT_MORE_REPLY
	}

	p := s2c.GG_Userlist_Reply{
		Type:    reqType,
		Request: []byte(c.UserListBuf[i]),
	}
	log.StructPPrint("GG_USERLIST_REPLY", p.PrettyPrint())
	pOut := packets.InitGG_Packet(s2c.GG_USERLIST_REPLY, p.Serialize())
	_, err := pOut.Send(c.Conn)
	if err != nil {
		log.L.Errorf("Error: %s", err)
	}
}

func (c *GGClient) SendGetUserListResp(userListBuf string) {
	chunkedList := utils.ChunkString(userListBuf, 2048)
	lastIndex := len(chunkedList) - 1
	for i, str := range chunkedList {
		replyType := constants.GG_USERLIST_GET_MORE_REPLY
		if i == lastIndex {
			// The last part of a list import is type GG_USERLIST_GET_REPLY
			replyType = constants.GG_USERLIST_GET_REPLY
		}
		p := s2c.GG_Userlist_Reply{
			Type:    uint8(replyType),
			Request: []byte(str),
		}
		log.StructPPrint("GG_USERLIST_REPLY", p.PrettyPrint())
		pOut := packets.InitGG_Packet(s2c.GG_USERLIST_REPLY, p.Serialize())
		_, err := pOut.Send(c.Conn)
		if err != nil {
			log.L.Errorf("Error: %s", err)
		}
	}
}

func (c *GGClient) SendLoginOK() {
	pOut := packets.InitGG_Packet(s2c.GG_LOGIN_OK, []byte{})
	_, err := pOut.Send(c.Conn)
	if err != nil {
		fmt.Println("Error: ", err)
	}
}

func (c *GGClient) SendLoginFail() {
	pOut := packets.InitGG_Packet(s2c.GG_LOGIN_FAILED, []byte{})
	_, err := pOut.Send(c.Conn)
	if err != nil {
		fmt.Println("Error: ", err)
	}
}

func (c *GGClient) SendStatus(statusChange sharedstructs.StatusChangeMsg) {
	if c.Version >= 0x2a {
		c.SendStatus77(statusChange)
	} else if c.Version >= 0x20 {
		c.SendStatus60(statusChange)
	} else if c.Version <= 0x18 {
		c.SendStatus50(statusChange)
	}
}

func (c *GGClient) SendStatus50(statusChange sharedstructs.StatusChangeMsg) {
	p := s2c.GG_Status{
		UIN:         statusChange.UIN,
		Status:      statusChange.Status,
		Description: statusChange.Description,
	}
	log.StructPPrint("GG_STATUS", p.PrettyPrint())
	pOut := packets.InitGG_Packet(s2c.GG_STATUS, p.Serialize())
	_, err := pOut.Send(c.Conn)
	if err != nil {
		log.L.Errorf("Error: %s", err)
	}
}

func (c *GGClient) SendStatus60(statusChange sharedstructs.StatusChangeMsg) {
	p := s2c.GG_Status60{
		UIN:         statusChange.UIN,
		Status:      uint8(statusChange.Status),
		RemoteIP:    0,
		RemotePort:  0,
		Version:     0,
		ImageSize:   0,
		Description: statusChange.Description,
	}
	log.StructPPrint("GG_STATUS60", p.PrettyPrint())
	pOut := packets.InitGG_Packet(s2c.GG_STATUS60, p.Serialize())
	_, err := pOut.Send(c.Conn)
	if err != nil {
		log.L.Errorf("Error: %s", err)
	}
}

func (c *GGClient) SendStatus77(statusChange sharedstructs.StatusChangeMsg) {
	p := s2c.GG_Status77{
		UIN:         statusChange.UIN,
		Status:      uint8(statusChange.Status),
		RemoteIP:    0,
		RemotePort:  0,
		Version:     0,
		ImageSize:   0,
		Description: statusChange.Description,
	}
	log.StructPPrint("GG_STATUS77", p.PrettyPrint())
	pOut := packets.InitGG_Packet(s2c.GG_STATUS77, p.Serialize())
	_, err := pOut.Send(c.Conn)
	if err != nil {
		log.L.Errorf("Error: %s", err)
	}
}

func (c *GGClient) SendRecvMsg(msg sharedstructs.Message) {
	pS := s2c.GG_Recv_MSG{
		Sender:   msg.From,
		Seq:      0,
		Time:     uint32(time.Now().Unix()),
		MsgClass: msg.MsgClass,
		Content:  msg.Content,
	}
	log.StructPPrint("GG_RECV_MSG", pS.PrettyPrint())
	pOut := packets.InitGG_Packet(s2c.GG_RECV_MSG, pS.Serialize())
	_, err := pOut.Send(c.Conn)
	if err != nil {
		log.L.Errorf("Error: %s", err)
	}
}

func (c *GGClient) SendNotifyReply(data []byte) {
	var pOut *packets.GG_Packet
	if c.Version >= 0x2a {
		pOut = packets.InitGG_Packet(s2c.GG_NOTIFY_REPLY77, data)
	} else {
		pOut = packets.InitGG_Packet(s2c.GG_NOTIFY_REPLY60, data)
	}
	_, err := pOut.Send(c.Conn)
	if err != nil {
		log.L.Debugf("Error: %s", err)
	}
}

func (c *GGClient) SendPong() {
	pOut := packets.InitGG_Packet(s2c.GG_PONG, []byte{})
	_, err := pOut.Send(c.Conn)
	if err != nil {
		log.L.Errorf("Error: %s", err)
	}
}

func (c *GGClient) Clean() {
	// Change user's status to not available
	cache.SetUserStatus(sharedstructs.StatusChangeMsg{
		UIN:    c.UIN,
		Status: statuses.GG_STATUS_NOT_AVAIL,
	})
}

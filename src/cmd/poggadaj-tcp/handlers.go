// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package main

import (
	"bytes"
	"database/sql"
	"errors"
	"strconv"

	"codeberg.org/or3e/poggadaj/cmd/poggadaj-tcp/protocol"
	"codeberg.org/or3e/poggadaj/internal/constants"
	"codeberg.org/or3e/poggadaj/internal/logging"
	"codeberg.org/or3e/poggadaj/internal/structs"
	"codeberg.org/or3e/poggadaj/internal/utils"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

// Handlers contains packet handlers that only authenticated users can use
var Handlers = map[uint32]func(*Client, *utils.IOStream) error{}

// LoginHandlers contains packet handlers relating to login packets that can only be used by unauthenticated users
var LoginHandlers = map[uint32]func(*Client, *utils.IOStream) error{}

func handleNotify30(c *Client, pRecv *utils.IOStream) error {
	p := protocol.GG_Notify30{}
	p.Deserialize(pRecv)
	logging.StructPPrint(c.logger, "GG_NOTIFY30", p.PrettyPrint())
	for _, uin := range p.UINs {
		contact := structs.GG_NotifyContact{
			UIN:  uin,
			Type: 0x03,
		}
		c.NotifyList = append(c.NotifyList, contact)
	}
	return nil
}

func handleNotifyFirst(c *Client, pRecv *utils.IOStream) error {
	structs.GG_NotifyContactDeserialize(pRecv, &c.NotifyList)
	return nil
}

func handleNotifyLast(c *Client, pRecv *utils.IOStream) error {
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
			logging.StructPPrint(c.logger, "GG_NOTIFY_REPLY77", notifyReply.PrettyPrint())
			packet = &notifyReply
		} else {
			notifyReply := protocol.GG_Notify_Reply60{
				UIN:         statusChange.UIN,
				Status:      uint8(statusChange.Status),
				Description: statusChange.Description,
			}
			logging.StructPPrint(c.logger, "GG_NOTIFY_REPLY60", notifyReply.PrettyPrint())
			packet = &notifyReply
		}
	}

	c.SendNotifyReply(packet)
	return nil
}

func handleAddNotify(c *Client, pRecv *utils.IOStream) error {
	contact := structs.GG_AddNotify(pRecv, &c.NotifyList)
	c.SendStatus(c.server.cache.FetchUserStatus(contact.UIN))
	return nil
}

func handleRemoveNotify(c *Client, pRecv *utils.IOStream) error {
	p := protocol.GG_Remove_Notify{}
	p.Deserialize(pRecv)

	// Look for the contact that matches
	for i, notify := range c.NotifyList {
		if notify.UIN == p.UIN {
			c.logger.Debugf("Removed UIN: %d", notify.UIN)
			c.NotifyList[i] = structs.GG_NotifyContact{}
			return nil // We don't need to look further
		}
	}
	return nil
}

func handleNotifyEmpty(c *Client, stream *utils.IOStream) error {
	return nil
}

func handleNewStatus(c *Client, pRecv *utils.IOStream) error {
	p := protocol.GG_New_Status{}
	p.Deserialize(pRecv)

	c.server.cache.SetUserStatus(structs.StatusChangeMsg{
		UIN:         c.UIN,
		Status:      p.Status,
		Description: p.Description,
	})

	c.logger.Debugf("New status: 0x00%x, Description: %s", p.Status, p.Description)
	return nil
}

func handleSendMsg(c *Client, pRecv *utils.IOStream) error {
	p := protocol.GG_Send_MSG{}
	p.Deserialize(pRecv)
	logging.StructPPrint(c.logger, "GG_SEND_MSG", p.PrettyPrint())
	c.server.cache.PublishMessageChannel(p.Recipient, structs.Message{c.UIN, p.MsgClass, []byte(p.Content)})
	return nil
}

func handleUserlistReq(c *Client, pRecv *utils.IOStream) error {
	packetLength := pRecv.Len()
	p := protocol.GG_Userlist_Request{}
	p.Deserialize(pRecv)
	logging.StructPPrint(c.logger, "GG_USERLIST_REQUEST", p.PrettyPrint())

	switch p.Type {
	case constants.GG_USERLIST_PUT, constants.GG_USERLIST_PUT_MORE:
		if packetLength == 1 {
			// Client sends 1-sized userlist on userlist delete
			err := c.server.db.DeleteUserList(c.UIN)
			if err != nil {
				c.logger.Errorf("Failed to delete userlist: %s", err)
				return err
			}

			p := protocol.GG_Userlist_Reply{
				Type:    constants.GG_USERLIST_PUT_REPLY,
				Request: p.Request,
			}
			logging.StructPPrint(c.logger, "GG_USERLIST_REPLY", p.PrettyPrint())
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
			return nil
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
	return nil
}

func handlePubdirReq(c *Client, pRecv *utils.IOStream) error {
	p := protocol.GG_Pubdir50_Request{}
	p.Deserialize(pRecv)
	logging.StructPPrint(c.logger, "GG_PUBDIR50_REQUEST", p.PrettyPrint())

	switch p.Type {
	case constants.GG_PUBDIR50_SEARCH:
		req := structs.PubdirEntry{}
		err := req.Read(p.Request)
		if err != nil {
			c.logger.Errorf("Failed to read pubdir entry: %s", err)
			return err
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
			return err
		}
		c.logger.Debugf("Pubdir lookup returned %d rows", len(resp))

		if len(resp) == 0 {
			c.SendPubdirResp(
				constants.GG_PUBDIR50_SEARCH_REPLY,
				p.Seq,
				nil,
			)
			return err
		}

		var respBuilder bytes.Buffer

		writer := transform.NewWriter(&respBuilder, charmap.Windows1250.NewEncoder())
		defer writer.Close()

		respBuilder.Write(structs.PubdirWriteRange(resp)) // TODO: make that use stream directly?
		structs.WriteSingleParam(writer, "\x00nextstart", nextStart)

		c.SendPubdirResp(
			constants.GG_PUBDIR50_SEARCH_REPLY,
			p.Seq,
			respBuilder.Bytes(),
		)
	case constants.GG_PUBDIR50_READ:
		resp, err := c.server.db.GetPubdirDataByUin(c.UIN)
		if errors.Is(err, sql.ErrNoRows) {
			c.logger.Infof("Creating empty pubdir entry for UIN %d", c.UIN)
			resp = &structs.PubdirEntry{}
			c.server.db.WritePubdirData(c.UIN, resp)
		} else if err != nil {
			c.logger.Errorf("Failed to retreive pubdir data for UIN %d: %v", c.UIN, err)
			c.SendPubdirResp(
				constants.GG_PUBDIR50_ERROR,
				p.Seq,
				nil,
			)
			return err
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
		req := structs.PubdirEntry{}
		err := req.Read(p.Request)
		if err != nil {
			c.logger.Errorf("Failed to read pubdir entry: %s", err)
			return err
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
			return err
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
	return nil
}

func handlePing(c *Client, stream *utils.IOStream) error {
	c.SendPong()
	return nil
}

func handleLogin30(c *Client, stream *utils.IOStream) error {
	c.ProtocolLevel = 30
	p := protocol.GG_Login30{}
	p.Deserialize(stream)
	logging.StructPPrint(c.logger, "GG_LOGIN30", p.PrettyPrint())

	c.UIN = p.UIN

	passHash, _ := c.server.db.GetAncientHash(c.UIN)
	if p.Hash == passHash {
		c.Status = p.Status
		return nil
	}

	return errors.New("not authenticated")
}

func handleLogin(c *Client, stream *utils.IOStream) error {
	p := protocol.GG_Login{}
	p.Deserialize(stream)
	logging.StructPPrint(c.logger, "GG_LOGIN", p.PrettyPrint())

	c.UIN = p.UIN

	c.logger.Debugf("Sending login response")
	passHash, _ := c.server.db.GetGG32Hash(c.UIN)
	if p.Hash == passHash {
		c.Status = p.Status
		c.Version, c.VOIP = utils.GetVersionAndVOIP(p.Version)
		return nil
	}
	return errors.New("not authenticated")
}

func handleLogin60(c *Client, stream *utils.IOStream) error {
	p := protocol.GG_Login60{}
	p.Deserialize(stream)
	logging.StructPPrint(c.logger, "GG_LOGIN60", p.PrettyPrint())

	c.UIN = p.UIN

	c.logger.Debugf("Sending login response")
	passHash, _ := c.server.db.GetGG32Hash(c.UIN)
	if p.Hash == passHash {
		c.Status = p.Status
		c.Version, c.VOIP = utils.GetVersionAndVOIP(p.Version)
		return nil
	}
	return errors.New("not authenticated")
}

func handleLogin70(c *Client, stream *utils.IOStream) error {
	p := protocol.GG_Login70{}
	p.Deserialize(stream)
	logging.StructPPrint(c.logger, "GG_LOGIN70", p.PrettyPrint())

	c.UIN = p.UIN

	c.logger.Debugf("Sending login response")
	passHash, _ := c.server.db.GetSHA1Hash(c.UIN)
	if utils.StringifySHA1(p.Hash) == passHash {
		c.Status = p.Status
		c.Version, c.VOIP = utils.GetVersionAndVOIP(p.Version)

		return nil
	}
	return errors.New("not authenticated")
}

func init() {
	Handlers[protocol.GG_NOTIFY30] = handleNotify30
	Handlers[protocol.GG_NOTIFY_FIRST] = handleNotifyFirst
	Handlers[protocol.GG_NOTIFY_LAST] = handleNotifyLast
	Handlers[protocol.GG_ADD_NOTIFY] = handleAddNotify
	Handlers[protocol.GG_REMOVE_NOTIFY] = handleRemoveNotify
	Handlers[protocol.GG_LIST_EMPTY] = handleNotifyLast
	Handlers[protocol.GG_NEW_STATUS] = handleNewStatus
	Handlers[protocol.GG_SEND_MSG] = handleSendMsg
	Handlers[protocol.GG_USERLIST_REQUEST] = handleUserlistReq
	Handlers[protocol.GG_PUBDIR50_REQUEST] = handlePubdirReq
	Handlers[protocol.GG_PING] = handlePing

	LoginHandlers[protocol.GG_LOGIN30] = handleLogin30
	LoginHandlers[protocol.GG_LOGIN] = handleLogin
	LoginHandlers[protocol.GG_LOGIN60] = handleLogin60
	LoginHandlers[protocol.GG_LOGIN70] = handleLogin70
}

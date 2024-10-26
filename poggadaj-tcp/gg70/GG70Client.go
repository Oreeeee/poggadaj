package gg70

import (
	"bytes"
	"encoding/binary"
	"fmt"
	db "poggadaj-tcp/database"
	"poggadaj-tcp/gg60"
	log "poggadaj-tcp/logging"
	"poggadaj-tcp/structs"
	uv "poggadaj-tcp/universal"
	"poggadaj-tcp/utils"
	"time"
)

type GG70Client struct {
	cI structs.ClientInfo
}

func (c *GG70Client) HandleLogin(pRecv uv.GG_Packet) bool {
	p := GG_Login70{}
	p.Deserialize(pRecv.Data)
	log.StructPPrint("GG_LOGIN70", p.PrettyPrint())

	c.cI.UIN = p.UIN
	log.L.Debugf("Sending login response")
	passHash, _ := db.GetSHA1Hash(c.cI.UIN)
	if utils.StringifySHA1(p.Hash) == passHash {
		c.cI.Authenticated = true
		c.cI.Status = p.Status

		log.L.Debugf("Sending GG_LOGIN_OK")
		c.SendLoginOK()

		// Set user's status
		db.SetUserStatus(uv.StatusChangeMsg{
			UIN:    c.cI.UIN,
			Status: p.Status,
		})

		return true
	} else {
		log.L.Debugf("Sending GG_LOGIN_FAILED")
		c.SendLoginFail()
		return false
	}
	return false
}

func (c *GG70Client) HandleNotifyFirst(pRecv uv.GG_Packet) {
	uv.GG_NotifyContactDeserialize(pRecv.Data, pRecv.Length, &c.cI.NotifyList)
}

func (c *GG70Client) HandleNotifyLast(pRecv uv.GG_Packet) {
	uv.GG_NotifyContactDeserialize(pRecv.Data, pRecv.Length, &c.cI.NotifyList)
	//fmt.Println(currConn.NotifyList)

	// Respond with GG_NOTIFY_REPLY
	response := make([]byte, 0)
	buf := bytes.NewBuffer(response)
	for _, notifyContact := range c.cI.NotifyList {
		statusChange := db.FetchUserStatus(notifyContact.UIN)
		notifyReply := gg60.GG_Notify_Reply60{
			UIN:         statusChange.UIN,
			Status:      uint8(statusChange.Status),
			Description: statusChange.Description,
		}
		binary.Write(buf, binary.LittleEndian, notifyReply.Serialize())
	}

	c.SendNotifyReply(buf.Bytes())
}

func (c *GG70Client) HandleAddNotify(pRecv uv.GG_Packet) {
	contact := uv.GG_AddNotify(pRecv.Data, &c.cI.NotifyList)
	c.SendStatus(db.FetchUserStatus(contact.UIN))
}

func (c *GG70Client) HandleRemoveNotify(pRecv uv.GG_Packet) {
	p := uv.GG_Remove_Notify{}
	p.Deserialize(pRecv.Data)

	// Look for the contact that matches
	for i, notify := range c.cI.NotifyList {
		if notify.UIN == p.UIN {
			log.L.Debugf("Removed UIN: %d", notify.UIN)
			c.cI.NotifyList[i] = uv.GG_NotifyContact{}
			return // We don't need to look further
		}
	}
}

func (c *GG70Client) HandleNewStatus(pRecv uv.GG_Packet) {
	p := uv.GG_New_Status{}
	p.Deserialize(pRecv.Data, pRecv.Length)

	db.SetUserStatus(uv.StatusChangeMsg{
		UIN:         c.cI.UIN,
		Status:      p.Status,
		Description: p.Description,
	})

	log.L.Debugf("New status: 0x00%x, Description: %s", p.Status, p.Description)
}

func (c *GG70Client) HandleSendMsg(pRecv uv.GG_Packet) {
	p := uv.GG_Send_MSG{}
	p.Deserialize(pRecv.Data, pRecv.Length)
	db.PublishMessageChannel(p.Recipient, structs.Message{c.cI.UIN, p.Content})
}

// S2C
func (c *GG70Client) SendLoginOK() {
	pOut := uv.InitGG_Packet(uv.GG_LOGIN_OK, []byte{})
	_, err := pOut.Send(c.cI.Conn)
	if err != nil {
		fmt.Println("Error: ", err)
	}
}

func (c *GG70Client) SendLoginFail() {
	pOut := uv.InitGG_Packet(uv.GG_LOGIN_FAILED, []byte{})
	_, err := pOut.Send(c.cI.Conn)
	if err != nil {
		fmt.Println("Error: ", err)
	}
}

func (c *GG70Client) SendStatus(statusChange uv.StatusChangeMsg) {
	p := gg60.GG_Status60{
		UIN:         statusChange.UIN,
		Status:      uint8(statusChange.Status),
		RemoteIP:    0,
		RemotePort:  0,
		Version:     0,
		ImageSize:   0,
		Unknown1:    0,
		Description: statusChange.Description,
	}
	pOut := uv.InitGG_Packet(uv.GG_STATUS60, p.Serialize())
	_, err := pOut.Send(c.cI.Conn)
	if err != nil {
		log.L.Errorf("Error: %s", err)
	}
}

func (c *GG70Client) SendRecvMsg(msg structs.Message) {
	pS := uv.GG_Recv_MSG{
		Sender:   msg.From,
		Seq:      0,
		Time:     uint32(time.Now().Unix()),
		MsgClass: 0x08,
		Content:  msg.Content,
	}
	pOut := uv.InitGG_Packet(uv.GG_RECV_MSG, pS.Serialize())
	_, err := pOut.Send(c.cI.Conn)
	if err != nil {
		log.L.Errorf("Error: %s", err)
	}
}

func (c *GG70Client) SendNotifyReply(data []byte) {
	pOut := uv.InitGG_Packet(uv.GG_NOTIFY_REPLY60, data)
	_, err := pOut.Send(c.cI.Conn)
	if err != nil {
		log.L.Debugf("Error: %s", err)
	}
}

func (c *GG70Client) SendPong() {
	pOut := uv.InitGG_Packet(uv.GG_PONG, []byte{})
	_, err := pOut.Send(c.cI.Conn)
	if err != nil {
		log.L.Errorf("Error: %s", err)
	}
}

// Non-protocol
func (c *GG70Client) GetClientInfoPtr() *structs.ClientInfo {
	return &c.cI
}

func (c *GG70Client) Clean() {
	// Change user's status to not available
	db.SetUserStatus(uv.StatusChangeMsg{
		UIN:    c.cI.UIN,
		Status: uv.GG_STATUS_NOT_AVAIL,
	})
}

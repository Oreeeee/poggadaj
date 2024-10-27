package clients

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	db "poggadaj-tcp/database"
	log "poggadaj-tcp/logging"
	"poggadaj-tcp/protocol/packets"
	"poggadaj-tcp/protocol/packets/c2s"
	"poggadaj-tcp/protocol/packets/s2c"
	"poggadaj-tcp/structs"
	uv "poggadaj-tcp/universal"
	"poggadaj-tcp/utils"
	"time"
)

type GGClient struct {
	Conn          net.Conn
	UIN           uint32
	Status        uint32
	Authenticated bool
	NotifyList    []uv.GG_NotifyContact
	Version       uint32
}

func (c *GGClient) HandleLogin(packetType uint32, pRecv packets.GG_Packet) bool {
	switch packetType {
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
			db.SetUserStatus(uv.StatusChangeMsg{
				UIN:    c.UIN,
				Status: p.Status,
			})

			c.Version = p.Version

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
			db.SetUserStatus(uv.StatusChangeMsg{
				UIN:    c.UIN,
				Status: p.Status,
			})

			c.Version = p.Version

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

func (c *GGClient) HandleNotifyFirst(pRecv packets.GG_Packet) {
	uv.GG_NotifyContactDeserialize(pRecv.Data, pRecv.Length, &c.NotifyList)
}

func (c *GGClient) HandleNotifyLast(pRecv packets.GG_Packet) {
	uv.GG_NotifyContactDeserialize(pRecv.Data, pRecv.Length, &c.NotifyList)

	// Respond with GG_NOTIFY_REPLY
	response := make([]byte, 0)
	buf := bytes.NewBuffer(response)
	for _, notifyContact := range c.NotifyList {
		statusChange := db.FetchUserStatus(notifyContact.UIN)
		notifyReply := s2c.GG_Notify_Reply60{
			UIN:         statusChange.UIN,
			Status:      uint8(statusChange.Status),
			Description: statusChange.Description,
		}
		binary.Write(buf, binary.LittleEndian, notifyReply.Serialize())
	}

	c.SendNotifyReply(buf.Bytes())
}

func (c *GGClient) HandleAddNotify(pRecv packets.GG_Packet) {
	contact := uv.GG_AddNotify(pRecv.Data, &c.NotifyList)
	c.SendStatus(db.FetchUserStatus(contact.UIN))
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
	p := uv.GG_New_Status{}
	p.Deserialize(pRecv.Data, pRecv.Length)

	db.SetUserStatus(uv.StatusChangeMsg{
		UIN:         c.UIN,
		Status:      p.Status,
		Description: p.Description,
	})

	log.L.Debugf("New status: 0x00%x, Description: %s", p.Status, p.Description)
}

func (c *GGClient) HandleSendMsg(pRecv packets.GG_Packet) {
	p := c2s.GG_Send_MSG{}
	p.Deserialize(pRecv.Data, pRecv.Length)
	db.PublishMessageChannel(p.Recipient, structs.Message{c.UIN, p.Content})
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

func (c *GGClient) SendStatus(statusChange uv.StatusChangeMsg) {
	p := s2c.GG_Status60{
		UIN:         statusChange.UIN,
		Status:      uint8(statusChange.Status),
		RemoteIP:    0,
		RemotePort:  0,
		Version:     0,
		ImageSize:   0,
		Unknown1:    0,
		Description: statusChange.Description,
	}
	pOut := packets.InitGG_Packet(s2c.GG_STATUS60, p.Serialize())
	_, err := pOut.Send(c.Conn)
	if err != nil {
		log.L.Errorf("Error: %s", err)
	}
}

func (c *GGClient) SendRecvMsg(msg structs.Message) {
	pS := s2c.GG_Recv_MSG{
		Sender:   msg.From,
		Seq:      0,
		Time:     uint32(time.Now().Unix()),
		MsgClass: 0x08,
		Content:  msg.Content,
	}
	pOut := packets.InitGG_Packet(s2c.GG_RECV_MSG, pS.Serialize())
	_, err := pOut.Send(c.Conn)
	if err != nil {
		log.L.Errorf("Error: %s", err)
	}
}

func (c *GGClient) SendNotifyReply(data []byte) {
	pOut := packets.InitGG_Packet(s2c.GG_NOTIFY_REPLY60, data)
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
	db.SetUserStatus(uv.StatusChangeMsg{
		UIN:    c.UIN,
		Status: uv.GG_STATUS_NOT_AVAIL,
	})
}
package gg70

import (
	"fmt"
	db "poggadaj-tcp/database"
	log "poggadaj-tcp/logging"
	"poggadaj-tcp/structs"
	uv "poggadaj-tcp/universal"
	"poggadaj-tcp/utils"
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

func (c *GG70Client) HandleNotifyFirst(pRecv uv.GG_Packet)  {}
func (c *GG70Client) HandleNotifyLast(pRecv uv.GG_Packet)   {}
func (c *GG70Client) HandleAddNotify(pRecv uv.GG_Packet)    {}
func (c *GG70Client) HandleRemoveNotify(pRecv uv.GG_Packet) {}
func (c *GG70Client) HandleNewStatus(pRecv uv.GG_Packet)    {}
func (c *GG70Client) HandleSendMsg(pRecv uv.GG_Packet)      {}

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

func (c *GG70Client) SendStatus(statusChange uv.StatusChangeMsg) {}
func (c *GG70Client) SendRecvMsg(msg structs.Message)            {}
func (c *GG70Client) SendNotifyReply(data []byte)                {}
func (c *GG70Client) SendPong()                                  {}

// Non-protocol
func (c *GG70Client) GetClientInfoPtr() *structs.ClientInfo {
	return &c.cI
}

func (c *GG70Client) Clean() {}

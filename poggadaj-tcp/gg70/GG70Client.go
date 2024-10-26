package gg70

import (
	log "poggadaj-tcp/logging"
	"poggadaj-tcp/structs"
	uv "poggadaj-tcp/universal"
)

type GG70Client struct {
	cI structs.ClientInfo
}

func (c *GG70Client) HandleLogin(pRecv uv.GG_Packet) bool {
	p := GG_Login70{}
	p.Deserialize(pRecv.Data)
	log.StructPPrint("GG_LOGIN70", p.PrettyPrint())

	c.cI.UIN = p.UIN
	return false
}

func (c *GG70Client) HandleNotifyFirst(pRecv uv.GG_Packet)  {}
func (c *GG70Client) HandleNotifyLast(pRecv uv.GG_Packet)   {}
func (c *GG70Client) HandleAddNotify(pRecv uv.GG_Packet)    {}
func (c *GG70Client) HandleRemoveNotify(pRecv uv.GG_Packet) {}
func (c *GG70Client) HandleNewStatus(pRecv uv.GG_Packet)    {}
func (c *GG70Client) HandleSendMsg(pRecv uv.GG_Packet)      {}

// S2C
func (c *GG70Client) SendLoginOK()                               {}
func (c *GG70Client) SendLoginFail()                             {}
func (c *GG70Client) SendStatus(statusChange uv.StatusChangeMsg) {}
func (c *GG70Client) SendRecvMsg(msg structs.Message)            {}
func (c *GG70Client) SendNotifyReply(data []byte)                {}
func (c *GG70Client) SendPong()                                  {}

// Non-protocol
func (c *GG70Client) GetClientInfoPtr() *structs.ClientInfo {
	return &c.cI
}

func (c *GG70Client) Clean() {}

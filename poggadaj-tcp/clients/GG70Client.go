package clients

import (
	db "poggadaj-tcp/database"
	gh "poggadaj-tcp/generichandlers"
	log "poggadaj-tcp/logging"
	"poggadaj-tcp/protocol/packets"
	"poggadaj-tcp/protocol/packets/c2s"
	"poggadaj-tcp/structs"
	uv "poggadaj-tcp/universal"
	"poggadaj-tcp/utils"
)

type GG70Client struct {
	cI structs.ClientInfo
}

func (c *GG70Client) HandleLogin(pRecv packets.GG_Packet) bool {
	p := c2s.GG_Login70{}
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

func (c *GG70Client) HandleNotifyFirst(pRecv packets.GG_Packet) {
	gh.HandleNotifyFirst(c, pRecv)
}

func (c *GG70Client) HandleNotifyLast(pRecv packets.GG_Packet) {
	gh.HandleNotifyLast(c, pRecv)
}

func (c *GG70Client) HandleAddNotify(pRecv packets.GG_Packet) {
	gh.HandleAddNotify(c, pRecv)
}

func (c *GG70Client) HandleRemoveNotify(pRecv packets.GG_Packet) {
	gh.HandleRemoveNotify(c, pRecv)
}

func (c *GG70Client) HandleNewStatus(pRecv packets.GG_Packet) {
	gh.HandleNewStatus(c, pRecv)
}

func (c *GG70Client) HandleSendMsg(pRecv packets.GG_Packet) {
	gh.HandleSendMsg(c, pRecv)
}

// S2C
func (c *GG70Client) SendLoginOK() {
	gh.SendLoginOK(c)
}

func (c *GG70Client) SendLoginFail() {
	gh.SendLoginFail(c)
}

func (c *GG70Client) SendStatus(statusChange uv.StatusChangeMsg) {
	gh.SendStatus(c, statusChange)
}

func (c *GG70Client) SendRecvMsg(msg structs.Message) {
	gh.SendRecvMsg(c, msg)
}

func (c *GG70Client) SendNotifyReply(data []byte) {
	gh.SendNotifyReply(c, data)
}

func (c *GG70Client) SendPong() {
	gh.SendPong(c)
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

package gg60

import (
	structs "poggadaj-tcp/structs"
)

type GG60Client struct {
	cI structs.ClientInfo
}

// C2S
func (c *GG60Client) HandleLogin()        {}
func (c *GG60Client) HandleNotifyFirst()  {}
func (c *GG60Client) HandleNotifyLast()   {}
func (c *GG60Client) HandleAddNotify()    {}
func (c *GG60Client) HandleRemoveNotify() {}
func (c *GG60Client) HandleNewStatus()    {}
func (c *GG60Client) HandleSendMsg()      {}

// S2C
func (c *GG60Client) SendLoginOK()   {}
func (c *GG60Client) SendLoginFail() {}
func (c *GG60Client) SendStatus()    {}
func (c *GG60Client) SendRecvMsg()   {}

// Non-protocol
func (c *GG60Client) GetClientInfoPtr() *structs.ClientInfo {
	return &c.cI
}

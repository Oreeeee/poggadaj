package main

import "poggadaj-tcp/structs"

type GGClient interface {
	// C2S
	HandleLogin()
	HandleNotifyFirst()
	HandleNotifyLast()
	HandleAddNotify()
	HandleRemoveNotify()
	HandleNewStatus()
	HandleSendMsg()

	// S2C
	SendLoginOK()
	SendLoginFail()
	SendStatus()
	SendRecvMsg()

	// Non-protocol
	GetClientInfoPtr() *structs.ClientInfo
}

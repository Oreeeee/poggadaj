package main

import (
	"poggadaj-tcp/structs"
	uv "poggadaj-tcp/universal"
)

type GGClient interface {
	// C2S
	HandleLogin(pRecv uv.GG_Packet) bool
	HandleNotifyFirst(pRecv uv.GG_Packet)
	HandleNotifyLast(pRecv uv.GG_Packet)
	HandleAddNotify(pRecv uv.GG_Packet)
	HandleRemoveNotify(pRecv uv.GG_Packet)
	HandleNewStatus(pRecv uv.GG_Packet)
	HandleSendMsg(pRecv uv.GG_Packet)

	// S2C
	SendLoginOK()
	SendLoginFail()
	SendStatus(statusChange uv.StatusChangeMsg)
	SendRecvMsg(msg structs.Message)
	SendNotifyReply(data []byte)
	SendPong()

	// Non-protocol
	GetClientInfoPtr() *structs.ClientInfo
	Clean()
}

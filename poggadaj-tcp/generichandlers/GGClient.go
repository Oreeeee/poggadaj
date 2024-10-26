package generichandlers

import (
	"poggadaj-tcp/structs"
	"poggadaj-tcp/universal"
)

type GGClient interface {
	// C2S
	HandleLogin(pRecv universal.GG_Packet) bool
	HandleNotifyFirst(pRecv universal.GG_Packet)
	HandleNotifyLast(pRecv universal.GG_Packet)
	HandleAddNotify(pRecv universal.GG_Packet)
	HandleRemoveNotify(pRecv universal.GG_Packet)
	HandleNewStatus(pRecv universal.GG_Packet)
	HandleSendMsg(pRecv universal.GG_Packet)

	// S2C
	SendLoginOK()
	SendLoginFail()
	SendStatus(statusChange universal.StatusChangeMsg)
	SendRecvMsg(msg structs.Message)
	SendNotifyReply(data []byte)
	SendPong()

	// Non-protocol
	GetClientInfoPtr() *structs.ClientInfo
	Clean()
}

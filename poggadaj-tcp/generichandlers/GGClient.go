package generichandlers

import (
	"poggadaj-tcp/protocol/packets"
	"poggadaj-tcp/structs"
	"poggadaj-tcp/universal"
)

type GGClient interface {
	// C2S
	HandleLogin(pRecv packets.GG_Packet) bool
	HandleNotifyFirst(pRecv packets.GG_Packet)
	HandleNotifyLast(pRecv packets.GG_Packet)
	HandleAddNotify(pRecv packets.GG_Packet)
	HandleRemoveNotify(pRecv packets.GG_Packet)
	HandleNewStatus(pRecv packets.GG_Packet)
	HandleSendMsg(pRecv packets.GG_Packet)

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

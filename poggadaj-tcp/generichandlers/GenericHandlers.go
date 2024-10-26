package generichandlers

import (
	"bytes"
	"encoding/binary"
	"fmt"
	db "poggadaj-tcp/database"
	"poggadaj-tcp/gg60"
	log "poggadaj-tcp/logging"
	"poggadaj-tcp/structs"
	"poggadaj-tcp/universal"
	"time"
)

// C2S
func HandleNotifyFirst(c GGClient, pRecv universal.GG_Packet) {
	cI := c.GetClientInfoPtr()
	universal.GG_NotifyContactDeserialize(pRecv.Data, pRecv.Length, &cI.NotifyList)
}

func HandleNotifyLast(c GGClient, pRecv universal.GG_Packet) {
	cI := c.GetClientInfoPtr()
	universal.GG_NotifyContactDeserialize(pRecv.Data, pRecv.Length, &cI.NotifyList)

	// Respond with GG_NOTIFY_REPLY
	response := make([]byte, 0)
	buf := bytes.NewBuffer(response)
	for _, notifyContact := range cI.NotifyList {
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

func HandleAddNotify(c GGClient, pRecv universal.GG_Packet) {
	cI := c.GetClientInfoPtr()
	contact := universal.GG_AddNotify(pRecv.Data, &cI.NotifyList)
	c.SendStatus(db.FetchUserStatus(contact.UIN))
}

func HandleRemoveNotify(c GGClient, pRecv universal.GG_Packet) {
	cI := c.GetClientInfoPtr()
	p := universal.GG_Remove_Notify{}
	p.Deserialize(pRecv.Data)

	// Look for the contact that matches
	for i, notify := range cI.NotifyList {
		if notify.UIN == p.UIN {
			log.L.Debugf("Removed UIN: %d", notify.UIN)
			cI.NotifyList[i] = universal.GG_NotifyContact{}
			return // We don't need to look further
		}
	}
}

func HandleNewStatus(c GGClient, pRecv universal.GG_Packet) {
	cI := c.GetClientInfoPtr()
	p := universal.GG_New_Status{}
	p.Deserialize(pRecv.Data, pRecv.Length)

	db.SetUserStatus(universal.StatusChangeMsg{
		UIN:         cI.UIN,
		Status:      p.Status,
		Description: p.Description,
	})

	log.L.Debugf("New status: 0x00%x, Description: %s", p.Status, p.Description)
}

func HandleSendMsg(c GGClient, pRecv universal.GG_Packet) {
	cI := c.GetClientInfoPtr()
	p := universal.GG_Send_MSG{}
	p.Deserialize(pRecv.Data, pRecv.Length)
	db.PublishMessageChannel(p.Recipient, structs.Message{cI.UIN, p.Content})
}

// S2C
func SendLoginOK(c GGClient) {
	cI := c.GetClientInfoPtr()
	pOut := universal.InitGG_Packet(universal.GG_LOGIN_OK, []byte{})
	_, err := pOut.Send(cI.Conn)
	if err != nil {
		fmt.Println("Error: ", err)
	}
}

func SendLoginFail(c GGClient) {
	cI := c.GetClientInfoPtr()
	pOut := universal.InitGG_Packet(universal.GG_LOGIN_FAILED, []byte{})
	_, err := pOut.Send(cI.Conn)
	if err != nil {
		fmt.Println("Error: ", err)
	}
}

func SendStatus(c GGClient, statusChange universal.StatusChangeMsg) {
	cI := c.GetClientInfoPtr()
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
	pOut := universal.InitGG_Packet(universal.GG_STATUS60, p.Serialize())
	_, err := pOut.Send(cI.Conn)
	if err != nil {
		log.L.Errorf("Error: %s", err)
	}
}

func SendRecvMsg(c GGClient, msg structs.Message) {
	cI := c.GetClientInfoPtr()
	pS := universal.GG_Recv_MSG{
		Sender:   msg.From,
		Seq:      0,
		Time:     uint32(time.Now().Unix()),
		MsgClass: 0x08,
		Content:  msg.Content,
	}
	pOut := universal.InitGG_Packet(universal.GG_RECV_MSG, pS.Serialize())
	_, err := pOut.Send(cI.Conn)
	if err != nil {
		log.L.Errorf("Error: %s", err)
	}
}

func SendNotifyReply(c GGClient, data []byte) {
	cI := c.GetClientInfoPtr()
	pOut := universal.InitGG_Packet(universal.GG_NOTIFY_REPLY60, data)
	_, err := pOut.Send(cI.Conn)
	if err != nil {
		log.L.Debugf("Error: %s", err)
	}
}

func SendPong(c GGClient) {
	cI := c.GetClientInfoPtr()
	pOut := universal.InitGG_Packet(universal.GG_PONG, []byte{})
	_, err := pOut.Send(cI.Conn)
	if err != nil {
		log.L.Errorf("Error: %s", err)
	}
}

package universal

import (
	"bytes"
	"encoding/binary"
)

func GG_NotifyContactDeserialize(data []byte, packetSize uint32, contactList *[]GG_NotifyContact) {
	buf := bytes.NewBuffer(data)
	contactListLen := int(packetSize / GG_NOTIFYCONTACT_SIZE)
	for i := 0; i < contactListLen; i++ {
		contact := GG_NotifyContact{}
		binary.Read(buf, binary.LittleEndian, &contact.UIN)
		binary.Read(buf, binary.LittleEndian, &contact.Type)
		*contactList = append(*contactList, contact)
	}
}

func GG_NotifyReplySerialize(statusChange StatusChangeMsg) []byte {
	// Don't serialize if user not online or invisible
	switch statusChange.Status {
	case GG_STATUS_NOT_AVAIL:
		return make([]byte, 0)
	case GG_STATUS_INVISIBLE:
		return make([]byte, 0)
	}

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, statusChange.UIN)
	binary.Write(buf, binary.LittleEndian, uint8(statusChange.Status)) // It's uint8 only in GG60!
	binary.Write(buf, binary.LittleEndian, uint32(0))
	binary.Write(buf, binary.LittleEndian, uint16(0))
	binary.Write(buf, binary.LittleEndian, uint32(0))
	binary.Write(buf, binary.LittleEndian, uint16(0))
	return buf.Bytes()
}

func GG_NotifyReplyBatchSerialize(uin uint32, contactList []GG_NotifyContact, statusChanges *[]StatusChangeMsg) []byte {
	buf := new(bytes.Buffer)
	//for _, statusChange := range contactList {
	//	binary.Write(buf, binary.LittleEndian, GG_NotifyReplySerialize(StatusChangeMsg{uin, Fetch}))
	//}
	return buf.Bytes()
}

func GG_AddNotify(data []byte, contactList *[]GG_NotifyContact) {
	buf := bytes.NewBuffer(data)
	contact := GG_NotifyContact{}
	binary.Read(buf, binary.LittleEndian, &contact.UIN)
	binary.Read(buf, binary.LittleEndian, &contact.Type)
	*contactList = append(*contactList, contact)
}

package universal

import (
	"bytes"
	"encoding/binary"
	"poggadaj-tcp/constants"
)

func GG_NotifyContactDeserialize(data []byte, packetSize uint32, contactList *[]GG_NotifyContact) {
	buf := bytes.NewBuffer(data)
	contactListLen := int(packetSize / constants.GG_NOTIFYCONTACT_SIZE)
	for i := 0; i < contactListLen; i++ {
		contact := GG_NotifyContact{}
		binary.Read(buf, binary.LittleEndian, &contact.UIN)
		binary.Read(buf, binary.LittleEndian, &contact.Type)
		*contactList = append(*contactList, contact)
	}
}

func GG_AddNotify(data []byte, contactList *[]GG_NotifyContact) GG_NotifyContact {
	buf := bytes.NewBuffer(data)
	contact := GG_NotifyContact{}
	binary.Read(buf, binary.LittleEndian, &contact.UIN)
	binary.Read(buf, binary.LittleEndian, &contact.Type)
	*contactList = append(*contactList, contact)
	return contact
}

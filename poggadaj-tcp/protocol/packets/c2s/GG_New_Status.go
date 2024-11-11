package c2s

import (
	"bytes"
	"encoding/binary"
)

type GG_New_Status struct {
	Status      uint32
	Description []byte
}

func (p *GG_New_Status) Deserialize(data []byte, packetSize uint32) {
	descriptionLength := packetSize - 4
	buf := bytes.NewBuffer(data)

	binary.Read(buf, binary.LittleEndian, &p.Status)
	p.Description = make([]byte, descriptionLength)
	binary.Read(buf, binary.LittleEndian, &p.Description)
}

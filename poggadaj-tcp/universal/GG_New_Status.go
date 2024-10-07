package universal

import (
	"bytes"
	"encoding/binary"
)

type GG_New_Status struct {
	Status uint32
}

func (p *GG_New_Status) Deserialize(data []byte) {
	buf := bytes.NewBuffer(data)

	binary.Read(buf, binary.LittleEndian, &p.Status)
}

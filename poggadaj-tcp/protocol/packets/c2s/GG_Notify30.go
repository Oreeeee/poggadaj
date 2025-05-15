package c2s

import (
	"bytes"
	"encoding/binary"
	"fmt"
	log "poggadaj-shared/logging"
)

type GG_Notify30 struct {
	UINs []uint32
}

func (p *GG_Notify30) Deserialize(data []byte, packetSize uint32) {
	log.L.Debugf("GG_NOTIFY30 contents: %x", data)
	buf := bytes.NewBuffer(data)
	p.UINs = make([]uint32, packetSize/4)
	binary.Read(buf, binary.LittleEndian, &p.UINs)
}

func (p *GG_Notify30) PrettyPrint() []string {
	s := []string{
		fmt.Sprintf("UINs: %d", p.UINs),
	}
	return s
}

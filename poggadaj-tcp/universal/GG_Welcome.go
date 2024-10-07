package universal

import (
	"bytes"
	"encoding/binary"
)

type GG_Welcome struct {
	seed uint32
}

func InitGG_Welcome() *GG_Welcome {
	return &GG_Welcome{seed: 0xFFFFFFFF}
}

func (g *GG_Welcome) Serialize() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, g.seed)
	return buf.Bytes()
}

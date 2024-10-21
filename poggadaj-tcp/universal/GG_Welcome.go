package universal

import (
	"bytes"
	"encoding/binary"
	"os"
	"strconv"
)

type GG_Welcome struct {
	seed uint32
}

func InitGG_Welcome() *GG_Welcome {
	seed64, _ := strconv.ParseUint(os.Getenv("GG_SEED"), 10, 32)
	return &GG_Welcome{seed: uint32(seed64)}
}

func (g *GG_Welcome) Serialize() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, g.seed)
	return buf.Bytes()
}

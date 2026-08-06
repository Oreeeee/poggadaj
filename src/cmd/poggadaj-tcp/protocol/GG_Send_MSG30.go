package protocol

import (
	"fmt"

	"codeberg.org/or3e/poggadaj/internal/utils"
)

type GG_Send_MSG30 struct {
	Recipient uint32
	Seq       uint32
	Content   string
}

func (p *GG_Send_MSG30) Serialize(stream *utils.IOStream) {}

func (p *GG_Send_MSG30) Deserialize(stream *utils.IOStream) {
	p.Recipient = stream.ReadU32()
	p.Seq = stream.ReadU32()
	p.Content = stream.ReadString(-1)
}

func (p *GG_Send_MSG30) PrettyPrint() []string {
	s := []string{
		fmt.Sprintf("Recipient: %d", p.Recipient),
		fmt.Sprintf("Seq: %d", p.Seq),
		fmt.Sprintf("Content: %x", p.Content),
	}
	return s
}

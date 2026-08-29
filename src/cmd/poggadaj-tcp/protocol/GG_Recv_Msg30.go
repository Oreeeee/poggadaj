package protocol

import (
	"fmt"

	"codeberg.org/or3e/poggadaj/internal/utils"
)

type GG_Recv_MSG30 struct {
	Sender   uint32
	Unknown1 uint32
	Content  string
}

func (p *GG_Recv_MSG30) Serialize(stream *utils.IOStream) {
	stream.WriteU32(p.Sender)
	stream.WriteU32(p.Unknown1)
	stream.Write(stream.SerializeString(p.Content, true))
}

func (p *GG_Recv_MSG30) Deserialize(stream *utils.IOStream) {}

func (p *GG_Recv_MSG30) PrettyPrint() []string {
	s := []string{
		fmt.Sprintf("Sender: %d", p.Sender),
		fmt.Sprintf("Seq: %d", p.Unknown1),
		fmt.Sprintf("Content: %x", p.Content),
	}
	return s
}

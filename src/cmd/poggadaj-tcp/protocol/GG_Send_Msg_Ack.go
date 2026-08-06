package protocol

import (
	"fmt"

	"codeberg.org/or3e/poggadaj/internal/utils"
)

type GG_Send_Msg_Ack struct {
	Status    uint32
	Recipient uint32
	Seq       uint32
}

func (p *GG_Send_Msg_Ack) Serialize(stream *utils.IOStream) {
	stream.WriteU32(p.Status)
	stream.WriteU32(p.Recipient)
	stream.WriteU32(p.Seq)
}

func (p *GG_Send_Msg_Ack) Deserialize(stream *utils.IOStream) {}

func (p *GG_Send_Msg_Ack) PrettyPrint() []string {
	s := []string{
		fmt.Sprintf("Status: %d", p.Status),
		fmt.Sprintf("Recipient: %d", p.Recipient),
		fmt.Sprintf("Seq: %d", p.Seq),
	}
	return s
}

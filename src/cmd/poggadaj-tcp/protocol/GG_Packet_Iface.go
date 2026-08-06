package protocol

import "codeberg.org/or3e/poggadaj/internal/utils"

type GG_Packet_Iface interface {
	Serialize(*utils.IOStream)
	Deserialize(*utils.IOStream)
}

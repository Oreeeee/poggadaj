package protocol

import "poggadaj-tcp/utils"

type GG_Packet_Iface interface {
	Serialize(*utils.IOStream)
	Deserialize(*utils.IOStream)
}

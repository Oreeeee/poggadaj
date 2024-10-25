package utils

import (
	"encoding/binary"
	"net"
)

func LeIntToIPv4(ipn uint32) net.IP {
	ip := make(net.IP, 4)
	binary.LittleEndian.PutUint32(ip, ipn)
	return ip
}

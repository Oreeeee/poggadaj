package utils

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strconv"
)

func LeIntToIPv4(ipn uint32) net.IP {
	ip := make(net.IP, 4)
	binary.LittleEndian.PutUint32(ip, ipn)
	return ip
}

func StringifySHA1(hash [64]byte) string {
	s := fmt.Sprintf("%x", hash[:20])
	return s
}

func ChunkString(s string, n int) []string {
	var chunks []string
	for i := 0; i < len(s); i += n {
		end := i + n
		if end > len(s) {
			end = len(s)
		}
		chunks = append(chunks, s[i:end])
	}
	return chunks
}

func CloseChannel(r *bool) {
	*r = false
}

func BoolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func PasswordFitsRestrictions(password string) bool {
	if len(password) < 8 || len(password) > 64 {
		return false
	}
	return true
}

func GetSeed() uint32 {
	seed64, _ := strconv.ParseUint(os.Getenv("GG_SEED"), 10, 32)
	return uint32(seed64)
}

package utils

import "fmt"

func StringifySHA1(hash [64]byte) string {
	s := fmt.Sprintf("%x", hash[:20])
	return s
}

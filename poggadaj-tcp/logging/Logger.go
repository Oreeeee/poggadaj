package logging

import "github.com/charmbracelet/log"

var L *log.Logger

func StructPPrint(packetName string, packetLines []string) {
	L.Debugf(packetName)
	for _, line := range packetLines {
		L.Debugf(line)
	}
}

package logging

func StructPPrint(packetName string, packetLines []string) {
	L.Debugf(packetName)
	for _, line := range packetLines {
		L.Debugf(line)
	}
}

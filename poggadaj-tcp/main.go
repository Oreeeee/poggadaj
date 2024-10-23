package main

import (
	"fmt"
	"github.com/charmbracelet/log"
	"net"
	"os"
	"poggadaj-tcp/database"
	"poggadaj-tcp/gg60"
	"poggadaj-tcp/logging"
	"poggadaj-tcp/universal"
	"time"
)

// This function handles connections before the client version of GG is known.
// Its purpose is to send GG_WELCOME, receive the packet type of the incoming
// packet, and to redirect the handling to the proper handler for the protocol.
func handleConnection(conn net.Conn) {
	defer conn.Close()

	// Here we create a GG_WELCOME packet once the client connects to the server
	ggw := universal.InitGG_Welcome()
	ggwB := ggw.Serialize()
	packet := universal.InitGG_Packet(universal.GG_WELCOME, ggwB)

	_, err := packet.Send(conn)
	if err != nil {
		logging.L.Errorf("Error: %s", err)
	}

	// Wait for the next packet, which will tell us the protocol version handler we need
	pRecv := universal.GG_Packet{}
	if pRecv.Receive(conn) != nil {
		logging.L.Errorf("Error receiving data, dropping connection!: %s", err)
		return
	}

	if pRecv.PacketType == gg60.GG_LOGIN60 {
		logging.L.Infof("Gadu-Gadu 6.0 protocol detected")
		Handle_GG60(conn, pRecv)
	}
}

func main() {
	dbconn, err := database.GetDBConn()
	database.DatabaseConn = dbconn

	database.CacheConn = database.GetCacheConn()

	logging.L = log.NewWithOptions(os.Stdout, log.Options{
		ReportCaller:    true,
		ReportTimestamp: true,
		TimeFormat:      time.DateTime,
		Level:           log.DebugLevel,
	})

	l, err := net.Listen("tcp", fmt.Sprintf("%s:8074", os.Getenv("LISTEN_ADDRESS")))
	if err != nil {
		logging.L.Fatal(err)
		return
	}
	defer l.Close()
	defer database.DatabaseConn.Close()

	logging.L.Infof("Listening on %s:%d", os.Getenv("LISTEN_ADDRESS"), 8074)

	for {
		conn, err := l.Accept()
		if err != nil {
			logging.L.Errorf("Error accepting from %s: %s", conn.RemoteAddr(), err)
			continue
		}

		logging.L.Infof("Accepted connection from %s", conn.RemoteAddr())

		go handleConnection(conn)
	}
}

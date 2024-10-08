package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"poggadaj-tcp/gg60"
	"poggadaj-tcp/universal"
)

// This function handles connections before the client version of GG is known.
// Its purpose is to send GG_WELCOME, receive the packet type of the incoming
// packet, and to redirect the handling to the proper handler for the protocol.
func handleConnection(currConn GGConnection) {
	defer currConn.Conn.Close()

	// Here we create a GG_WELCOME packet once the client connects to the server
	ggw := universal.InitGG_Welcome()
	ggwB := ggw.Serialize()
	packet := universal.InitGG_Packet(universal.GG_WELCOME, ggwB)

	_, err := packet.Send(currConn.Conn)
	if err != nil {
		fmt.Println("Error: ", err)
	}
	fmt.Println("Sent data")

	// Wait for the next packet, which will tell us the protocol version handler we need
	pRecv := universal.GG_Packet{}
	if pRecv.Receive(currConn.Conn) != nil {
		fmt.Println("Error receiving data, dropping connection!: ", err)
		return
	}

	if pRecv.PacketType == gg60.GG_LOGIN60 {
		fmt.Println("Received GG_LOGIN60, switching to Handle_GG60")
		Handle_GG60(currConn, pRecv)
	}
}

func main() {
	ip := "0.0.0.0:8074"
	if len(os.Args) > 1 && os.Args[1] == "dockerless" {
		ip = "127.0.0.1:8074"
	}

	l, err := net.Listen("tcp", ip)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer l.Close()

	fmt.Println("Listening...")

	var connList []*GGConnection
	msgChans := make(map[uint32]chan Message)

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err)
			continue
		}

		fmt.Println("Accepted connection: ", conn.RemoteAddr())

		// Create a connection object
		ggConn := &GGConnection{}
		ggConn.Conn = conn
		ggConn.Authenticated = false
		ggConn.MsgChans = &msgChans
		connList = append(connList, ggConn)

		go handleConnection(*ggConn)
	}
}

package main

import (
	"fmt"
	"log"
	"net"
	"poggadaj-tcp/gg60"
	"poggadaj-tcp/universal"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// Here we create a GG_WELCOME packet once the client connects to the server
	ggw := universal.InitGG_Welcome()
	ggwB := ggw.Serialize()
	packet := universal.InitGG_Packet(universal.GG_WELCOME, ggwB)

	_, err := packet.Send(conn)
	if err != nil {
		fmt.Println("Error: ", err)
	}
	fmt.Println("Sent data")

	for {
		pRecv := universal.GG_Packet{}
		pRecv.Receive(conn)
		fmt.Println("Received data")

		if pRecv.PacketType == gg60.GG_LOGIN60 {
			fmt.Println("Received GG_LOGIN60")
			p := gg60.GG_Login60{}
			p.Deserialize(pRecv.Data)
			fmt.Println("Decoded data: ", p)
			continue
		}
	}
}

func main() {
	l, err := net.Listen("tcp", "127.0.0.1:8074")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer l.Close()

	fmt.Println("Listening...")

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err)
			continue
		}

		fmt.Println("Accepted connection: ", conn.RemoteAddr())
		go handleConnection(conn)
	}
}

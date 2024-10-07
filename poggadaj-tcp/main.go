package main

import (
	"fmt"
	"log"
	"net"
	"poggadaj-tcp/gg60"
	"poggadaj-tcp/universal"
)

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

	for {
		pRecv := universal.GG_Packet{}
		if pRecv.Receive(currConn.Conn) != nil {
			fmt.Println("Error receiving data, dropping connection!: ", err)
			break
		}

		if pRecv.PacketType == gg60.GG_LOGIN60 {
			fmt.Println("Received GG_LOGIN60")
			p := gg60.GG_Login60{}
			p.Deserialize(pRecv.Data)
			fmt.Println("Decoded data: ", p)

			currConn.UIN = p.UIN

			fmt.Println("Sending login response")
			if p.Hash == 4105424095 { // Password: 123
				currConn.Authenticated = true
				fmt.Println("Sending GG_LOGIN_OK")
				pOut := universal.InitGG_Packet(universal.GG_LOGIN_OK, []byte{})
				_, err := pOut.Send(currConn.Conn)
				if err != nil {
					fmt.Println("Error: ", err)
				}
				continue
			} else {
				fmt.Println("Sending GG_LOGIN_FAILED")
				pOut := universal.InitGG_Packet(universal.GG_LOGIN_FAILED, []byte{})
				_, err := pOut.Send(currConn.Conn)
				if err != nil {
					fmt.Println("Error: ", err)
				}
				break
			}
		} else if pRecv.PacketType == universal.GG_LIST_EMPTY {
			fmt.Println("Received GG_LIST_EMPTY")
		} else if pRecv.PacketType == universal.GG_NEW_STATUS {
			fmt.Println("Received GG_NEW_STATUS")
			p := universal.GG_New_Status{}
			p.Deserialize(pRecv.Data)
			fmt.Println("New status: ", p.Status)
		} else {
			fmt.Println("Received unknown packet: ", pRecv.PacketType)
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

	var connList []*GGConnection

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
		connList = append(connList, ggConn)

		go handleConnection(*ggConn)
	}
}

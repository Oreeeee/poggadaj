package main

import (
	"fmt"
	"log"
	"net"
	"poggadaj-tcp/universal"
	"time"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()

	ggw := universal.InitGG_Welcome()
	ggwB := ggw.Serialize()
	response := universal.InitGG_Header(universal.GG_WELCOME, ggwB)
	responseBytes := response.Serialize()

	_, err := conn.Write(responseBytes)
	if err != nil {
		fmt.Println("Error: ", err)
	}
	fmt.Println("Sent data")
	time.Sleep(10 * time.Second)
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

package main

import (
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"log"
	"net"
	"os"
	"poggadaj-tcp/gg60"
	"poggadaj-tcp/universal"
)

var DatabaseConn *pgxpool.Pool
var CacheConn *redis.Client

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
	dbconn, err := GetDBConn()
	DatabaseConn = dbconn

	CacheConn = GetCacheConn()

	l, err := net.Listen("tcp", fmt.Sprintf("%s:8074", os.Getenv("LISTEN_ADDRESS")))
	if err != nil {
		log.Fatal(err)
		return
	}
	defer l.Close()
	defer DatabaseConn.Close()

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

package main

import (
	"fmt"
	"github.com/charmbracelet/log"
	"net"
	"os"
	poggadaj_shared "poggadaj-shared"
	"poggadaj-tcp/database"
	"poggadaj-tcp/logging"
	"time"
)

func main() {
	poggadaj_shared.SharedPrint()
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
		go HandleConnection(conn)
	}
}

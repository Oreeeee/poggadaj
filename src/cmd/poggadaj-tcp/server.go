package main

import (
	"net"
	"os"

	"charm.land/log/v2"
	"codeberg.org/or3e/poggadaj/internal/cache"
	"codeberg.org/or3e/poggadaj/internal/database"
	"codeberg.org/or3e/poggadaj/internal/logging"
)

type Server struct {
	ip          string
	listener    net.Listener
	db          *database.Database
	cache       *cache.Cache
	logger      *log.Logger
	connections []*Client
}

func (server *Server) handleConnection(conn net.Conn) {
	client, _ := NewClient(
		conn,
		server,
		logging.NewLoggerWithPrefix(conn.RemoteAddr().String()),
	)
	server.connections = append(server.connections, client)

	go client.Run()
}

func (server *Server) Run() error {
	l, err := net.Listen("tcp", server.ip)
	if err != nil {
		return err
	}
	defer l.Close()

	server.logger.Infof("Listening on %s:%d", os.Getenv("LISTEN_ADDRESS"), 8074)

	for {
		conn, err := l.Accept()
		if err != nil {
			server.logger.Errorf("Error accepting from %s: %s", conn.RemoteAddr(), err)
			continue
		}

		server.logger.Infof("Accepted connection from %s", conn.RemoteAddr())
		server.handleConnection(conn)
	}
}

func NewServer(dbCfg *database.DatabaseConfig, cache *cache.Cache, ip string) (*Server, error) {
	dbConn, err := database.NewDatabase(dbCfg)
	if err != nil {
		return nil, err
	}

	server := &Server{
		ip:     ip,
		db:     dbConn,
		logger: logging.NewLogger(),
		cache:  cache,
	}

	return server, nil
}

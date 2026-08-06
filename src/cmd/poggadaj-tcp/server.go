// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package main

import (
	"net"
	"os"
	"sync"

	"charm.land/log/v2"
	"codeberg.org/or3e/poggadaj/internal/cache"
	"codeberg.org/or3e/poggadaj/internal/database"
	"codeberg.org/or3e/poggadaj/internal/logging"
)

type Server struct {
	ip            string
	listener      net.Listener
	db            *database.Database
	cache         *cache.Cache
	logger        *log.Logger
	muConnections sync.RWMutex
	connections   []*Client
}

func (server *Server) handleConnection(conn net.Conn) {
	client, _ := NewClient(
		conn,
		server,
		logging.NewLoggerWithPrefix(conn.RemoteAddr().String()),
	)
	// TODO: add a mutex
	server.registerClient(client)
	client.Run()
	server.unregisterClient(client)
}

func (server *Server) registerClient(c *Client) {
	server.muConnections.Lock()
	defer server.muConnections.Unlock()

	// Try to find an empty space in the array first
	for i, client := range server.connections {
		if client != nil {
			continue
		}

		server.logger.Info("Found an empty space in the array")
		server.connections[i] = c
		return
	}

	// No empty space found, so append the connection
	server.logger.Info("No empty space in the array, appending the new connection")
	server.connections = append(server.connections, c)
}

func (server *Server) unregisterClient(c *Client) {
	server.muConnections.Lock()
	defer server.muConnections.Unlock()

	for i, client := range server.connections {
		if client != c {
			continue
		}

		server.connections[i] = nil
	}
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
		go server.handleConnection(conn)
	}
}

func NewServer(dbCfg *database.DatabaseConfig, cache *cache.Cache, ip string) (*Server, error) {
	dbConn, err := database.NewDatabase(dbCfg, logging.NewLoggerWithPrefix("database"))
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

// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package database

import (
	"context"
	"fmt"
	"net/url"

	"charm.land/log/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	Ads     AdsRepository
	Clients ClientsRepository
	Pubdir  PubdirRepository
	Users   UsersRepository

	conn   *pgxpool.Pool
	logger *log.Logger
}

type DatabaseConfig struct {
	Host, Port, Username, Password string
}

func NewDatabase(cfg *DatabaseConfig, logger *log.Logger) (*Database, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/poggadaj?sslmode=disable",
		cfg.Username, url.QueryEscape(cfg.Password), cfg.Host, cfg.Port)
	conn, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, err
	}

	db := &Database{
		Ads: AdsRepository{
			conn:   conn,
			logger: logger,
		},
		Clients: ClientsRepository{
			conn:   conn,
			logger: logger,
		},
		Pubdir: PubdirRepository{
			conn:   conn,
			logger: logger,
		},
		Users: UsersRepository{
			conn:   conn,
			logger: logger,
		},
		conn:   conn,
		logger: logger,
	}

	return db, nil
}

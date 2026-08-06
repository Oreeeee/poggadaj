package database

import (
	"context"
	"fmt"

	"charm.land/log/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	conn   *pgxpool.Pool
	logger *log.Logger
}

type DatabaseConfig struct {
	Host, Port, Username, Password string
}

func NewDatabase(cfg *DatabaseConfig, logger *log.Logger) (*Database, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/poggadaj?sslmode=disable",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port)
	conn, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, err
	}

	db := &Database{
		conn:   conn,
		logger: logger,
	}

	return db, nil
}

package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	conn *pgxpool.Pool
}

type DatabaseConfig struct {
	Host, Port, Username, Password string
}

func NewDatabase(cfg *DatabaseConfig) (*Database, error) {
	db := &Database{}
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/poggadaj?sslmode=disable",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port)
	conn, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, err
	}

	db.conn = conn

	return db, nil
}

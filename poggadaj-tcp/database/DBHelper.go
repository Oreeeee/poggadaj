package database

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
)

func GetDBConn() (*pgxpool.Pool, error) {
	dbaddr := os.Getenv("DB_ADDRESS")
	password := os.Getenv("DB_PASSWORD")
	return pgxpool.New(
		context.Background(),
		fmt.Sprintf(
			"user=poggadaj password=%s host=%s port=5432 dbname=poggadaj sslmode=disable",
			password,
			dbaddr,
		),
	)
}

func GetGG32Hash(uin uint32) (uint32, error) {
	var GG32Hash_i64 int64
	err := DatabaseConn.QueryRow(
		context.Background(),
		fmt.Sprintf("SELECT password_gg32 FROM gguser WHERE uin=%d", uin),
	).Scan(&GG32Hash_i64)
	return uint32(GG32Hash_i64), err
}

var DatabaseConn *pgxpool.Pool

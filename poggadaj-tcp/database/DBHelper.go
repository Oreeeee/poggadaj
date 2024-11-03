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

func GetAncientHash(uin uint32) (uint32, error) {
	var GGAncientHash int64
	err := DatabaseConn.QueryRow(
		context.Background(),
		"SELECT password_gg_ancient FROM gguser WHERE uin=$1",
		uin,
	).Scan(&GGAncientHash)
	return uint32(GGAncientHash), err
}

func GetGG32Hash(uin uint32) (uint32, error) {
	var GG32Hash_i64 int64
	err := DatabaseConn.QueryRow(
		context.Background(),
		"SELECT password_gg32 FROM gguser WHERE uin=$1",
		uin,
	).Scan(&GG32Hash_i64)
	return uint32(GG32Hash_i64), err
}

func GetSHA1Hash(uin uint32) (string, error) {
	var SHA1 string
	err := DatabaseConn.QueryRow(
		context.Background(),
		"SELECT password_sha1 FROM gguser WHERE uin=$1",
		uin,
	).Scan(&SHA1)
	return SHA1, err
}

var DatabaseConn *pgxpool.Pool

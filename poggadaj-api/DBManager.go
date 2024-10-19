package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
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

func CreateUser(name string, password string) (int, error) {
	// Hash the password
	pwdHash, err := HashPassword(password)
	if err != nil {
		return 0, err
	}

	// Create the user
	query := "INSERT INTO gguser (name, password) VALUES (@name, @password)"
	args := pgx.NamedArgs{
		"name":     name,
		"password": pwdHash,
	}
	_, err2 := DatabaseConn.Exec(context.Background(), query, args)
	if err2 != nil {
		return 0, err2
	}

	// Allocate a new UIN for the user
	var newUserUIN int
	query = "UPDATE gguser SET uin=nextval('uin_seq') WHERE name=$1 RETURNING uin"
	err3 := DatabaseConn.QueryRow(context.Background(), query, name).Scan(&newUserUIN)
	if err3 != nil {
		return 0, err3
	}

	return newUserUIN, nil
}

func GetUserPasswordHash(name string) (string, error) {
	query := "SELECT password FROM gguser WHERE name=$1"
	var passwordHash string
	err := DatabaseConn.QueryRow(context.Background(), query, name).Scan(&passwordHash)
	if err != nil {
		return "", err
	}
	return passwordHash, nil
}

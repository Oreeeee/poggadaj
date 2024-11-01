package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"poggadaj-api/errs"
	"time"
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

func CreateUser(regBody RegisterRequest) (int, error) {
	var GGAncientHash uint32
	var GG32Hash uint32
	var GGSHA1Hash string

	// Hash the password
	pwdHash, err := HashPassword(regBody.Password)
	if err != nil {
		return 0, err
	}

	dbArgs := pgx.NamedArgs{
		"name":     regBody.Username,
		"password": pwdHash,
	}

	if regBody.GGAncientPassword != "" {
		GGAncientHash = GGAncientLoginHash(regBody.GGAncientPassword, GetSeed())
		dbArgs["password_gg_ancient_hash"] = GGAncientHash
	}
	if regBody.GG32Password != "" {
		GG32Hash = GG32LoginHash(regBody.GG32Password, GetSeed())
		dbArgs["password_gg32"] = GG32Hash
	}
	if regBody.GGSHA1Password != "" {
		GGSHA1Hash = GGSHA1LoginHash(regBody.GGSHA1Password, GetSeed())
		dbArgs["password_sha1"] = GGSHA1Hash
	}

	// Create the user
	query := "INSERT INTO gguser (name, password, password_gg_ancient, password_gg32, password_sha1) VALUES (@name, @password, @password_gg_ancient, @password_gg32, @password_sha1)"
	_, err2 := DatabaseConn.Exec(context.Background(), query, dbArgs)
	if err2 != nil {
		return 0, err2
	}

	// Allocate a new UIN for the user
	var newUserUIN int
	query = "UPDATE gguser SET uin=nextval('uin_seq') WHERE name=$1 RETURNING uin"
	err3 := DatabaseConn.QueryRow(context.Background(), query, regBody.Username).Scan(&newUserUIN)
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

func UpdateUserPassword(name string, chgreq ChangePasswordRequest) error {
	switch chgreq.PasswordType {
	case 0:
		// Website password
		hashedPassword, err := HashPassword(chgreq.Password)
		if err != nil {
			return err
		}
		query := "UPDATE gguser SET password=$1 WHERE name=$2"
		_, err2 := DatabaseConn.Exec(context.Background(), query, hashedPassword, name)
		return err2
	case 1:
		// Ancient password
		hashedPassword := GGAncientLoginHash(chgreq.Password, GetSeed())
		query := "UPDATE gguser SET password_gg_ancient=$1 WHERE name=$2"
		_, err := DatabaseConn.Exec(context.Background(), query, hashedPassword, name)
		return err
	case 2:
		// GG32 password
		hashedPassword := GG32LoginHash(chgreq.Password, GetSeed())
		query := "UPDATE gguser SET password_gg32=$1 WHERE name=$2"
		_, err := DatabaseConn.Exec(context.Background(), query, hashedPassword, name)
		return err
	case 3:
		// SHA1 password
		hashedPassword := GGSHA1LoginHash(chgreq.Password, GetSeed())
		query := "UPDATE gguser SET password_sha1=$1 WHERE name=$2"
		_, err := DatabaseConn.Exec(context.Background(), query, hashedPassword, name)
		return err
	default:
		return errs.WrongPasswordType{PasswordType: chgreq.PasswordType}
	}
	return nil
}

func GetUserData(name string) (int, time.Time, error) {
	query := "SELECT uin, joined FROM gguser WHERE name=$1"
	var uin int
	var joined time.Time
	err := DatabaseConn.QueryRow(context.Background(), query, name).Scan(&uin, &joined)
	return uin, joined, err
}

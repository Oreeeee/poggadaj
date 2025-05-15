package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"poggadaj-api/errs"
	"poggadaj-shared/security/argon2"
	"poggadaj-shared/security/gg"
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
	pwdHash, err := argon2.HashPassword(regBody.Password)
	if err != nil {
		return 0, err
	}

	dbArgs := pgx.NamedArgs{
		"name":     regBody.Username,
		"password": pwdHash,
	}

	if regBody.GGAncientPassword != "" {
		GGAncientHash = gg.GGAncientLoginHash(regBody.GGAncientPassword, GetSeed())
		dbArgs["password_gg_ancient"] = GGAncientHash
	}
	if regBody.GG32Password != "" {
		GG32Hash = gg.GG32LoginHash(regBody.GG32Password, GetSeed())
		dbArgs["password_gg32"] = GG32Hash
	}
	if regBody.GGSHA1Password != "" {
		GGSHA1Hash = gg.GGSHA1LoginHash(regBody.GGSHA1Password, GetSeed())
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

func UpdateWebsitePassword(name string, password string) error {
	hashedPassword, err := argon2.HashPassword(password)
	if err != nil {
		return err
	}
	query := "UPDATE gguser SET password=$1 WHERE name=$2"
	_, err2 := DatabaseConn.Exec(context.Background(), query, hashedPassword, name)
	return err2
}

func UpdateAncientPassword(name string, password string) error {
	hashedPassword := gg.GGAncientLoginHash(password, GetSeed())
	query := "UPDATE gguser SET password_gg_ancient=$1 WHERE name=$2"
	_, err := DatabaseConn.Exec(context.Background(), query, hashedPassword, name)
	return err
}

func UpdateGG32Password(name string, password string) error {
	hashedPassword := gg.GG32LoginHash(password, GetSeed())
	query := "UPDATE gguser SET password_gg32=$1 WHERE name=$2"
	_, err := DatabaseConn.Exec(context.Background(), query, hashedPassword, name)
	return err
}

func UpdateSHA1Password(name string, password string) error {
	hashedPassword := gg.GGSHA1LoginHash(password, GetSeed())
	query := "UPDATE gguser SET password_sha1=$1 WHERE name=$2"
	_, err := DatabaseConn.Exec(context.Background(), query, hashedPassword, name)
	return err
}

func UpdateUserPassword(name string, chgreq ChangePasswordRequest) error {
	switch chgreq.PasswordType {
	case 0:
		// Website password
		return UpdateWebsitePassword(name, chgreq.Password)
	case 1:
		// Ancient password
		return UpdateAncientPassword(name, chgreq.Password)
	case 2:
		// GG32 password
		return UpdateGG32Password(name, chgreq.Password)
	case 3:
		return UpdateSHA1Password(name, chgreq.Password)
	default:
		return errs.WrongPasswordType{PasswordType: chgreq.PasswordType}
	}
}

func GetUserData(name string) (int, time.Time, error) {
	query := "SELECT uin, joined FROM gguser WHERE name=$1"
	var uin int
	var joined time.Time
	err := DatabaseConn.QueryRow(context.Background(), query, name).Scan(&uin, &joined)
	return uin, joined, err
}

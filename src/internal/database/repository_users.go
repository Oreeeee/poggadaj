// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package database

import (
	"context"

	"charm.land/log/v2"
	"codeberg.org/or3e/poggadaj/internal/structs"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UsersRepository struct {
	conn   *pgxpool.Pool
	logger *log.Logger
}

func (r *UsersRepository) GetPasswords(uin uint32) (*structs.PasswordSet, error) {
	set := &structs.PasswordSet{}

	var ancientTmp *int64
	var gg32Tmp *int64
	var sha1Tmp *string
	var argon2Tmp *string

	err := r.conn.QueryRow(
		context.Background(),
		"SELECT password_gg_ancient, password_gg32, password_sha1, password FROM gguser WHERE uin=$1",
		uin,
	).Scan(&ancientTmp, &gg32Tmp, &sha1Tmp, &argon2Tmp)

	if err != nil {
		return nil, err
	}

	if ancientTmp != nil {
		set.Ancient = uint32(*ancientTmp)
	}

	if gg32Tmp != nil {
		set.GG32 = uint32(*gg32Tmp)
	}

	if sha1Tmp != nil {
		set.SHA1 = *sha1Tmp
	} else {
		set.SHA1 = ""
	}

	if argon2Tmp != nil {
		set.Argon2 = *argon2Tmp
	} else {
		set.Argon2 = ""
	}

	return set, nil
}

func (r *UsersRepository) CreateUser(name string, email string, password string, ggAncientHash uint32, gg32Hash uint32, ggSha1Hash string) (int, error) {
	// TODO: use transation here in case something goes wrong

	_, err := r.conn.Exec(context.Background(),
		"INSERT INTO gguser (name, email, password, password_gg_ancient, password_gg32, password_sha1) VALUES ($1, $2, $3, $4, $5, $6)",
		name,
		email,
		password,
		ggAncientHash,
		gg32Hash,
		ggSha1Hash,
	)
	if err != nil {
		return 0, err
	}

	// Allocate a new UIN for the user
	var newUserUin int
	err = r.conn.QueryRow(
		context.Background(),
		"UPDATE gguser SET uin=nextval('uin_seq') WHERE name=$1 RETURNING uin",
		name,
	).Scan(&newUserUin)

	if err != nil {
		return 0, err
	}

	return newUserUin, nil
}

func (r *UsersRepository) GetUserPasswordHashWithUin(name string) (uint, string, error) {
	query := "SELECT uin, password FROM gguser WHERE name=$1"
	var uin uint
	var passwordHash string
	err := r.conn.QueryRow(context.Background(), query, name).Scan(&uin, &passwordHash)
	if err != nil {
		return 0, "", err
	}
	return uin, passwordHash, nil
}

func (r *UsersRepository) UpdateUserPassword(uin uint, password string, ggAncientHash uint32, gg32Hash uint32, ggSha1Hash string) error {
	_, err := r.conn.Exec(
		context.Background(),
		"UPDATE gguser SET password=$1, password_gg_ancient=$2, password_gg32=$3, password_sha1=$4 WHERE uin=$5",
		password,
		ggAncientHash,
		gg32Hash,
		ggSha1Hash,
		uin,
	)
	return err
}

func (r *UsersRepository) GetUserDataByUin(uin uint) (*structs.UserData, error) {
	data := &structs.UserData{}
	var emailTmp *string
	err := r.conn.QueryRow(context.Background(), "SELECT uin, name, email, joined FROM gguser WHERE uin = $1", uin).Scan(
		&data.UIN,
		&data.WebUsername,
		&emailTmp,
		&data.JoinedDate,
	)

	if emailTmp == nil {
		data.Email = ""
	} else {
		data.Email = *emailTmp
	}

	return data, err
}

func (r *UsersRepository) UpdateUserEmail(uin uint, email string) error {
	_, err := r.conn.Exec(
		context.Background(),
		"UPDATE gguser SET email = $1 WHERE uin = $2",
		email,
		uin,
	)
	return err
}

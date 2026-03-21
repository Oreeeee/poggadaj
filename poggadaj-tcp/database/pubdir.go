package database

import (
	"context"
	"poggadaj-tcp/pubdir"
)

func GetPubdirDataByUin(uin uint32) (*pubdir.PubdirEntry, error) {
	entry := &pubdir.PubdirEntry{}
	err := DatabaseConn.QueryRow(
		context.Background(),
		"SELECT uin, firstname, lastname, nickname, gender, birthyear, city, familyname, familycity FROM pubdir WHERE uin = $1",
		uin,
	).Scan(
		&entry.UIN,
		&entry.Firstname,
		&entry.Lastname,
		&entry.Nickname,
		&entry.Gender,
		&entry.Birthyear,
		&entry.City,
		&entry.FamilyName,
		&entry.FamilyCity,
	)

	if err != nil {
		return nil, err
	}

	return entry, nil
}

func WritePubdirData(uin uint32, entry *pubdir.PubdirEntry) error {
	_, err := DatabaseConn.Exec(context.Background(),
		`INSERT INTO pubdir (uin, firstname, lastname, nickname, gender, birthyear, city, familyname, familycity)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (uin) DO UPDATE SET
		uin = $1, firstname = $2, lastname = $3, nickname = $4, gender = $5, birthyear = $6, city = $7, familyname = $8, familycity = $9`,
		uin, entry.Firstname, entry.Lastname, entry.Nickname, entry.Gender, entry.Birthyear, entry.City, entry.FamilyName, entry.FamilyCity,
	)
	return err
}

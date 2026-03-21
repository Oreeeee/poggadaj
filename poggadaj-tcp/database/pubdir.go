package database

import (
	"bytes"
	"context"
	"fmt"
	"poggadaj-tcp/pubdir"

	"github.com/jackc/pgx/v5"
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

func SearchInPubdir(query *pubdir.PubdirEntry) ([]pubdir.PubdirEntry, error) {
	// TODO: Add support for age ranges
	// TODO: Add support for only-online option
	// TODO: Add support for pagination

	results := []pubdir.PubdirEntry{}

	// Since the lookup parameters can vary by query, we need to dynamically build the SQL query
	dbColumns := []string{}
	dbArgs := pgx.NamedArgs{
		"uin":           query.UIN,
		"firstname":     query.Firstname,
		"lastname":      query.Lastname,
		"nickname":      query.Nickname,
		"gender":        query.Gender,
		"min_birthyear": query.MinBirthyear,
		"max_birthyear": query.Birthyear,
		"city":          query.City,
	}

	if query.Firstname != "" {
		dbColumns = append(dbColumns, "firstname")
	}

	if query.Lastname != "" {
		dbColumns = append(dbColumns, "lastname")
	}

	if query.Nickname != "" {
		dbColumns = append(dbColumns, "nickname")
	}

	if query.Gender != 0 {
		dbColumns = append(dbColumns, "gender")
	}

	if query.City != "" {
		dbColumns = append(dbColumns, "city")
	}

	var stmtBuilder bytes.Buffer
	fmt.Fprint(&stmtBuilder, "SELECT uin, firstname, lastname, birthyear, city FROM pubdir")
	if len(dbColumns) != 0 {
		fmt.Fprint(&stmtBuilder, " WHERE ")
		lastIndexInColumns := len(dbColumns) - 1

		// Build the query with the specified columns.
		// Named args are used here to prevent injection
		for idx, v := range dbColumns {
			fmt.Fprintf(&stmtBuilder, "%s = @%s", v, v)

			if idx != lastIndexInColumns {
				// Only add the AND when the current arg isn't last
				fmt.Fprintf(&stmtBuilder, " AND ")
			}
		}
	}

	rows, err := DatabaseConn.Query(context.Background(), stmtBuilder.String(), dbArgs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		result := pubdir.PubdirEntry{}
		err = rows.Scan(&result.UIN, &result.Firstname, &result.Lastname, &result.Birthyear, &result.City)
		if err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	return results, nil
}

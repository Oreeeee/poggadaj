// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package database

import (
	"bytes"
	"context"
	"fmt"

	"charm.land/log/v2"

	"codeberg.org/or3e/poggadaj/internal/structs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PubdirRepository struct {
	conn   *pgxpool.Pool
	logger *log.Logger
}

func (r *PubdirRepository) PutUserList(userList []structs.UserListRequest, uin uint32) {
	// TODO: Clean up
	batch := &pgx.Batch{}
	for _, user := range userList {
		dbArgs := pgx.NamedArgs{
			"owner_uin":       uin,
			"firstname":       user.FirstName,
			"lastname":        user.LastName,
			"pseudonym":       user.Pseudonym,
			"display_name":    user.DisplayName,
			"mobile_number":   user.MobileNumber,
			"grp":             user.Group,
			"uin":             user.UIN,
			"email":           user.Email,
			"avail_sound":     user.AvailSound,
			"avail_path":      user.AvailPath,
			"msg_sound":       user.MsgSound,
			"msg_path":        user.MsgPath,
			"hidden":          user.Hidden,
			"landline_number": user.LandlineNumber,
		}
		batch.Queue("INSERT INTO ggcontact (owner_uin, firstname, lastname, pseudonym, display_name, mobile_number, grp, uin, email, avail_sound, avail_path, msg_sound, msg_path, hidden, landline_number) VALUES (@owner_uin, @firstname, @lastname, @pseudonym, @display_name, @mobile_number, @grp, @uin, @email, @avail_sound, @avail_path, @msg_sound, @msg_path, @hidden, @landline_number) ON CONFLICT (owner_uin, firstname, lastname, pseudonym, display_name, mobile_number, grp, uin, email, avail_sound, avail_path, msg_sound, msg_path, hidden, landline_number) DO NOTHING", dbArgs)
	}
	res := r.conn.SendBatch(context.Background(), batch)

	for i := 0; i < len(userList); i++ {
		_, err := res.Exec()
		if err != nil {
			r.logger.Errorf("Failed to execute batch insert: %v\n", err)
		}
	}

	err := res.Close()
	if err != nil {
		r.logger.Errorf("Failed to close batch results: %v\n", err)
	}
}

func (r *PubdirRepository) GetUserList(uin uint32) []structs.UserListRequest {
	rows, err := r.conn.Query(context.Background(), "SELECT firstname, lastname, pseudonym, display_name, mobile_number, grp, uin, email, avail_sound, avail_path, msg_sound, msg_path, hidden, landline_number FROM ggcontact WHERE owner_uin=$1", uin)
	if err != nil {
		r.logger.Errorf("Failed to execute query: %v\n", err)
	}
	defer rows.Close()

	var userList []structs.UserListRequest
	for rows.Next() {
		var user structs.UserListRequest
		err := rows.Scan(&user.FirstName, &user.LastName, &user.Pseudonym, &user.DisplayName, &user.MobileNumber, &user.Group, &user.UIN, &user.Email, &user.AvailSound, &user.AvailPath, &user.MsgSound, &user.MsgPath, &user.Hidden, &user.LandlineNumber)
		if err != nil {
			r.logger.Errorf("Failed to scan row: %v\n", err)
		}
		userList = append(userList, user)
	}

	if rows.Err() != nil {
		r.logger.Errorf("Failed to execute query: %v\n", rows.Err())
	}

	return userList
}

func (r *PubdirRepository) DeleteUserList(uin uint32) error {
	_, err := r.conn.Exec(context.Background(), "DELETE FROM ggcontact WHERE owner_uin=$1", uin)
	return err
}

func (r *PubdirRepository) GetPubdirDataByUin(uin uint32) (*structs.PubdirEntry, error) {
	entry := &structs.PubdirEntry{}
	err := r.conn.QueryRow(
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

func (r *PubdirRepository) WritePubdirData(uin uint32, entry *structs.PubdirEntry) error {
	_, err := r.conn.Exec(context.Background(),
		`INSERT INTO pubdir (uin, firstname, lastname, nickname, gender, birthyear, city, familyname, familycity)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (uin) DO UPDATE SET
		uin = $1, firstname = $2, lastname = $3, nickname = $4, gender = $5, birthyear = $6, city = $7, familyname = $8, familycity = $9`,
		uin, entry.Firstname, entry.Lastname, entry.Nickname, entry.Gender, entry.Birthyear, entry.City, entry.FamilyName, entry.FamilyCity,
	)
	return err
}

func (r *PubdirRepository) SearchInPubdir(query *structs.PubdirEntry) ([]structs.PubdirEntry, uint32, error) {
	// TODO: Add support for only-online option

	results := []structs.PubdirEntry{}

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
		"start":         query.Start,
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
	fmt.Fprint(&stmtBuilder, "SELECT uin, firstname, lastname, birthyear, city, gender FROM pubdir WHERE ")
	if len(dbColumns) != 0 {
		lastIndexInColumns := len(dbColumns) - 1

		// Build the query with the specified columns.
		// Named args are used here to prevent injection
		for idx, v := range dbColumns {
			switch dbArgs[v].(type) {
			case string:
				// Do case insensitivity, allow any string before and after the search arg
				fmt.Fprintf(&stmtBuilder, "%s ILIKE '%%' || @%s || '%%'", v, v)
			default:
				fmt.Fprintf(&stmtBuilder, "%s = @%s", v, v)
			}

			if idx != lastIndexInColumns {
				// Only add the AND when the current arg isn't last
				fmt.Fprintf(&stmtBuilder, " AND ")
			}
		}
	} else {
		// Put a neutral statement for later things
		fmt.Fprintf(&stmtBuilder, "TRUE")
	}

	if query.Start != 0 {
		// Continue the search
		fmt.Fprintf(&stmtBuilder, " AND uin > @start")
	}

	if query.YearIsRange {
		fmt.Fprintf(&stmtBuilder, " AND birthyear BETWEEN @min_birthyear AND @max_birthyear")
	}

	fmt.Fprintf(&stmtBuilder, " ORDER BY uin ASC LIMIT 20")

	rows, err := r.conn.Query(context.Background(), stmtBuilder.String(), dbArgs)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		result := structs.PubdirEntry{}
		err = rows.Scan(&result.UIN, &result.Firstname, &result.Lastname, &result.Birthyear, &result.City, &result.Gender)
		if err != nil {
			return nil, 0, err
		}

		results = append(results, result)
	}

	nextStart := uint32(0)
	if len(results) > 0 {
		nextStart = results[len(results)-1].UIN
	}

	return results, nextStart, nil
}

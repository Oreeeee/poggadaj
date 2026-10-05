// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package database

import (
	"bytes"
	"context"
	"fmt"
	"maps"
	"slices"

	"codeberg.org/or3e/poggadaj/internal/security/argon2"
	"codeberg.org/or3e/poggadaj/internal/structs"
	"github.com/jackc/pgx/v5"
)

func (db *Database) GetAncientHash(uin uint32) (uint32, error) {
	var GGAncientHash int64
	err := db.conn.QueryRow(
		context.Background(),
		"SELECT password_gg_ancient FROM gguser WHERE uin=$1",
		uin,
	).Scan(&GGAncientHash)
	return uint32(GGAncientHash), err
}

func (db *Database) GetGG32Hash(uin uint32) (uint32, error) {
	var GG32Hash_i64 int64
	err := db.conn.QueryRow(
		context.Background(),
		"SELECT password_gg32 FROM gguser WHERE uin=$1",
		uin,
	).Scan(&GG32Hash_i64)
	return uint32(GG32Hash_i64), err
}

func (db *Database) GetSHA1Hash(uin uint32) (string, error) {
	var SHA1 string
	err := db.conn.QueryRow(
		context.Background(),
		"SELECT password_sha1 FROM gguser WHERE uin=$1",
		uin,
	).Scan(&SHA1)
	return SHA1, err
}

func (db *Database) PutUserList(userList []structs.UserListRequest, uin uint32) {
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
	res := db.conn.SendBatch(context.Background(), batch)

	for i := 0; i < len(userList); i++ {
		_, err := res.Exec()
		if err != nil {
			db.logger.Errorf("Failed to execute batch insert: %v\n", err)
		}
	}

	err := res.Close()
	if err != nil {
		db.logger.Errorf("Failed to close batch results: %v\n", err)
	}
}

func (db *Database) GetUserList(uin uint32) []structs.UserListRequest {
	rows, err := db.conn.Query(context.Background(), "SELECT firstname, lastname, pseudonym, display_name, mobile_number, grp, uin, email, avail_sound, avail_path, msg_sound, msg_path, hidden, landline_number FROM ggcontact WHERE owner_uin=$1", uin)
	if err != nil {
		db.logger.Errorf("Failed to execute query: %v\n", err)
	}
	defer rows.Close()

	var userList []structs.UserListRequest
	for rows.Next() {
		var user structs.UserListRequest
		err := rows.Scan(&user.FirstName, &user.LastName, &user.Pseudonym, &user.DisplayName, &user.MobileNumber, &user.Group, &user.UIN, &user.Email, &user.AvailSound, &user.AvailPath, &user.MsgSound, &user.MsgPath, &user.Hidden, &user.LandlineNumber)
		if err != nil {
			db.logger.Errorf("Failed to scan row: %v\n", err)
		}
		userList = append(userList, user)
	}

	if rows.Err() != nil {
		db.logger.Errorf("Failed to execute query: %v\n", rows.Err())
	}

	return userList
}

func (db *Database) DeleteUserList(uin uint32) error {
	_, err := db.conn.Exec(context.Background(), "DELETE FROM ggcontact WHERE owner_uin=$1", uin)
	return err
}

func (db *Database) GetPubdirDataByUin(uin uint32) (*structs.PubdirEntry, error) {
	entry := &structs.PubdirEntry{}
	err := db.conn.QueryRow(
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

func (db *Database) WritePubdirData(uin uint32, entry *structs.PubdirEntry) error {
	_, err := db.conn.Exec(context.Background(),
		`INSERT INTO pubdir (uin, firstname, lastname, nickname, gender, birthyear, city, familyname, familycity)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (uin) DO UPDATE SET
		uin = $1, firstname = $2, lastname = $3, nickname = $4, gender = $5, birthyear = $6, city = $7, familyname = $8, familycity = $9`,
		uin, entry.Firstname, entry.Lastname, entry.Nickname, entry.Gender, entry.Birthyear, entry.City, entry.FamilyName, entry.FamilyCity,
	)
	return err
}

func (db *Database) SearchInPubdir(query *structs.PubdirEntry) ([]structs.PubdirEntry, uint32, error) {
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

	rows, err := db.conn.Query(context.Background(), stmtBuilder.String(), dbArgs)
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

func (db *Database) GetAds(bannerType int) []structs.Ad {
	query := fmt.Sprintf("SELECT adtype, bannertype, image, html FROM adserver_ad WHERE bannertype=%d", bannerType)
	ads := make([]structs.Ad, 0)

	rows, err := db.conn.Query(context.Background(), query)
	if err != nil {
		fmt.Println(err)
		return ads
	}
	defer rows.Close()

	for rows.Next() {
		ad := structs.Ad{}
		err := rows.Scan(&ad.AdType, &ad.BannerType, &ad.Image, &ad.Html)
		if err != nil {
			fmt.Println(err)
		}
		ads = append(ads, ad)
	}

	return ads
}

func (db *Database) CreateUserNew(name string, email string, password string, ggAncientHash uint32, gg32Hash uint32, ggSha1Hash string) (int, error) {
	// TODO: use transation here in case something goes wrong

	_, err := db.conn.Exec(context.Background(),
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
	err = db.conn.QueryRow(
		context.Background(),
		"UPDATE gguser SET uin=nextval('uin_seq') WHERE name=$1 RETURNING uin",
		name,
	).Scan(&newUserUin)

	if err != nil {
		return 0, err
	}

	return newUserUin, nil
}

func (db *Database) GetUserPasswordHashWithUin(name string) (uint, string, error) {
	query := "SELECT uin, password FROM gguser WHERE name=$1"
	var uin uint
	var passwordHash string
	err := db.conn.QueryRow(context.Background(), query, name).Scan(&uin, &passwordHash)
	if err != nil {
		return 0, "", err
	}
	return uin, passwordHash, nil
}

func (db *Database) GetUserPasswordByUin(uin uint) (string, error) {
	query := "SELECT password FROM gguser WHERE uin=$1"
	var passwordHash string
	err := db.conn.QueryRow(context.Background(), query, uin).Scan(&passwordHash)
	if err != nil {
		return "", err
	}
	return passwordHash, nil
}

func (db *Database) UpdateUserPassword(uin uint, password string, ggAncientHash uint32, gg32Hash uint32, ggSha1Hash string) error {
	_, err := db.conn.Exec(
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

func (db *Database) UpdateWebsitePassword(name string, password string) error {
	hashedPassword, err := argon2.HashPassword(password)
	if err != nil {
		return err
	}
	query := "UPDATE gguser SET password=$1 WHERE name=$2"
	_, err2 := db.conn.Exec(context.Background(), query, hashedPassword, name)
	return err2
}

func (db *Database) GetUserDataByUin(uin uint) (*structs.UserData, error) {
	data := &structs.UserData{}
	var emailTmp *string
	err := db.conn.QueryRow(context.Background(), "SELECT uin, name, email, joined FROM gguser WHERE uin = $1", uin).Scan(
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

func (db *Database) GetClients(language string) ([]*structs.WebClient, error) {
	clients := []*structs.WebClient{}
	clientsById := map[int]*structs.WebClient{}

	rows, err := db.conn.Query(
		context.Background(),
		"SELECT c.id, c.name, c.image_url, d.description FROM clients c JOIN client_descriptions d ON c.id = d.client_id WHERE d.language = $1 ORDER BY c.priority ASC",
		language,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		client := &structs.WebClient{}

		err = rows.Scan(
			&client.Id,
			&client.Name,
			&client.ImageUrl,
			&client.Description,
		)

		if err != nil {
			return nil, err
		}

		clients = append(clients, client)
		clientsById[client.Id] = client
	}

	// Query the download links and add to the clients
	keys := slices.Collect(maps.Keys(clientsById))
	downloads, err := db.GetClientsDownloads(keys)
	if err != nil {
		return nil, err
	}

	for _, v := range slices.Collect(maps.Keys(downloads)) {
		for _, downloadsEntry := range downloads[v] {
			clientsById[v].Downloads = append(clientsById[v].Downloads, downloadsEntry)
		}
	}

	return clients, nil
}

func (db *Database) GetClientsDownloads(clientIds []int) (map[int][]*structs.WebClientDownload, error) {
	downloads := map[int][]*structs.WebClientDownload{}
	if len(clientIds) == 0 {
		// TODO: Would be nice to return a custom error here?
		return downloads, nil
	}

	rows, err := db.conn.Query(
		context.Background(),
		"SELECT client_id, file_variant, url FROM client_downloads WHERE client_id = ANY($1) ORDER BY file_variant ASC",
		clientIds,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		clientId := 0
		downloadEntry := &structs.WebClientDownload{}

		err = rows.Scan(
			&clientId,
			&downloadEntry.Type,
			&downloadEntry.Url,
		)
		if err != nil {
			return nil, err
		}

		if _, found := downloads[clientId]; !found {
			downloads[clientId] = []*structs.WebClientDownload{}
		}

		downloads[clientId] = append(downloads[clientId], downloadEntry)
	}

	return downloads, nil
}

func (db *Database) UpdateUserEmail(uin uint, email string) error {
	_, err := db.conn.Exec(
		context.Background(),
		"UPDATE gguser SET email = $1 WHERE uin = $2",
		email,
		uin,
	)
	return err
}

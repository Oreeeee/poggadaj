package database

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"codeberg.org/or3e/poggadaj/cmd/poggadaj-api/errs"
	"codeberg.org/or3e/poggadaj/cmd/poggadaj-tcp/pubdir"
	"codeberg.org/or3e/poggadaj/internal/security/argon2"
	"codeberg.org/or3e/poggadaj/internal/security/gg"
	"codeberg.org/or3e/poggadaj/internal/structs"
	"codeberg.org/or3e/poggadaj/internal/utils"
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
			log.L.Errorf("Failed to execute batch insert: %v\n", err)
		}
	}

	err := res.Close()
	if err != nil {
		log.L.Errorf("Failed to close batch results: %v\n", err)
	}
}

func (db *Database) GetUserList(uin uint32) []structs.UserListRequest {
	rows, err := db.conn.Query(context.Background(), "SELECT firstname, lastname, pseudonym, display_name, mobile_number, grp, uin, email, avail_sound, avail_path, msg_sound, msg_path, hidden, landline_number FROM ggcontact WHERE owner_uin=$1", uin)
	if err != nil {
		log.L.Errorf("Failed to execute query: %v\n", err)
	}
	defer rows.Close()

	var userList []structs.UserListRequest
	for rows.Next() {
		var user structs.UserListRequest
		err := rows.Scan(&user.FirstName, &user.LastName, &user.Pseudonym, &user.DisplayName, &user.MobileNumber, &user.Group, &user.UIN, &user.Email, &user.AvailSound, &user.AvailPath, &user.MsgSound, &user.MsgPath, &user.Hidden, &user.LandlineNumber)
		if err != nil {
			log.L.Errorf("Failed to scan row: %v\n", err)
		}
		userList = append(userList, user)
	}

	if rows.Err() != nil {
		log.L.Errorf("Failed to execute query: %v\n", rows.Err())
	}

	return userList
}

func (db *Database) DeleteUserList(uin uint32) error {
	_, err := db.conn.Exec(context.Background(), "DELETE FROM ggcontact WHERE owner_uin=$1", uin)
	return err
}

func (db *Database) GetPubdirDataByUin(uin uint32) (*pubdir.PubdirEntry, error) {
	entry := &pubdir.PubdirEntry{}
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

func (db *Database) WritePubdirData(uin uint32, entry *pubdir.PubdirEntry) error {
	_, err := db.conn.Exec(context.Background(),
		`INSERT INTO pubdir (uin, firstname, lastname, nickname, gender, birthyear, city, familyname, familycity)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (uin) DO UPDATE SET
		uin = $1, firstname = $2, lastname = $3, nickname = $4, gender = $5, birthyear = $6, city = $7, familyname = $8, familycity = $9`,
		uin, entry.Firstname, entry.Lastname, entry.Nickname, entry.Gender, entry.Birthyear, entry.City, entry.FamilyName, entry.FamilyCity,
	)
	return err
}

func (db *Database) SearchInPubdir(query *pubdir.PubdirEntry) ([]pubdir.PubdirEntry, uint32, error) {
	// TODO: Add support for only-online option

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
		result := pubdir.PubdirEntry{}
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

func (db *Database) GetAds(bannerType int) []Ad {
	query := fmt.Sprintf("SELECT adtype, bannertype, image, html FROM adserver_ad WHERE bannertype=%d", bannerType)
	ads := make([]Ad, 0)

	rows, err := db.conn.Query(context.Background(), query)
	if err != nil {
		fmt.Println(err)
		return ads
	}
	defer rows.Close()

	for rows.Next() {
		ad := Ad{}
		err := rows.Scan(&ad.AdType, &ad.BannerType, &ad.Image, &ad.Html)
		if err != nil {
			fmt.Println(err)
		}
		ads = append(ads, ad)
	}

	return ads
}

func (db *Database) CreateUser(regBody RegisterRequest) (int, error) {
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
		GGAncientHash = gg.GGAncientLoginHash(regBody.GGAncientPassword, utils.GetSeed())
		dbArgs["password_gg_ancient"] = GGAncientHash
	}
	if regBody.GG32Password != "" {
		GG32Hash = gg.GG32LoginHash(regBody.GG32Password, utils.GetSeed())
		dbArgs["password_gg32"] = GG32Hash
	}
	if regBody.GGSHA1Password != "" {
		GGSHA1Hash = gg.GGSHA1LoginHash(regBody.GGSHA1Password, utils.GetSeed())
		dbArgs["password_sha1"] = GGSHA1Hash
	}

	// Create the user
	query := "INSERT INTO gguser (name, password, password_gg_ancient, password_gg32, password_sha1) VALUES (@name, @password, @password_gg_ancient, @password_gg32, @password_sha1)"
	_, err2 := db.conn.Exec(context.Background(), query, dbArgs)
	if err2 != nil {
		return 0, err2
	}

	// Allocate a new UIN for the user
	var newUserUIN int
	query = "UPDATE gguser SET uin=nextval('uin_seq') WHERE name=$1 RETURNING uin"
	err3 := db.conn.QueryRow(context.Background(), query, regBody.Username).Scan(&newUserUIN)
	if err3 != nil {
		return 0, err3
	}

	return newUserUIN, nil
}

func (db *Database) GetUserPasswordHash(name string) (string, error) {
	query := "SELECT password FROM gguser WHERE name=$1"
	var passwordHash string
	err := db.conn.QueryRow(context.Background(), query, name).Scan(&passwordHash)
	if err != nil {
		return "", err
	}
	return passwordHash, nil
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

func (db *Database) UpdateAncientPassword(name string, password string) error {
	hashedPassword := gg.GGAncientLoginHash(password, utils.GetSeed())
	query := "UPDATE gguser SET password_gg_ancient=$1 WHERE name=$2"
	_, err := db.conn.Exec(context.Background(), query, hashedPassword, name)
	return err
}

func (db *Database) UpdateGG32Password(name string, password string) error {
	hashedPassword := gg.GG32LoginHash(password, utils.GetSeed())
	query := "UPDATE gguser SET password_gg32=$1 WHERE name=$2"
	_, err := db.conn.Exec(context.Background(), query, hashedPassword, name)
	return err
}

func (db *Database) UpdateSHA1Password(name string, password string) error {
	hashedPassword := gg.GGSHA1LoginHash(password, utils.GetSeed())
	query := "UPDATE gguser SET password_sha1=$1 WHERE name=$2"
	_, err := db.conn.Exec(context.Background(), query, hashedPassword, name)
	return err
}

func (db *Database) UpdateUserPassword(name string, chgreq ChangePasswordRequest) error {
	switch chgreq.PasswordType {
	case 0:
		// Website password
		return db.UpdateWebsitePassword(name, chgreq.Password)
	case 1:
		// Ancient password
		return db.UpdateAncientPassword(name, chgreq.Password)
	case 2:
		// GG32 password
		return db.UpdateGG32Password(name, chgreq.Password)
	case 3:
		return db.UpdateSHA1Password(name, chgreq.Password)
	default:
		return errs.WrongPasswordType{PasswordType: chgreq.PasswordType}
	}
}

func (db *Database) GetUserData(name string) (int, time.Time, error) {
	query := "SELECT uin, joined FROM gguser WHERE name=$1"
	var uin int
	var joined time.Time
	err := db.conn.QueryRow(context.Background(), query, name).Scan(&uin, &joined)
	return uin, joined, err
}

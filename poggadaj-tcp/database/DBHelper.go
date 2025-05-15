package database

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	log "poggadaj-shared/logging"
	"poggadaj-tcp/structs"
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

func PutUserList(userList []structs.UserListRequest, uin uint32) {
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
	res := DatabaseConn.SendBatch(context.Background(), batch)

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

func GetUserList(uin uint32) []structs.UserListRequest {
	rows, err := DatabaseConn.Query(context.Background(), "SELECT firstname, lastname, pseudonym, display_name, mobile_number, grp, uin, email, avail_sound, avail_path, msg_sound, msg_path, hidden, landline_number FROM ggcontact WHERE owner_uin=$1", uin)
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

func DeleteUserList(uin uint32) error {
	_, err := DatabaseConn.Exec(context.Background(), "DELETE FROM ggcontact WHERE owner_uin=$1", uin)
	return err
}

var DatabaseConn *pgxpool.Pool

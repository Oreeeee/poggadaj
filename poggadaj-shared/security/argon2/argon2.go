package argon2

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"golang.org/x/crypto/argon2"
	"strings"
)

func HashPassword(password string) (string, error) {
	salt := make([]byte, ARGON2_SALT_LEN)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}
	passwordHash := argon2.IDKey([]byte(password), salt, ARGON2_TIME, ARGON2_MEMORY, ARGON2_THREADS, ARGON2_KEY_LEN)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(passwordHash)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version,
		ARGON2_MEMORY, ARGON2_TIME, ARGON2_THREADS, b64Salt, b64Hash), nil
}

func ComparePasswords(password string, hash string) (bool, error) {
	vals := strings.Split(hash, "$")
	if len(vals) != 6 {
		return false, errors.New("invalid hash")
	}

	var version int
	_, err := fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return false, err
	}
	if version != argon2.Version {
		return false, errors.New("wrong argon version")
	}

	var mem, iter uint32
	var thread uint8
	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &mem, &iter, &thread)
	if err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(vals[4])
	if err != nil {
		return false, err
	}

	actualPasswordHash, err := base64.RawStdEncoding.DecodeString(vals[5])
	if err != nil {
		return false, err
	}

	providedPasswordHash := argon2.IDKey([]byte(password), salt, iter, mem, thread,
		uint32(len(actualPasswordHash)))
	return subtle.ConstantTimeCompare(actualPasswordHash, providedPasswordHash) == 1, nil
}

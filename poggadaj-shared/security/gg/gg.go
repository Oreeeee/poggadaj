// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2025 Oreeeee

package gg

import (
	"crypto/sha1"
	"encoding/hex"
)

func GGAncientLoginHash(password string, seed uint32) uint32 {
	var hash uint32 = 1
	for _, char := range password {
		hash *= uint32(char) + 1
	}
	hash *= seed
	return hash
}

func GG32LoginHash(password string, seed uint32) uint32 {
	var x, y, z uint32

	y = seed

	for i := 0; i < len(password); i++ {
		x = (x & 0xffffff00) | uint32(password[i])
		y ^= x
		y += x
		x <<= 8
		y ^= x
		x <<= 8
		y -= x
		x <<= 8
		y ^= x

		z = y & 0x1f
		y = (y << z) | (y >> (32 - z))
	}

	return y
}

func GGSHA1LoginHash(password string, seed uint32) string {
	hasher := sha1.New()

	hasher.Write([]byte(password))

	seedBytes := []byte{
		byte(seed),
		byte(seed >> 8),
		byte(seed >> 16),
		byte(seed >> 24),
	}
	hasher.Write(seedBytes)

	hash := hasher.Sum(nil)
	return hex.EncodeToString(hash)
}

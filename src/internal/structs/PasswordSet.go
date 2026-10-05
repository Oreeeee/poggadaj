// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package structs

// PasswordSet contains all password types for a user (Ancient, GG32, SHA1, Argon2)
type PasswordSet struct {
	Ancient uint32
	GG32    uint32
	SHA1    string
	Argon2  string
}

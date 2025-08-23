// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2025 Oreeeee

package main

type RegisterRequest struct {
	Username          string `json:"username"`
	Password          string `json:"password"`
	GGAncientPassword string `json:"gg_ancient_password"`
	GG32Password      string `json:"gg32_password"`
	GGSHA1Password    string `json:"gg_sha1_password"`
}

type ChangePasswordRequest struct {
	PasswordType int    `json:"password_type"`
	Password     string `json:"password"`
}

// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2025 Oreeeee

package main

import "time"

type RegisterResponse struct {
	Error string `json:"error"`
	UIN   int    `json:"uin"`
}

type UserDataResponse struct {
	UIN    int       `json:"uin"`
	Joined time.Time `json:"joined"`
}

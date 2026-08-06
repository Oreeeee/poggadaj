// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package main

import (
	"os"

	"codeberg.org/or3e/poggadaj/internal/database"
)

var Sessions []AuthorizedSession

func main() {
	dbCfg := &database.DatabaseConfig{
		Host:     os.Getenv("DB_ADDRESS"),
		Port:     "5432",
		Username: "poggadaj",
		Password: os.Getenv("DB_PASSWORD"),
	}

	server, _ := NewServer(dbCfg)
	server.Run()
}

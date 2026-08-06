// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package main

import (
	"fmt"
	"os"
	"time"

	"codeberg.org/or3e/poggadaj/internal/cache"
	"codeberg.org/or3e/poggadaj/internal/database"
	"codeberg.org/or3e/poggadaj/internal/logging"

	"charm.land/log/v2"
)

func main() {
	dbCfg := &database.DatabaseConfig{
		Host:     os.Getenv("DB_ADDRESS"),
		Port:     "5432",
		Username: "poggadaj",
		Password: os.Getenv("DB_PASSWORD"),
	}

	cache, _ := cache.NewCache()

	logging.L = log.NewWithOptions(os.Stdout, log.Options{
		ReportCaller:    true,
		ReportTimestamp: true,
		TimeFormat:      time.DateTime,
		Level:           log.DebugLevel,
	})

	server, err := NewServer(dbCfg, cache, fmt.Sprintf("%s:8074", os.Getenv("LISTEN_ADDRESS")))
	if err != nil {
		panic(err)
	}
	server.Run()
}

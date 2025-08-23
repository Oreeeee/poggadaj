// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2025 Oreeeee

package main

import "database/sql"

type Ad struct {
	AdType     int
	BannerType int
	Image      sql.NullString
	Html       sql.NullString
}

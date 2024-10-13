package main

import "database/sql"

type Ad struct {
	AdType     int
	BannerType int
	Image      sql.NullString
	Html       sql.NullString
}

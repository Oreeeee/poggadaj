module poggadaj_http

go 1.26

replace poggadaj-shared => ./shared

require (
	github.com/jackc/pgx/v5 v5.9.1
	github.com/labstack/echo/v5 v5.1.0
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	golang.org/x/net v0.52.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/text v0.36.0 // indirect
)

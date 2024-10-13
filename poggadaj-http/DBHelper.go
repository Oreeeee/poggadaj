package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
)

func GetDBConn() (*pgxpool.Pool, error) {
	dbaddr := os.Getenv("DB_ADDRESS")
	password := os.Getenv("DB_PASSWORD")
	return pgxpool.New(
		context.Background(),
		fmt.Sprintf(
			"user=poggadaj password=%s host=%s port=5432 dbname=poggadaj sslmode=disable",
			password,
			dbaddr,
		),
	)
}

func GetAds(bannerType int) []Ad {
	query := fmt.Sprintf("SELECT adtype, bannertype, image, html FROM adserver_ad WHERE bannertype=%d", bannerType)
	ads := make([]Ad, 0)

	rows, err := DatabaseConn.Query(context.Background(), query)
	if err != nil {
		fmt.Println(err)
		return ads
	}
	defer rows.Close()

	for rows.Next() {
		ad := Ad{}
		err := rows.Scan(&ad.AdType, &ad.BannerType, &ad.Image, &ad.Html)
		if err != nil {
			fmt.Println(err)
		}
		ads = append(ads, ad)
	}

	return ads
}

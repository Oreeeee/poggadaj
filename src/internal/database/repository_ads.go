// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package database

import (
	"context"
	"fmt"

	"charm.land/log/v2"

	"codeberg.org/or3e/poggadaj/internal/structs"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AdsRepository struct {
	conn   *pgxpool.Pool
	logger *log.Logger
}

func (r *AdsRepository) GetAds(bannerType int) []structs.Ad {
	query := fmt.Sprintf("SELECT adtype, bannertype, image, html FROM adserver_ad WHERE bannertype=%d", bannerType)
	ads := make([]structs.Ad, 0)

	rows, err := r.conn.Query(context.Background(), query)
	if err != nil {
		fmt.Println(err)
		return ads
	}
	defer rows.Close()

	for rows.Next() {
		ad := structs.Ad{}
		err := rows.Scan(&ad.AdType, &ad.BannerType, &ad.Image, &ad.Html)
		if err != nil {
			fmt.Println(err)
		}
		ads = append(ads, ad)
	}

	return ads
}

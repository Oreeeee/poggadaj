// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
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

func GetClients() ([]HtmlClient, error) {
	query := `SELECT cd.name, cd.image_url, cd.installer_download_url, cd.extracted_download_url FROM client_downloads AS cd
	JOIN client_downloads_descriptions AS ds ON ds.`
	rows, err := DatabaseConn.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	clients := []HtmlClient{}
	for rows.Next() {
		client := HtmlClient{}
		err := rows.Scan(&client.Name, &client.Description, &client.ImageUrl, &client.InstallerDownloadUrl, &client.ExtractedDownloadUrl)
		if err != nil {
			return nil, err
		}
		clients = append(clients, client)
	}

	if rows.Err() != nil {
		return nil, err
	}

	return clients, nil
}

var DatabaseConn *pgxpool.Pool

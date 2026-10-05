// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package database

import (
	"context"
	"maps"
	"slices"

	"charm.land/log/v2"

	"codeberg.org/or3e/poggadaj/internal/structs"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ClientsRepository struct {
	conn   *pgxpool.Pool
	logger *log.Logger
}

func (r *ClientsRepository) GetClients(language string) ([]*structs.WebClient, error) {
	clients := []*structs.WebClient{}
	clientsById := map[int]*structs.WebClient{}

	rows, err := r.conn.Query(
		context.Background(),
		"SELECT c.id, c.name, c.image_url, d.description FROM clients c JOIN client_descriptions d ON c.id = d.client_id WHERE d.language = $1 ORDER BY c.priority ASC",
		language,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		client := &structs.WebClient{}

		err = rows.Scan(
			&client.Id,
			&client.Name,
			&client.ImageUrl,
			&client.Description,
		)

		if err != nil {
			return nil, err
		}

		clients = append(clients, client)
		clientsById[client.Id] = client
	}

	// Query the download links and add to the clients
	keys := slices.Collect(maps.Keys(clientsById))
	downloads, err := r.GetClientsDownloads(keys)
	if err != nil {
		return nil, err
	}

	for _, v := range slices.Collect(maps.Keys(downloads)) {
		for _, downloadsEntry := range downloads[v] {
			clientsById[v].Downloads = append(clientsById[v].Downloads, downloadsEntry)
		}
	}

	return clients, nil
}

func (r *ClientsRepository) GetClientsDownloads(clientIds []int) (map[int][]*structs.WebClientDownload, error) {
	downloads := map[int][]*structs.WebClientDownload{}
	if len(clientIds) == 0 {
		// TODO: Would be nice to return a custom error here?
		return downloads, nil
	}

	rows, err := r.conn.Query(
		context.Background(),
		"SELECT client_id, file_variant, url FROM client_downloads WHERE client_id = ANY($1) ORDER BY file_variant ASC",
		clientIds,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		clientId := 0
		downloadEntry := &structs.WebClientDownload{}

		err = rows.Scan(
			&clientId,
			&downloadEntry.Type,
			&downloadEntry.Url,
		)
		if err != nil {
			return nil, err
		}

		if _, found := downloads[clientId]; !found {
			downloads[clientId] = []*structs.WebClientDownload{}
		}

		downloads[clientId] = append(downloads[clientId], downloadEntry)
	}

	return downloads, nil
}

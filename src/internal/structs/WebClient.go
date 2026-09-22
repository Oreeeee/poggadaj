// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package structs

type DownloadType uint8

const (
	INSTALLER DownloadType = iota
	UNPACKED
)

type WebClient struct {
	Id          int
	Name        string
	ImageUrl    string
	Description string
	Downloads   []*WebClientDownload
}

type WebClientDownload struct {
	Type DownloadType
	Url  string
}

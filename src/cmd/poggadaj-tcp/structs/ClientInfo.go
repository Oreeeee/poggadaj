// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package structs

import (
	"net"

	uv "codeberg.org/or3e/poggadaj/cmd/poggadaj-tcp/universal"
)

type ClientInfo struct {
	Conn          net.Conn
	UIN           uint32
	Status        uint32
	Authenticated bool
	NotifyList    []uv.GG_NotifyContact
}

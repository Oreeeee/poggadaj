// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package protocol

import "codeberg.org/or3e/poggadaj/internal/utils"

type GG_Packet_Iface interface {
	Serialize(*utils.IOStream)
	Deserialize(*utils.IOStream)
}

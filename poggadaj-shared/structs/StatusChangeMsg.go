// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2025 Oreeeee

package structs

type StatusChangeMsg struct {
	UIN         uint32 `json:"uin"`
	Status      uint32 `json:"status"`
	Description []byte `json:"description"`
}

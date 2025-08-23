// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2025 Oreeeee

package structs

type Message struct {
	From     uint32 `json:"from"`
	MsgClass uint32 `json:"msg_class"`
	Content  []byte `json:"content"`
}

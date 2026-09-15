// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package structs

import "time"

type UserData struct {
	UIN         uint32
	WebUsername string
	JoinedDate  time.Time
}

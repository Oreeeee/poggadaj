// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2025 Oreeeee

package utils

import "fmt"

func StringifySHA1(hash [64]byte) string {
	s := fmt.Sprintf("%x", hash[:20])
	return s
}

// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package logging

import "charm.land/log/v2"

func StructPPrint(logger *log.Logger, packetName string, packetLines []string) {
	logger.Debug(packetName)
	for _, line := range packetLines {
		logger.Debug(line)
	}
}

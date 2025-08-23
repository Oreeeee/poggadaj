// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2025 Oreeeee

package logging

func StructPPrint(packetName string, packetLines []string) {
	L.Debugf(packetName)
	for _, line := range packetLines {
		L.Debugf(line)
	}
}

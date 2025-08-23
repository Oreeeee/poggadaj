// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2025 Oreeeee

package main

import (
	"poggadaj-shared/cache"
	"poggadaj-shared/logging"
	"poggadaj-shared/statuses"
	"poggadaj-tcp/clients"
)

func MsgChannel(c *clients.GGClient, run *bool) {
	defer logging.L.Debugf("Quitting message channel")
	pubsub := cache.GetMessageChannel(c.UIN)
	for *run {
		msg := cache.RecvMessageChannel(pubsub)
		if !*run {
			// Sanity check to not accidentally write to a closed socket
			continue
		}

		logging.L.Debugf("%d received a message!", c.UIN)
		c.SendRecvMsg(msg)
	}
}

func StatusChannel(c *clients.GGClient, run *bool) {
	defer logging.L.Debugf("Quitting status channel")
	pubsub := cache.GetStatusChannel()
	for *run {
		statusChange := cache.RecvStatusChannel(pubsub)
		if !*run {
			// Sanity check to not accidentally write to a closed socket
			continue
		}

		// Check if the status change is applicable for this connection
		for _, e := range c.NotifyList {
			if e.UIN == statusChange.UIN {
				logging.L.Debugf("%d's status change is relevant for %d", statusChange.UIN, c.UIN)

				switch statusChange.Status {
				case statuses.GG_STATUS_INVISIBLE:
					logging.L.Debugf("Got GG_STATUS_INVISIBLE, sending GG_STATUS_NOT_AVAIL")
					statusChange.Status = statuses.GG_STATUS_NOT_AVAIL
				case statuses.GG_STATUS_INVISIBLE_DESCR:
					logging.L.Debugf("Got GG_STATUS_INVISIBLE_DESCR, sending GG_STATUS_NOT_AVAIL_DESCR")
					statusChange.Status = statuses.GG_STATUS_NOT_AVAIL_DESCR
				}

				c.SendStatus(statusChange)
			}
		}
	}
}

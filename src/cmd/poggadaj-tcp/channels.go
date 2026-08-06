// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package main

import (
	"codeberg.org/or3e/poggadaj/internal/statuses"
)

func (c *Client) MsgChannel(run *bool) {
	defer c.logger.Debugf("Quitting message channel")
	pubsub := c.server.cache.GetMessageChannel(c.UIN)
	for *run {
		msg := c.server.cache.RecvMessageChannel(pubsub)
		if !*run {
			// Sanity check to not accidentally write to a closed socket
			continue
		}

		c.logger.Debugf("%d received a message!", c.UIN)
		c.SendRecvMsg(msg)
	}
}

func (c *Client) StatusChannel(run *bool) {
	defer c.logger.Debugf("Quitting status channel")
	pubsub := c.server.cache.GetStatusChannel()
	for *run {
		statusChange := c.server.cache.RecvStatusChannel(pubsub)
		if !*run {
			// Sanity check to not accidentally write to a closed socket
			continue
		}

		// Check if the status change is applicable for this connection
		for _, e := range c.NotifyList {
			if e.UIN == statusChange.UIN {
				c.logger.Debugf("%d's status change is relevant for %d", statusChange.UIN, c.UIN)

				switch statusChange.Status {
				case statuses.GG_STATUS_INVISIBLE:
					c.logger.Debugf("Got GG_STATUS_INVISIBLE, sending GG_STATUS_NOT_AVAIL")
					statusChange.Status = statuses.GG_STATUS_NOT_AVAIL
				case statuses.GG_STATUS_INVISIBLE_DESCR:
					c.logger.Debugf("Got GG_STATUS_INVISIBLE_DESCR, sending GG_STATUS_NOT_AVAIL_DESCR")
					statusChange.Status = statuses.GG_STATUS_NOT_AVAIL_DESCR
				}

				c.SendStatus(statusChange)
			}
		}
	}
}

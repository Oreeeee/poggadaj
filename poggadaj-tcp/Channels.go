package main

import (
	"poggadaj-tcp/clients"
	"poggadaj-tcp/database"
	"poggadaj-tcp/logging"
	"poggadaj-tcp/universal"
)

func MsgChannel(c *clients.GGClient, run *bool) {
	defer logging.L.Debugf("Quitting message channel")
	pubsub := database.GetMessageChannel(c.UIN)
	for *run {
		msg := database.RecvMessageChannel(pubsub)
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
	pubsub := database.GetStatusChannel()
	for *run {
		statusChange := database.RecvStatusChannel(pubsub)
		if !*run {
			// Sanity check to not accidentally write to a closed socket
			continue
		}

		// Check if the status change is applicable for this connection
		for _, e := range c.NotifyList {
			if e.UIN == statusChange.UIN {
				logging.L.Debugf("%d's status change is relevant for %d", statusChange.UIN, c.UIN)

				switch statusChange.Status {
				case universal.GG_STATUS_INVISIBLE:
					logging.L.Debugf("Got GG_STATUS_INVISIBLE, sending GG_STATUS_NOT_AVAIL")
					statusChange.Status = universal.GG_STATUS_NOT_AVAIL
				case universal.GG_STATUS_INVISIBLE_DESCR:
					logging.L.Debugf("Got GG_STATUS_INVISIBLE_DESCR, sending GG_STATUS_NOT_AVAIL_DESCR")
					statusChange.Status = universal.GG_STATUS_NOT_AVAIL_DESCR
				}

				c.SendStatus(statusChange)
			}
		}
	}
}

package main

import (
	"poggadaj-tcp/database"
	"poggadaj-tcp/generichandlers"
	"poggadaj-tcp/logging"
	"poggadaj-tcp/structs"
	"poggadaj-tcp/universal"
)

func MsgChannel(c generichandlers.GGClient, cI *structs.ClientInfo, run *bool) {
	defer logging.L.Debugf("Quitting message channel")
	pubsub := database.GetMessageChannel(cI.UIN)
	for *run {
		msg := database.RecvMessageChannel(pubsub)
		if !*run {
			// Sanity check to not accidentally write to a closed socket
			continue
		}

		logging.L.Debugf("%d received a message!", cI.UIN)
		c.SendRecvMsg(msg)
	}
}

func StatusChannel(c generichandlers.GGClient, cI *structs.ClientInfo, run *bool) {
	defer logging.L.Debugf("Quitting status channel")
	pubsub := database.GetStatusChannel()
	for *run {
		statusChange := database.RecvStatusChannel(pubsub)
		if !*run {
			// Sanity check to not accidentally write to a closed socket
			continue
		}

		// Check if the status change is applicable for this connection
		for _, e := range cI.NotifyList {
			if e.UIN == statusChange.UIN {
				logging.L.Debugf("%d's status change is relevant for %d", statusChange.UIN, cI.UIN)

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

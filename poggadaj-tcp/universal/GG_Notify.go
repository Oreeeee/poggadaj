// SPDX-License-Identifier: AGPL-3.0-or-later
// SPDX-FileCopyrightText: 2024-2026 Oreeeee

package universal

import (
	"poggadaj-tcp/constants"
	"poggadaj-tcp/utils"
)

func GG_NotifyContactDeserialize(stream *utils.IOStream, contactList *[]GG_NotifyContact) {
	contactListLen := int(stream.Available() / constants.GG_NOTIFYCONTACT_SIZE)
	for i := 0; i < contactListLen; i++ {
		contact := GG_NotifyContact{
			UIN:  stream.ReadU32(),
			Type: stream.ReadU8(),
		}
		*contactList = append(*contactList, contact)
	}
}

func GG_AddNotify(stream *utils.IOStream, contactList *[]GG_NotifyContact) GG_NotifyContact {
	contact := GG_NotifyContact{
		UIN:  stream.ReadU32(),
		Type: stream.ReadU8(),
	}
	*contactList = append(*contactList, contact)
	return contact
}

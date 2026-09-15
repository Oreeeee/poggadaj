package structs

import "time"

type UserData struct {
	UIN         uint32
	WebUsername string
	JoinedDate  time.Time
}

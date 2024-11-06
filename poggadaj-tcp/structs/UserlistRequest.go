package structs

import (
	"fmt"
	"poggadaj-tcp/utils"
	"strconv"
	"strings"
)

type UserListRequest struct {
	FirstName      string
	LastName       string
	Pseudonym      string
	DisplayName    string
	MobileNumber   uint32
	Group          string
	UIN            uint32
	Email          string
	AvailSound     uint8
	AvailPath      string
	MsgSound       uint8
	MsgPath        string
	Hidden         bool
	LandlineNumber uint32
}

func (l *UserListRequest) Read(data string) error {
	splittedData := strings.Split(data, ";")

	if len(splittedData) != 14 {
		// TODO: Add a custom error
		return nil
	}

	l.FirstName = splittedData[0]
	l.LastName = splittedData[1]
	l.Pseudonym = splittedData[2]
	l.DisplayName = splittedData[3]

	mobileNumberTmp, _ := strconv.ParseUint(splittedData[4], 10, 32)
	l.MobileNumber = uint32(mobileNumberTmp)

	l.Group = splittedData[5]

	uinTmp, _ := strconv.ParseUint(splittedData[6], 10, 32)
	l.UIN = uint32(uinTmp)

	l.Email = splittedData[7]

	availSoundTmp, _ := strconv.ParseUint(splittedData[8], 10, 8)
	l.AvailSound = uint8(availSoundTmp)

	l.AvailPath = splittedData[9]

	msgSoundTmp, _ := strconv.ParseUint(splittedData[10], 10, 8)
	l.MsgSound = uint8(msgSoundTmp)
	
	l.MsgPath = splittedData[11]

	hidden, _ := strconv.ParseBool(splittedData[12])
	l.Hidden = hidden

	landlineNumberTmp, _ := strconv.ParseUint(splittedData[13], 10, 32)
	l.LandlineNumber = uint32(landlineNumberTmp)

	return nil
}

func (l *UserListRequest) Write() string {
	return fmt.Sprintf(
		"%s;%s;%s;%s;%d;%s;%d;%s;%d;%s;%d;%s;%d;%d",
		l.FirstName,
		l.LastName,
		l.Pseudonym,
		l.DisplayName,
		l.MobileNumber,
		l.Group,
		l.UIN,
		l.Email,
		l.AvailSound,
		l.AvailPath,
		l.MsgSound,
		l.MsgPath,
		utils.BoolToInt(l.Hidden),
		l.LandlineNumber,
	)
}

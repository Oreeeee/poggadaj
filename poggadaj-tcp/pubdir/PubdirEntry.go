package pubdir

import (
	"bytes"
	"errors"
	"fmt"
	"poggadaj-tcp/utils"
	"reflect"
	"strconv"
	"strings"
)

type PubdirEntry struct {
	UIN        uint32
	Firstname  string
	Lastname   string
	Nickname   string
	Birthyear  uint16
	City       string
	Gender     uint8
	ActiveOnly bool
	FamilyName string
	FamilyCity string
	Status     uint8
}

func (p *PubdirEntry) Read(data []byte) error {
	strData := string(data)
	splitData := strings.Split(strData, "\x00")

	// Pubdir strings are key-value pairs, length of them should be even
	if len(splitData)%2 != 0 {
		return errors.New("length of string is not even")
	}

	// Parse the key-value pairs, as seen in the documentation
	for i := 0; i < len(splitData); i += 2 {
		v := splitData[i+1]
		switch splitData[i] {
		case "FmNumber":
			uin, _ := strconv.ParseUint(v, 10, 32)
			p.UIN = uint32(uin)
		case "firstname":
			p.Firstname = v
		case "lastname":
			p.Lastname = v
		case "nickname":
			p.Nickname = v
		case "birthyear":
			birthyear, _ := strconv.ParseUint(v, 10, 16)
			p.Birthyear = uint16(birthyear)
		case "city":
			p.City = v
		case "gender":
			gender, _ := strconv.ParseUint(v, 10, 8)
			p.Gender = uint8(gender)
		case "ActiveOnly":
			if v == "1" {
				p.ActiveOnly = true
			} else {
				p.ActiveOnly = false
			}
		case "familyname":
			p.FamilyName = v
		case "familycity":
			p.FamilyCity = v
		case "FmStatus":
			status, _ := strconv.ParseUint(v, 10, 8)
			p.Status = uint8(status)
		}
	}

	return nil
}

func (p *PubdirEntry) Write() []byte {
	var resultBuilder bytes.Buffer

	structFields := reflect.VisibleFields(reflect.TypeOf(p))
	for _, field := range structFields {
		switch field.Name {
		case "UIN":
			resultBuilder.WriteString(fmt.Sprintf("FmNumber\x00%d", p.UIN))
		case "Firstname":
			resultBuilder.WriteString("firstname\x00" + p.Firstname)
		case "Lastname":
			resultBuilder.WriteString("lastname\x00" + p.Lastname)
		case "Nickname":
			resultBuilder.WriteString("nickname\x00" + p.Nickname)
		case "Birthyear":
			resultBuilder.WriteString(fmt.Sprintf("birthyear\x00%d", p.Birthyear))
		case "City":
			resultBuilder.WriteString("city\x00" + p.City)
		case "gender":
			resultBuilder.WriteString(fmt.Sprintf("gender\x00%d", p.Gender))
		case "ActiveOnly":
			resultBuilder.WriteString(fmt.Sprintf("ActiveOnly\x00%d", utils.BoolToInt(p.ActiveOnly)))
		case "FamilyName":
			resultBuilder.WriteString("familyname\x00" + p.FamilyName)
		case "FamilyCity":
			resultBuilder.WriteString("familycity\x00" + p.FamilyCity)
		case "FmStatus":
			resultBuilder.WriteString(fmt.Sprintf("FmStatus\x00%d", p.Status))
		}

		// Append null byte at the end
		resultBuilder.WriteByte(0x00)
	}

	return resultBuilder.Bytes()
}

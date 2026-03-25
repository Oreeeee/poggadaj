package pubdir

import (
	"bytes"
	"fmt"
	"poggadaj-shared/logging"
	"poggadaj-tcp/utils"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type PubdirEntry struct {
	UIN          uint32
	Firstname    string
	Lastname     string
	Nickname     string
	MinBirthyear uint16 // Only used when YearIsRange is true
	Birthyear    uint16
	YearIsRange  bool // Indicates whether MinBirthyear is used (in lookup queries only)
	City         string
	Gender       uint8
	ActiveOnly   bool
	FamilyName   string
	FamilyCity   string
	Status       uint8
	Start        uint32
}

func (p *PubdirEntry) Read(data []byte) error {
	strData := string(data)
	splitData := strings.Split(strData, "\x00")

	// Pubdir strings are key-value pairs, length of them should be even
	//if len(splitData)%2 != 0 {
	//	return errors.New("length of string is not even")
	//}

	// Parse the key-value pairs, as seen in the documentation
	// Gadu-Gadu client writes an additional 0x00 at the end, and it would cause index out of range panics, hence the -1
	for i := 0; i < len(splitData)-1; i += 2 {
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
			// Handle age ranges
			// If the age is a range, then the birthyear field contains a space
			// which separates the min and max birth year
			// Example: 1999 2001
			if strings.ContainsRune(v, ' ') {
				p.YearIsRange = true
				years := strings.Split(v, " ")

				if len(years) != 2 {
					logging.L.Errorf("Invalid age range provided: %v", v)
					p.YearIsRange = false
					continue
					//return errors.New("invalid age range provided")
				}

				minyear, _ := strconv.ParseUint(years[0], 10, 16)
				maxyear, _ := strconv.ParseUint(years[1], 10, 16)

				p.MinBirthyear = uint16(minyear)
				p.Birthyear = uint16(maxyear)
				continue
			}

			birthyear, _ := strconv.ParseUint(v, 10, 16)
			p.Birthyear = uint16(birthyear)
			p.MinBirthyear = p.Birthyear
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
		case "fmstart":
			start, _ := strconv.ParseUint(v, 10, 32)
			p.Start = uint32(start)
		}
	}

	return nil
}

func (p *PubdirEntry) Write() []byte {
	var resultBuilder bytes.Buffer

	writer := transform.NewWriter(&resultBuilder, charmap.Windows1250.NewEncoder())
	defer writer.Close()

	WriteSingleParam(writer, "FmNumber", p.UIN)
	WriteSingleParam(writer, "firstname", p.Firstname)
	WriteSingleParam(writer, "lastname", p.Lastname)
	WriteSingleParam(writer, "nickname", p.Nickname)
	WriteSingleParam(writer, "birthyear", p.Birthyear)
	WriteSingleParam(writer, "city", p.City)
	WriteSingleParam(writer, "gender", p.Gender)
	WriteSingleParam(writer, "ActiveOnly", utils.BoolToInt(p.ActiveOnly))
	WriteSingleParam(writer, "familyname", p.FamilyName)
	WriteSingleParam(writer, "familycity", p.FamilyCity)
	WriteSingleParam(writer, "FmStatus", p.Status)

	return resultBuilder.Bytes()
}

// WriteSingleParam writes a param in PubdirEntry while skipping empty strings
func WriteSingleParam(builder *transform.Writer, key string, value any) {
	if value == "" {
		return
	}

	fmt.Fprintf(builder, "%s\x00%v\x00", key, value)
}

// PubdirWriteRange serializes a slice of PubdirEntries
func PubdirWriteRange(entries []PubdirEntry) []byte {
	var resultBuilder bytes.Buffer
	lastIndex := len(entries) - 1

	for idx, v := range entries {
		resultBuilder.Write(v.Write())

		if idx != lastIndex {
			resultBuilder.WriteByte('\x00')
		}
	}

	return resultBuilder.Bytes()
}

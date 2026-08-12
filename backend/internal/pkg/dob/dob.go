package dob

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

var ErrInvalidDobString = errors.New("dob: Invalid date of birth string")
var ErrImpossibleDob = errors.New("dob: Impossible date of birth")

type Dob struct {
	Year  uint16
	Month uint8
	Day   uint8
}

// year-month-date
//
// 2000-12-30
func NewDobFromString(dobString string) (Dob, error) {
	dobSpl := strings.Split(dobString, "-")

	if len(dobSpl) != 3 {
		return Dob{}, ErrInvalidDobString
	}

	year, err := strconv.Atoi(dobSpl[0])
	month, err := strconv.Atoi(dobSpl[1])
	day, err := strconv.Atoi(dobSpl[2])

	if err != nil {
		return Dob{}, ErrInvalidDobString
	}

	if year > time.Now().Year() || month > 12 || day > 31 {
		return Dob{}, ErrImpossibleDob
	}

	dob := Dob{
		Year:  uint16(year),
		Month: uint8(month),
		Day:   uint8(day),
	}

	return dob, nil
}

// Calculates age from dob instance
func (d Dob) Age() uint8 {
	now := time.Now()
	year := uint16(now.Year())
	day := uint8(now.Day())
	month := uint8(now.Month())

	age := uint8(year - d.Year)

	if month < d.Month ||
		(month == d.Month && day < d.Day) {
		age--
	}
	return age
}

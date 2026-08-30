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
	Year  int
	Month int
	Day   int
}

// year-month-date
//
// 2000-12-30
func NewDobFromString(dobString string) (Dob, error) {
	dobSpl := strings.Split(dobString, "-")

	if len(dobSpl) != 3 {
		return Dob{}, ErrInvalidDobString
	}

	var year, month, day int
	var err error
	year, err = strconv.Atoi(dobSpl[0])
	if err != nil {
		return Dob{}, ErrInvalidDobString
	}
	month, err = strconv.Atoi(dobSpl[1])
	if err != nil {
		return Dob{}, ErrInvalidDobString
	}
	day, err = strconv.Atoi(dobSpl[2])
	if err != nil {
		return Dob{}, ErrInvalidDobString
	}

	if year > time.Now().Year() || month > 12 || day > 31 {
		return Dob{}, ErrImpossibleDob
	}

	dob := Dob{
		Year:  year,
		Month: month,
		Day:   day,
	}

	return dob, nil
}

// Calculates age from dob instance
func (d Dob) Age() int {
	now := time.Now()
	year := now.Year()
	day := now.Day()
	month := int(now.Month())

	age := year - d.Year
	if month < d.Month ||
		(month == d.Month && day < d.Day) {
		age--
	}
	return age
}

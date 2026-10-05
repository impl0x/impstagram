package dob

import (
	"errors"
	"strconv"
	"time"
)

type Dob struct {
	Year  uint16
	Month uint16
	Day   uint16
}

// same as parse but panics if parsing fails
func MustParse(dobString string) Dob {
	d, err := Parse(dobString)
	if err != nil {
		panic("dob: fail to parse dob string")
	}
	return d
}

// Dob errors
var (
	ErrEmpty           = errors.New("dob: empty")
	ErrIncorrectFormat = errors.New("dob: wrong format, format must be YYYY-MM-DD")
	ErrNotParseable    = errors.New("dob: not parseable")
	ErrImpossible      = errors.New("dob: dob is impossible")
)

// format: year-month-day, YYYY-MM-DD. Strict.
//
// example: 2000-12-30, 2005-01-03
//
// possible errors: [ErrEmpty], [ErrIncorrectFormat], [ErrNotParseable], [ErrImpossible]
func Parse(s string) (Dob, error) {
	if s == "" {
		return Dob{}, ErrEmpty
	}
	if len(s) != 10 || s[4] != '-' || s[7] != '-' {
		return Dob{}, ErrIncorrectFormat
	}
	y, err := strconv.Atoi(s[:4])
	if err != nil {
		return Dob{}, ErrNotParseable
	}
	if y > time.Now().Year() || y < 1900 { // 1900 is the lower limit, the oldest living person as of now was born on 1909, so we can safely put this.
		return Dob{}, ErrImpossible
	}
	m, err := strconv.Atoi(s[5:7])
	if err != nil {
		return Dob{}, ErrNotParseable
	}
	if m > 12 || m <= 0 {
		return Dob{}, ErrImpossible
	}
	d, err := strconv.Atoi(s[8:])
	if d > 31 || d <= 0 {
		return Dob{}, ErrImpossible
	}
	return Dob{uint16(y), uint16(m), uint16(d)}, nil

}

// Calculates age from dob instance
func (d Dob) Age() uint16 {
	now := time.Now()
	year := uint16(now.Year()) // we are not reaching 65,536 years in the future for this function to fail
	day := uint16(now.Day())
	month := uint16(now.Month())

	age := year - d.Year
	if month < d.Month ||
		(month == d.Month && day < d.Day) {
		age--
	}
	return age
}

func (d Dob) String() string {
	return strconv.Itoa(int(d.Year)) + "-" + strconv.Itoa(int(d.Month)) + "-" + strconv.Itoa(int(d.Day))
}

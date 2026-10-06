package dob

import (
	"errors"
	"strconv"
	"time"
)

type Dob struct {
	Year  int
	Month int
	Day   int
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
	return Dob{y, m, d}, nil

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

// Converts the Dob instance to string using the format defined
func (d Dob) String() string {
	return strconv.Itoa(d.Year) + "-" + strconv.Itoa(d.Month) + "-" + strconv.Itoa(d.Day)
}

package dob

import (
	"errors"
	"strconv"
	"time"
)

type Dob struct {
	Year  uint
	Month uint
	Day   uint
}

// Dob errors
var (
	ErrEmpty           = errors.New("dob: empty")
	ErrIncorrectFormat = errors.New("dob: wrong format, format must be YYYY-MM-DD")
	ErrNotParseable    = errors.New("dob: not parseable")
	ErrImpossible      = errors.New("dob: dob is impossible")
)

var possibleMaxMonthDays = [12]uint{31, 29, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}

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
	yt, err := strconv.ParseUint(s[:4], 10, 64)
	y := uint(yt)
	if err != nil {
		return Dob{}, ErrNotParseable
	}
	if y > uint(time.Now().Year()) || y < 1900 { // 1900 is the lower limit, the oldest living person as of now was born on 1909, so we can safely put this.
		return Dob{}, ErrImpossible
	}
	mt, err := strconv.ParseUint(s[5:7], 10, 64)
	m := uint(mt)
	if err != nil {
		return Dob{}, ErrNotParseable
	}
	if m > 12 || m <= 0 {
		return Dob{}, ErrImpossible
	}
	dt, err := strconv.ParseUint(s[8:], 10, 64)
	d := uint(dt)
	// if days more than max possible day of a month
	if d > possibleMaxMonthDays[m-1] || d <= 0 {
		return Dob{}, ErrImpossible
	}
	// if feb and non leap and day==29
	if m == 2 && y%4 != 0 && d == 29 {
		return Dob{}, ErrImpossible
	}
	return Dob{y, m, d}, nil
}

// Assumes string s is validated previously,
// and does not check for errors,
// it will panic if string indexing fails
func MustParse(s string) Dob {
	y, _ := strconv.ParseUint(s[:4], 10, 64)
	m, _ := strconv.ParseUint(s[5:7], 10, 64)
	d, _ := strconv.ParseUint(s[8:], 10, 64)
	return Dob{uint(y), uint(m), uint(d)}
}

// Calculates age from dob instance
func (d Dob) Age() uint {
	now := time.Now()
	year := uint(now.Year())
	day := uint(now.Day())
	month := uint(now.Month())

	age := year - d.Year
	if month < d.Month ||
		(month == d.Month && day < d.Day) {
		age--
	}
	return age
}

// Converts the Dob instance to string using the format defined
func (d Dob) String() string {
	return strconv.FormatUint(uint64(d.Year), 10) +
		"-" +
		strconv.FormatUint(uint64(d.Month), 10) +
		"-" +
		strconv.FormatUint(uint64(d.Day), 10)
}

package dob

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"
)

type Dob struct {
	Year  uint16
	Month uint8
	Day   uint8
}

// Dob errors
var (
	ErrEmpty           = errors.New("dob: empty")
	ErrIncorrectFormat = errors.New("dob: wrong format, format must be YYYY-MM-DD")
	ErrNotParseable    = errors.New("dob: not parseable")
	ErrImpossible      = errors.New("dob: dob is impossible")
)

var possibleMaxMonthDays = [12]uint8{31, 29, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}

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
	y := uint16(yt)
	if err != nil {
		return Dob{}, ErrNotParseable
	}
	if y > uint16(time.Now().Year()) || y < 1900 { // 1900 is the lower limit, the oldest living person as of now was born on 1909, so we can safely put this.
		return Dob{}, ErrImpossible
	}
	mt, err := strconv.ParseUint(s[5:7], 10, 64)
	m := uint8(mt)
	if err != nil {
		return Dob{}, ErrNotParseable
	}
	if m > 12 || m <= 0 {
		return Dob{}, ErrImpossible
	}
	dt, err := strconv.ParseUint(s[8:], 10, 64)
	d := uint8(dt)
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

// Wraps [Parse] in a way such that if it returns a error then this panics
//
// recommended to use [AlwaysParse] instead if you are sure that string will always be valid
func MustParse(s string) Dob {
	d, err := Parse(s)
	if err != nil {
		panic(fmt.Errorf("Dob: fail to parse, %w", err))
	}
	return d
}

// Returns a Dob object from a string
//
// call this only if [Parse] has parsed this string before without errors.
//
// this is a faster version of [MustParse] and is recommended to use if you
// are sure string is valid and does not contain invalid data or format
//
// WARN:
// this function does not validate the string or check for errors
// and this can panic if string indexing fails (incorrect string format).
// if format matches and the values have invalid/impossible year/month/date values
// then it will not fail and return bad data which will inevitably panic if called [Dob.Age] later on.
// Always make sure string is validated according to [Parse]
func AlwaysParse(s string) Dob {
	y, _ := strconv.ParseUint(s[:4], 10, 16)
	m, _ := strconv.ParseUint(s[5:7], 10, 8)
	d, _ := strconv.ParseUint(s[8:], 10, 8)
	return Dob{uint16(y), uint8(m), uint8(d)}
}

// Calculates age from using current time
//
// assumes [Dob] was constructed using [Parse] or [MustParse].
// if age comes out to be greater than the maximum value of an
// uint8 type then this function panics
func (d Dob) Age() uint8 {
	now := time.Now()
	year := uint16(now.Year())
	day := uint8(now.Day())
	month := uint8(now.Month())

	age := year - d.Year
	if month < d.Month ||
		(month == d.Month && day < d.Day) {
		age--
	}
	// there are a few cases when this will get triggered
	// 1. if user constructs [Dob] by himself instead of using [Parse]
	if age > math.MaxUint8 {
		panic("age is computed to be greater than " + strconv.Itoa(math.MaxUint8) + ", which is impossible.")
	}
	return uint8(age)
}

// Converts the Dob instance to string using the format defined
func (d Dob) String() string {
	return strconv.FormatUint(uint64(d.Year), 10) +
		"-" +
		strconv.FormatUint(uint64(d.Month), 10) +
		"-" +
		strconv.FormatUint(uint64(d.Day), 10)
}

package jwt

import "time"

// taken ideas from [github.com/golang-jwt/jwt/v5]

// type to represent unix time in jwt
type NumericTime uint

// returns a new [NumericTime] from a [time.Time]
func NewNumericTime(t time.Time) NumericTime {
	return NumericTime(t.Unix())
}

// returns the [time.Time] from the unix value of [NumericTime]
func (t NumericTime) Time() time.Time {
	return time.Unix(int64(t), 0)
}

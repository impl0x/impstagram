package jwt

import "time"

// taken ideas from [github.com/golang-jwt/jwt/v5]

type NumericTime uint

func NewNumericTime(t time.Time) NumericTime {
	return NumericTime(t.Unix())
}

func (t NumericTime) Time() time.Time {
	return time.Unix(int64(t), 0)
}

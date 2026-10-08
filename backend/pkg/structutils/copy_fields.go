package structutils

import (
	"errors"
	"reflect"
)

var (
	errArgsMustBeStruct   = errors.New("both arguments must be structs")
	errArgTypesMustBeSame = errors.New("struct types must be the same")
)

// CopyFields copies fields from b and writes them into a if a
// is not already written to that field and has its zero value intact
//
// for example if two instances of the same struct type A contains a
// field called "f" of type int, and the two instances contain values
// 0 and 1 respectively, in this case considering the first instance is a
// and the second instance is b, the output will result in a's field "f"
// getting the value 1 because it was zero value at first and b had a value for
// that field specifically, now similarly here are some examples explaining it,
// where the first number is from struct a and second from struct b:
//   - 0 and 1 -> 1
//   - 1 and 2 -> 1
//   - 1 and 0 -> 1
//   - 0 and 0 -> 0
//
// errors can be [errArgsMustBeStruct] and [errArgTypesMustBeSame],
//   - a and b must be of the same type
//   - a needs to be a pointer to the struct type otherwise it is impossible to set a value
//   - b can be either a pointer or a value struct, does not matter as it is read only
func CopyFields(a, b any) error {
	valA := reflect.ValueOf(a)
	valB := reflect.ValueOf(b)

	if valA.Kind() == reflect.Pointer {
		valA = valA.Elem()
	}
	if valB.Kind() == reflect.Pointer {
		valB = valB.Elem()
	}

	if valA.Kind() != reflect.Struct || valB.Kind() != reflect.Struct {
		return errArgsMustBeStruct
	}

	if valA.Type() != valB.Type() {
		return errArgTypesMustBeSame
	}
	for i := 0; i < valA.NumField(); i++ {
		fieldA := valA.Field(i)
		fieldB := valB.Field(i)

		if fieldA.Kind() == reflect.Struct && !fieldA.CanSet() {
			if err := CopyFields(fieldA.Addr().Interface(), fieldB.Interface()); err != nil {
				return err
			}
		} else if fieldA.CanSet() {
			if fieldA.IsZero() && fieldB.CanInterface() {
				fieldA.Set(fieldB)
			}
		}
	}
	return nil
}

package env

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// the tag name to be used in struct tags
const TagName = "env"

// binds the struct with the environment variables
//
// currently only supported tag format is:
//
//	Field type `env:"FIELD_NAME,required,default=DEFAULT_NAME"`
//
// where the first name segment is required, rest two are optional.
//
// examples:
//
//	Port int `env:"PORT,required,default=8080"`
//	DBConn string `env:"DB_CONN",required`
//	AppName string `env:"APP_NAME"
//
// the format must be followed to some extent where the first element must always
// be the name of the environment variable and the following 2nd and 3rd element
// can be "required" or "default=..." interchangeably
func Parse(a any) error {
	rv := reflect.ValueOf(a)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return errors.New("env: variable passed is not of type struct, cannot bind env variables to it")
	}
	return parse(rv)
}

func parse(v reflect.Value) error {
	if v.Kind() != reflect.Struct {
		panic("kind not struct")
	}
	t := v.Type()
	for i := range t.NumField() {
		ct := t.Field(i)
		if !ct.IsExported() {
			continue
		}
		cv := v.Field(i)
		kind := ct.Type.Kind()
		if kind == reflect.Struct {
			err := parse(cv)
			if err != nil {
				return err
			}
		}
		tag := ct.Tag.Get(TagName)
		if tag == "" {
			continue
		}
		segs := strings.Split(tag, ",")
		if len(segs) > 3 {
			return errors.New("env: found more than 3 segments in the tag value, incorrect format please read documentation for the function, field: " + ct.Name)
		}
		var isRequired bool
		var defaultValue string
		for i := 1; i < len(segs); i++ {
			e := segs[i]
			if e == "required" {
				isRequired = true
			} else if e[:8] == "default=" {
				defaultValue = e[8:]
			} else {
				return errors.New("env: incorrect format, second or third element does not contain either required or default, field: " + ct.Name)
			}
		}
		envName := segs[0]
		envVal := os.Getenv(envName)

		if envVal == "" {
			if isRequired {
				return errors.New("env: required variable not found in environment, field: " + ct.Name)
			}
			envVal = defaultValue
		}

		if !cv.CanSet() {
			return errors.New("env: cannot set field value, field: " + ct.Name)
		}
		err := setValueFromString(cv, envVal)
		if err != nil {
			return fmt.Errorf("env: cannot set field value, %w, field: %s", err, ct.Name)
		}
	}
	return nil
}

// *Modified*
// Source - https://stackoverflow.com/q/39891689
// Posted by Tarion, modified by community. See post 'Timeline' for change history
// Retrieved 2026-10-06, License - CC BY-SA 3.0
//
// sets a string value to a reflect.Value appropriately
func setValueFromString(v reflect.Value, strVal string) error {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var t time.Duration
		if v.Type()==reflect.TypeOf(t){
			t,err:=time.ParseDuration(strVal)
			if err!=nil{
				return fmt.Errorf("invalid time.Duration value, %w",err)
			}
			v.SetInt(int64(t))
			return nil
		}
		val, err := strconv.ParseInt(strVal, 0, 64)
		if err != nil && strVal != "" {
			return err
		}
		if v.OverflowInt(val) {
			return errors.New("Int value too big: " + strVal)
		}
		v.SetInt(val)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val, err := strconv.ParseUint(strVal, 0, 64)
		if err != nil && strVal != "" {
			return err
		}
		if v.OverflowUint(val) {
			return errors.New("UInt value too big: " + strVal)
		}
		v.SetUint(val)
	case reflect.Float32:
		val, err := strconv.ParseFloat(strVal, 32)
		if err != nil && strVal != "" {
			return err
		}
		v.SetFloat(val)
	case reflect.Float64:
		val, err := strconv.ParseFloat(strVal, 64)
		if err != nil && strVal != "" {
			return err
		}
		v.SetFloat(val)
	case reflect.String:
		v.SetString(strVal)
	case reflect.Bool:
		val, err := strconv.ParseBool(strVal)
		if err != nil && strVal != "" {
			return err
		}
		v.SetBool(val)
	default:
		return errors.New("Unsupported kind: " + v.Kind().String())
	}
	return nil
}

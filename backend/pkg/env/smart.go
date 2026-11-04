package env

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
)

// this is a experimental method of env parsing where structs
// without explicit struct tags are parsed by particular naming
// rules to env variables, as elaborated below.
//
// for example previously and the way [env.go] does it, the default
// way to parse environment variables into structs was as follows.
// 	type Config struct {
// 		DatabaseURL string `env:"DATABASE_URL"`
// 		ServerPort  uint   `env:"SERVER_PORT"`
// 	}
// we needed to put struct tags on each field with the environment variable name,
// which was causing a problem in my certain use case where I did not want
// service layer importing global config nor did I want it to know about environment
// variable names, and this problem was not fixable unless I wanted to map 2 different
// structs with the same data. So I thought of a new way to parse and use env tags.
//
// Instead of explicitly setting struct tags, we can implicitly infer the variable name,
// because by convention environment variables are written in SCREAMING_SNAKE_CASE, while
// golang variables are written in CamelCase, so it is possible to for example set a rule
// where env vars can be read from a struct containing no struct tags or explicit naming
// by using these conventions, where we can infer if the field name is "DatabaseURL" then
// the environment variable name must be "DATABASE_URL".
//
// Although this is extremely specific for my use case and not at flexible, because what if
// the environment variable is named "DB_URL"? Not every time is the convention followed,
// but its actually better to enforce the convention than to write mapping functions or allow
// leaking of concerns.
//
// Another concern is with required and default options, for my use case every field is required
// and default is declared as another struct, and the pattern is used where only populated fields
// from the env vars are filled into the default and then that modified struct is passed onto service.

// TODO: add comments to everything

type smartParseSettings struct {
	globalPrefix string
	localPrefix  string
}

type SmartParseOption func(*smartParseSettings)

func WithGlobalPrefix(prefix string) SmartParseOption {
	return func(s *smartParseSettings) {
		s.globalPrefix = prefix
	}
}
func withLocalPrefix(prefix string) SmartParseOption {
	return func(s *smartParseSettings) {
		s.localPrefix = prefix
	}
}

func SmartParse(a any, opts ...SmartParseOption) error {
	var st smartParseSettings
	for _, o := range opts {
		o(&st)
	}
	rv := reflect.ValueOf(a)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return errors.New("env: variable passed is not of type struct, cannot bind env variables to it")
	}
	return smartParse(rv, st)
}

func smartParse(v reflect.Value, st smartParseSettings, opts ...SmartParseOption) error {
	if v.Kind() != reflect.Struct {
		panic("kind not struct")
	}
	for _, o := range opts {
		o(&st)
	}
	t := v.Type()
	for i := range t.NumField() {
		ct := t.Field(i)
		if !ct.IsExported() {
			continue
		}
		cv := v.Field(i)
		kind := ct.Type.Kind()
		switch kind {
		case reflect.Struct:
			err := smartParse(cv, st, withLocalPrefix(screamingSnakeCasify(ct.Name)))
			if err != nil {
				return err
			}
			continue
		case reflect.Pointer:
			v = v.Elem()
			if v.Kind() != reflect.Struct {
				return errors.New("env: Pointer to non struct object, nil setting is unsupported as of now")
			}
			continue
		}
		var envName string
		if st.globalPrefix != "" {
			envName += st.globalPrefix + "_"
		}
		if st.localPrefix != "" {
			envName += st.localPrefix + "_"
		}
		envName += screamingSnakeCasify(ct.Name)
		envVal, ok := os.LookupEnv(envName)
		if !ok {
			continue
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

func screamingSnakeCasify(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 16) // buffer of 16
	for i := 0; i < len(s); i++ {
		if i != 0 && isUpper(s[i]) {
			b.WriteByte('_')
		}
		b.WriteByte(toUpper(s[i]))
	}
	return b.String()
}

func toUpper(b byte) byte {
	if isLower(b) {
		return b - 32
	}
	return b
}

func isLower(b byte) bool {
	return b >= 97 && b <= 122
}

func isUpper(b byte) bool {
	return b >= 65 && b <= 90
}

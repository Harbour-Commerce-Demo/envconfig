// Package envconfig fills structs from environment variables.
package envconfig

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Load fills the exported fields of the struct v points to. A field reads the
// variable named by its `env` tag, or its upper-cased name, joined to prefix
// with "_". A `default` tag is used when the variable is unset. Nested structs
// and struct pointers extend the prefix with their own name.
func Load(prefix string, v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.Elem().Kind() != reflect.Struct {
		return errors.New("envconfig: Load needs a pointer to a struct")
	}
	return load(prefix, rv.Elem())
}

func load(prefix string, rv reflect.Value) error {
	rt := rv.Type()
	for i := range rt.NumField() {
		f := rt.Field(i)
		if !f.IsExported() {
			continue
		}
		name := f.Tag.Get("env")
		if name == "" {
			name = strings.ToUpper(f.Name)
		}
		if prefix != "" {
			name = prefix + "_" + name
		}

		fv := rv.Field(i)
		if f.Type.Kind() == reflect.Struct && f.Type != durationType {
			if err := load(name, fv); err != nil {
				return err
			}
			continue
		}
		if f.Type.Kind() == reflect.Pointer && f.Type.Elem().Kind() == reflect.Struct {
			if fv.IsNil() {
				fv.Set(reflect.New(f.Type.Elem()))
			}
			if err := load(name, fv.Elem()); err != nil {
				return err
			}
			continue
		}

		raw, ok := os.LookupEnv(name)
		if !ok {
			raw, ok = f.Tag.Lookup("default")
		}
		if !ok {
			continue
		}
		if err := set(fv, raw); err != nil {
			return fmt.Errorf("envconfig: %s: %w", name, err)
		}
	}
	return nil
}

var durationType = reflect.TypeFor[time.Duration]()

func set(fv reflect.Value, raw string) error {
	if fv.Type() == durationType {
		d, err := time.ParseDuration(raw)
		if err != nil {
			return err
		}
		fv.SetInt(int64(d))
		return nil
	}

	switch fv.Kind() {
	case reflect.String:
		fv.SetString(raw)
	case reflect.Bool:
		b, err := strconv.ParseBool(raw)
		if err != nil {
			return err
		}
		fv.SetBool(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(raw, 10, fv.Type().Bits())
		if err != nil {
			return err
		}
		fv.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(raw, 10, fv.Type().Bits())
		if err != nil {
			return err
		}
		fv.SetUint(n)
	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(raw, fv.Type().Bits())
		if err != nil {
			return err
		}
		fv.SetFloat(n)
	case reflect.Slice:
		if fv.Type().Elem().Kind() != reflect.String {
			return fmt.Errorf("unsupported type %s", fv.Type())
		}
		var parts []string
		for p := range strings.SplitSeq(raw, ",") {
			if p = strings.TrimSpace(p); p != "" {
				parts = append(parts, p)
			}
		}
		fv.Set(reflect.ValueOf(parts))
	default:
		return fmt.Errorf("unsupported type %s", fv.Type())
	}
	return nil
}

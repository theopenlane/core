package envparse

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/stoewer/go-strcase"
)

// ErrInvalidSpecification indicates that a specification is of the wrong type.
var ErrInvalidSpecification = errors.New("specification must be a struct pointer")

type Config struct {
	// FieldTagName is the name of the struct tag to use for the field name
	FieldTagName string
	// Skipper is the value of the tag to skip parsing of the field
	Skipper string
}

// varInfo maintains information about the configuration variable
type varInfo struct {
	FieldName string
	FullPath  string
	Key       string
	Type      reflect.Type
	Tags      reflect.StructTag
}

// GatherEnvInfo gathers information about the specified struct, including defaults and environment variable names.
func (c Config) GatherEnvInfo(prefix string, spec interface{}) ([]varInfo, error) {
	s := reflect.ValueOf(spec)

	// Ensure the specification is a pointer to a struct
	if s.Kind() != reflect.Ptr {
		return nil, ErrInvalidSpecification
	}

	s = s.Elem()
	if s.Kind() != reflect.Struct {
		return nil, ErrInvalidSpecification
	}

	typeOfSpec := s.Type()

	// Create a slice to hold the information about the configuration variables
	var infos []varInfo

	// Iterate over the struct fields
	for i := range s.NumField() {
		f := s.Field(i)
		ftype := typeOfSpec.Field(i)

		if !f.CanSet() {
			continue
		}

		for f.Kind() == reflect.Ptr {
			if f.IsNil() {
				if f.Type().Elem().Kind() != reflect.Struct {
					// nil pointer to a non-struct: leave it alone
					break
				}

				// nil pointer to struct: create a zero instance
				f.Set(reflect.New(f.Type().Elem()))
			}

			f = f.Elem()
		}

		// Capture information about the config variable
		fieldName := c.getFieldName(ftype)
		if fieldName == c.Skipper {
			continue
		}

		info := varInfo{
			FieldName: fieldName,
			FullPath:  ftype.Name,
			Type:      ftype.Type,
			Tags:      ftype.Tag,
		}

		// Default to the field name as the env var name (will be upcased)
		info.Key = info.FieldName

		if prefix != "" {
			info.Key = fmt.Sprintf("%s_%s", prefix, info.Key)
			info.FullPath = fmt.Sprintf("%s.%s", strcase.LowerCamelCase(strings.Replace(prefix, "_", ".", -1)), info.FieldName) // nolint: gocritic
		}

		info.Key = strings.ToUpper(info.Key)
		infos = append(infos, info)

		if f.Kind() == reflect.Struct {
			innerPrefix := prefix

			if !ftype.Anonymous {
				innerPrefix = prefix + "_" + info.Tags.Get("json")
			}

			embeddedPtr := f.Addr().Interface()

			// Recursively gather information about the embedded struct
			embeddedInfos, err := c.GatherEnvInfo(innerPrefix, embeddedPtr)
			if err != nil {
				return nil, err
			}

			infos = append(infos[:len(infos)-1], embeddedInfos...)

			continue
		}
	}

	return infos, nil
}

func (c Config) getFieldName(ftype reflect.StructField) string {
	if ftype.Tag.Get(c.FieldTagName) != "" {
		return ftype.Tag.Get(c.FieldTagName)
	}

	// default to skip if the koanf tag is not present
	return c.Skipper
}

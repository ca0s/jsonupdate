package jsonupdate

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/ca0s/jsonfilter"
	"golang.org/x/exp/slices"
)

type JsonUpdatable interface{}

type Update struct {
	Field  string      `json:"field"`
	Action string      `json:"action"`
	Value  interface{} `json:"value"`
}

var (
	ErrPathNotFound  = fmt.Errorf("jsonupdate: path not found")
	ErrInvalidPath   = fmt.Errorf("jsonupdate: part of the update path is not compatible")
	ErrInvalidType   = fmt.Errorf("jsonupdate: destination field is a non compatible type")
	ErrUnknownAction = fmt.Errorf("jsonupdate: unknown update action")
	ErrInvalidUpdate = fmt.Errorf("jsonupdate: invalid update config")
)

func (u *Update) Validate() error {
	u.Field = strings.TrimSpace(u.Field)
	if u.Field == "" {
		return fmt.Errorf("%w: field cannot be empty", ErrInvalidUpdate)
	}

	if !slices.Contains([]string{"set", "delete"}, u.Action) {
		return fmt.Errorf("%w: invalid action", ErrInvalidUpdate)
	}

	return nil
}

func (u *Update) Apply(item JsonUpdatable) error {
	path := u.Field

	switch u.Action {
	case "set":
		return SetField(item, path, u.Value)
	case "delete":
		return SetField(item, path, nil)
	}

	return ErrUnknownAction
}

var ErrItemNotPointer = fmt.Errorf("item is not a pointer")
var ErrItemFieldCannotBeSet = fmt.Errorf("item field cannot be set")
var ErrIncompatibleTypes = fmt.Errorf("incompatible fields")

func SetField(item interface{}, name string, value interface{}) error {
	if strings.Contains(name, ".") {
		p := strings.SplitN(name, ".", 2)

		first, err := jsonfilter.GetField(item, p[0])
		if err != nil {
			return err
		}

		return SetField(first, strings.Join(p[1:], "."), value)
	}

	t := reflect.TypeOf(item)
	if t.Kind() != reflect.Pointer && t.Kind() != reflect.Interface {
		return ErrItemNotPointer
	}

	te := t.Elem()

	v := reflect.ValueOf(item)
	if v.Kind() != reflect.Pointer && v.Kind() != reflect.Interface {
		return ErrItemNotPointer
	}

	ve := v.Elem()

	nvv := reflect.ValueOf(value)
	nvt := reflect.TypeOf(value)

	for i := 0; i < te.NumField(); i++ {
		for _, tagName := range jsonfilter.TagNames {
			if tag, ok := te.Field(i).Tag.Lookup(tagName); ok {
				if tag == name {
					fieldVal := ve.Field(i)

					if !fieldVal.CanAddr() {
						return ErrItemNotPointer
					}

					if !fieldVal.CanSet() {
						return ErrItemFieldCannotBeSet
					}

					if !nvt.AssignableTo(fieldVal.Type()) {
						if nvt.ConvertibleTo(fieldVal.Type()) {
							nvv = nvv.Convert(fieldVal.Type())
						} else {
							return ErrIncompatibleTypes
						}
					}

					fieldVal.Set(nvv)
					return nil
				}
			}
		}
	}

	return jsonfilter.ErrFieldNotPresent
}

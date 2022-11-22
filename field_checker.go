package field_checker

import (
	"fmt"
	"reflect"
)

func isEmpty(v reflect.Value) bool {
	kind := v.Kind()
	if kind == reflect.Ptr {
		if v.IsNil() {
			return true
		}
	}
	if v.IsZero() {
		return true
	}

	switch kind {
	case reflect.Int:
		if v == reflect.ValueOf(int(0)) {
			return true
		}
	case reflect.Int8:
		if v == reflect.ValueOf(int8(0)) {
			return true
		}
	case reflect.Int16:
		if v == reflect.ValueOf(int16(0)) {
			return true
		}
	case reflect.Int32:
		if v == reflect.ValueOf(int32(0)) {
			return true
		}
	case reflect.Int64:
		if v == reflect.ValueOf(int64(0)) {
			return true
		}
	case reflect.Float32:
		if v == reflect.ValueOf(float32(0)) {
			return true
		}
	case reflect.Float64:
		if v == reflect.ValueOf(float64(0)) {
			return true
		}
	}
	return false
}

func isIgnoreTypes(_type interface{}, ignoreTypes []interface{}) bool {
	for i := range ignoreTypes {
		if reflect.TypeOf(_type) == reflect.TypeOf(ignoreTypes[i]) {
			return true
		}
	}
	return false
}

func checkHorizontal(src interface{}, ignoreTypes []interface{}) ([]reflect.Value, error) {
	var verticalFields []reflect.Value

	var s reflect.Value
	if reflect.TypeOf(src) != reflect.TypeOf(reflect.Value{}) {
		s = reflect.ValueOf(src)
	} else {
		s = src.(reflect.Value)
	}

	if isEmpty(s) && !isIgnoreTypes(s.Type(), ignoreTypes) {
		return nil, fmt.Errorf("name: %s is empty\n", reflect.TypeOf(s).Name())
	}

	if s.Kind() == reflect.Ptr {
		s = reflect.Indirect(s)
	}

	if s.Type().Kind() == reflect.Slice {
		if s.Len() == 0 {
			return nil, fmt.Errorf("name: %s is empty, type = %s\n", s.Type().Name(), s.Type().String())
		}
		for i := 0; i < s.Len(); i++ {
			if s.Index(i).Type().Kind() == reflect.Struct || s.Index(i).Type().Kind() == reflect.Interface || s.Index(i).Type().Kind() == reflect.Ptr {
				verticalFields = append(verticalFields, s.Index(i))
			}
		}
	}

	if s.Type().Kind() == reflect.Map {
		if s.Len() == 0 {
			return nil, fmt.Errorf("name: %s is empty, type = %s\n", s.Type().Name(), s.Type().String())
		}

		keys := s.MapKeys()
		for i := range keys {
			if s.MapIndex(keys[i]).Type().Kind() == reflect.Struct || s.MapIndex(keys[i]).Type().Kind() == reflect.Interface || s.Index(i).Type().Kind() == reflect.Ptr {
				verticalFields = append(verticalFields, s.MapIndex(keys[i]))
			}
		}
	}

	var exportFields []reflect.StructField
	if s.Type().Kind() == reflect.Struct {
		exportFields = reflect.VisibleFields(s.Type())
	}

	for i := 0; i < len(exportFields); i++ {
		if !exportFields[i].IsExported() {
			continue
		}
		f := s.FieldByName(exportFields[i].Name)

		if isEmpty(s) && !isIgnoreTypes(s.Type(), ignoreTypes) {
			return nil, fmt.Errorf("parent := %s name: %s is empty,  type = %s\n", s.Type().Name(), s.Type().Field(i).Name, s.Type().Field(i).Type.String())
		}

		if f.Kind() == reflect.Struct || f.Kind() == reflect.Ptr {
			verticalFields = append(verticalFields, f)
		}

		if f.Kind() == reflect.Slice {
			if f.Len() == 0 {
				return nil, fmt.Errorf("parent := %s name: %s is empty, type = %s \n ", s.Type().Name(), s.Type().Field(i).Name, s.Type().Field(i).Type.String())
			}
			for i := 0; i < f.Len(); i++ {
				if f.Index(i).Type().Kind() == reflect.Struct || f.Index(i).Type().Kind() == reflect.Interface || f.Index(i).Type().Kind() == reflect.Ptr {
					verticalFields = append(verticalFields, f.Index(i))
				}
			}
		}

		if f.Kind() == reflect.Map {
			if f.Len() == 0 {
				return nil, fmt.Errorf("parent := %s  name: %s is empty, type = %s\n", s.Type().Name(), s.Type().Field(i).Name, s.Type().Field(i).Type.String())
			}
			keys := f.MapKeys()
			for i := 0; i < f.Len(); i++ {
				if f.MapIndex(keys[i]).Type().Kind() == reflect.Struct || f.MapIndex(keys[i]).Type().Kind() == reflect.Interface || f.MapIndex(keys[i]).Type().Kind() == reflect.Ptr || f.MapIndex(keys[i]).Type().Kind() == reflect.Slice {
					verticalFields = append(verticalFields, f.MapIndex(keys[i]))
				}
			}
		}
	}
	return verticalFields, nil
}

func CheckStruct(src interface{}, ignoreTypes []interface{}) error {
	shouldCheckers, err := checkHorizontal(src, ignoreTypes)
	if err != nil {
		return err
	}
	var uncheckFields []reflect.Value
	for len(shouldCheckers) != 0 {
		tmp, err := checkHorizontal(shouldCheckers[0], ignoreTypes)
		if err != nil {
			return err
		}
		uncheckFields = append(uncheckFields, tmp...)
		if len(shouldCheckers) > 1 {
			shouldCheckers = shouldCheckers[1:]
		} else {
			shouldCheckers = uncheckFields
			uncheckFields = make([]reflect.Value, 0)
		}
	}
	return nil
}

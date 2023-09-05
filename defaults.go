package piranhas

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	errSyntax = errors.New("invalid syntax")
)

// SetDefaults sets default values for fields in a struct, elements in a slice, or values in a map, based on the type of the provided pointer
func SetDefaults(ptr interface{}) (err error) {
	// obtain the reflect.Value of the provided pointer
	v := reflect.ValueOf(ptr)
	// check if the provided value is a pointer
	if v.Kind() != reflect.Ptr {
		return
	}

	// determine the type of the object
	objType := v.Type()
	for objType.Kind() == reflect.Ptr {
		objType = objType.Elem()
	}

	// obtain the kind of the value that the pointer points to
	switch objType.Kind() {
	case reflect.Struct:
		// set defaults for struct fields
		err = setDefaultsStruct(ptr)
	case reflect.Slice, reflect.Array:
		// set defaults for slice and array elements
		err = setDefaultsSlice(ptr)
	case reflect.Map:
		// set defaults for map values
		err = setDefaultsMap(ptr)
	}

	return err
}

// setDefaultsStruct sets default values for elements in a struct
func setDefaultsStruct(ptr interface{}) (err error) {
	// read all pointers away
	objValue := reflect.ValueOf(ptr)
	for {
		if objValue.Kind() == reflect.Ptr {
			if objValue.IsNil() {
				return nil
			}
			objValue = objValue.Elem()
		} else {
			// break the loop if objValue is not a pointer
			break
		}
	}
	objType := objValue.Type()

	// check if the value the pointer points to is a struct
	if objType.Kind() != reflect.Struct {
		return nil
	}

	// iterate over all fields of the struct
	for i := 0; i < objType.NumField(); i++ {
		// Get field and its value
		field := objType.Field(i)
		fieldValue := objValue.Field(i)
		defaultTag := field.Tag.Get("default")
		layoutTag := field.Tag.Get("layout")

		// determine the type of the field element
		fieldValueType := fieldValue.Type()
		for fieldValueType.Kind() == reflect.Ptr {
			fieldValueType = fieldValueType.Elem()
		}

		// set or call recursively based on field type
		switch fieldValueType.Kind() {
		case reflect.Invalid:
			// do nothing for invalid type
		case reflect.Struct:
			if fieldValue.Type().String() == "time.Time" && defaultTag != "" {
				defaultValue, err := parseDefaultValue(defaultTag, layoutTag, fieldValueType)
				if err != nil {
					return fmt.Errorf("failed to parse default tag for field %s: %s", field.Name, err)
				}

				// overwrite the value with the default value
				setUnexportedField(fieldValue, defaultValue)
			} else {
				err = setDefaultsStruct(getPtrInterface(fieldValue))
			}

		case reflect.Slice, reflect.Array:
			err = setDefaultsSlice(getPtrInterface(fieldValue))
		case reflect.Map:
			err = setDefaultsMap(getPtrInterface(fieldValue))
		default:
			// handle scalar data types
			if defaultTag != "" {
				defaultValue, err := parseDefaultValue(defaultTag, layoutTag, fieldValueType)
				if err != nil {
					return fmt.Errorf("failed to parse default tag for field %s: %s", field.Name, err)
				}

				// overwrite the value with the default value
				setUnexportedField(fieldValue, defaultValue)
			}
		}
	}

	return
}

// setDefaultsSlice sets default values for elements in a slice or array
func setDefaultsSlice(ptr interface{}) (err error) {
	// read all pointers away
	objValue := reflect.ValueOf(ptr)
	for {
		if objValue.Kind() == reflect.Ptr {
			if objValue.IsNil() {
				return nil
			}
			objValue = objValue.Elem()
		} else {
			// break the loop if objValue is not a pointer
			break
		}
	}
	objType := objValue.Type()

	// check if the value the pointer points to is a slice or array
	if objType.Kind() != reflect.Slice && objType.Kind() != reflect.Array {
		return nil
	}

	// determine the slice element type
	sliceType := objValue.Type().Elem()
	if sliceType.Kind() == reflect.Ptr {
		sliceType = sliceType.Elem()
	}

	// iterate through each element in the slice
	for i := 0; i < objValue.Len(); i++ {
		elemValue := objValue.Index(i)

		// determine the type of the slice or array element
		elemValueType := elemValue.Type()
		for elemValueType.Kind() == reflect.Ptr {
			elemValueType = elemValueType.Elem()
		}

		// process different kinds of elements
		switch elemValueType.Kind() {
		case reflect.Struct:
			// recursively set defaults for struct elements
			err = setDefaultsStruct(getPtrInterface(elemValue))
		case reflect.Slice, reflect.Array:
			// recursively set defaults for slice or array elements
			err = setDefaultsSlice(getPtrInterface(elemValue))
		case reflect.Map:
			// recursively set defaults for map elements
			err = setDefaultsMap(getPtrInterface(elemValue))
		}

		// if an error occurs during setting defaults, return the error
		if err != nil {
			return err
		}
	}

	return
}

// setDefaultsMap sets default values for elements in a map
func setDefaultsMap(ptr interface{}) (err error) {
	// read all pointers away
	objValue := reflect.ValueOf(ptr)
	for {
		if objValue.Kind() == reflect.Ptr {
			if objValue.IsNil() {
				return nil
			}
			objValue = objValue.Elem()
		} else {
			// break the loop if objValue is not a pointer
			break
		}
	}
	objType := objValue.Type()

	// check if the value the pointer points to is a map
	if objType.Kind() != reflect.Map {
		return nil
	}

	// iterate through keys of the map
	for _, key := range objValue.MapKeys() {
		elemValue := objValue.MapIndex(key)
		elemPtr := reflect.New(elemValue.Type()).Elem()
		elemPtr.Set(elemValue)

		// determine the type of the map element
		elemValueType := elemValue.Type()
		for elemValueType.Kind() == reflect.Ptr {
			elemValueType = elemValueType.Elem()
		}

		// process different kinds of map elements
		switch elemValueType.Kind() {
		case reflect.Struct:
			// recursively set defaults for struct elements
			err = setDefaultsStruct(getPtrInterface(elemPtr))
		case reflect.Slice, reflect.Array:
			// recursively set defaults for slice or array elements
			err = setDefaultsSlice(getPtrInterface(elemPtr))
		case reflect.Map:
			// recursively set defaults for map elements
			err = setDefaultsMap(getPtrInterface(elemPtr))
		}

		// if an error occurs during setting defaults, return the error
		if err != nil {
			return err
		}

		// store the updated elements because elemValue is not storable
		objValue.SetMapIndex(key, elemPtr)
	}

	return nil
}

// parseDefaultValue parses the default tag and converts it to a value for scalar data types
func parseDefaultValue(defaultTag string, layoutTag string, fieldType reflect.Type) (reflect.Value, error) {
	kind := fieldType.Kind()

	// if the field type is a pointer, process the pointed-to type recursively
	if kind == reflect.Ptr {
		elemType := fieldType.Elem()
		defaultValue, err := parseDefaultValue(defaultTag, layoutTag, elemType)
		if err != nil {
			return reflect.Value{}, err
		}
		ptrValue := reflect.New(elemType)
		ptrValue.Elem().Set(defaultValue)
		return ptrValue, nil
	}

	switch kind {
	case reflect.String:
		// for string fields, return a reflect.Value with the defaultTag value
		return reflect.ValueOf(defaultTag), nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if fieldType.String() == "time.Duration" {
			dur, err := time.ParseDuration(defaultTag)
			if err != nil {
				return reflect.Value{}, errSyntax
			}
			return reflect.ValueOf(dur), nil
		}

		// for integer fields, parse the defaultTag as an integer and convert it to the field type
		defaultValue, err := strconv.ParseInt(defaultTag, 10, fieldType.Bits())
		if err != nil {
			return reflect.Value{}, errSyntax
		}
		return reflect.ValueOf(defaultValue).Convert(fieldType), nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		// for unsigned integer fields, parse the defaultTag as an unsigned integer and convert it to the field type
		defaultValue, err := strconv.ParseUint(defaultTag, 10, fieldType.Bits())
		if err != nil {
			return reflect.Value{}, errSyntax
		}
		return reflect.ValueOf(defaultValue).Convert(fieldType), nil

	case reflect.Float32, reflect.Float64:
		// for floating-point fields, parse the defaultTag as a float and convert it to the field type
		defaultValue, err := strconv.ParseFloat(defaultTag, fieldType.Bits())
		if err != nil {
			return reflect.Value{}, errSyntax
		}
		return reflect.ValueOf(defaultValue).Convert(fieldType), nil

	case reflect.Complex64:
		// for complex54 fields, parse the defaultTag as a complex and convert it to the field type
		defaultValue, err := parseComplex(defaultTag)
		if err != nil {
			return reflect.Value{}, errSyntax
		}
		return reflect.ValueOf(complex64(defaultValue)).Convert(fieldType), nil

	case reflect.Complex128:
		// for complex128 fields, parse the defaultTag as a complex and convert it to the field type
		defaultValue, err := parseComplex(defaultTag)
		if err != nil {
			return reflect.Value{}, errSyntax
		}
		return reflect.ValueOf(defaultValue).Convert(fieldType), nil

	case reflect.Bool:
		// for boolean fields, parse the defaultTag as a boolean value
		defaultValue, err := strconv.ParseBool(defaultTag)
		if err != nil {
			return reflect.Value{}, errSyntax
		}
		return reflect.ValueOf(defaultValue), nil

	case reflect.Uintptr:
		defaultValue, err := strconv.ParseUint(defaultTag, 0, strconv.IntSize)
		if err != nil {
			return reflect.Value{}, errSyntax
		}
		return reflect.ValueOf(defaultValue), nil

	case reflect.Struct:
		if fieldType.String() == "time.Time" {
			if strings.ToLower(defaultTag) == "now" {
				return reflect.ValueOf(time.Now()), nil
			}

			switch strings.ToLower(layoutTag) {
			case "layout":
				layoutTag = time.Layout
			case "ansic":
				layoutTag = time.ANSIC
			case "unixdate":
				layoutTag = time.UnixDate
			case "rubydate":
				layoutTag = time.RubyDate
			case "rfc822":
				layoutTag = time.RFC822
			case "rfc850":
				layoutTag = time.RFC850
			case "rfc1123":
				layoutTag = time.RFC1123
			case "RFC1123Z":
				layoutTag = time.RFC1123Z
			case "rfc3339", "":
				layoutTag = time.RFC3339
			case "rfc3339Nano":
				layoutTag = time.RFC3339Nano
			case "kitchen":
				layoutTag = time.Kitchen
			case "stamp":
				layoutTag = time.Stamp
			case "stampmilli":
				layoutTag = time.StampMilli
			case "stampmicro":
				layoutTag = time.StampMicro
			case "stampnano":
				layoutTag = time.StampNano
			case "datetime":
				layoutTag = time.DateTime
			case "dateonly":
				layoutTag = time.DateOnly
			case "timeonly":
				layoutTag = time.TimeOnly
			}

			t, err := time.Parse(layoutTag, defaultTag)
			if err != nil {
				return reflect.Value{}, errSyntax
			}
			return reflect.ValueOf(t), nil
		}
		return reflect.Value{}, fmt.Errorf("unsupported field type: %s", fieldType.Kind())

	default:
		// for unsupported field types, return an error
		return reflect.Value{}, fmt.Errorf("unsupported field type: %s", fieldType.Kind())
	}
}

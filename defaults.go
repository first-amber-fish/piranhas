package piranhas

import (
	"fmt"
	"reflect"
	"strconv"
	"unsafe"
)

// SetDefaults sets default values for fields in a struct, elements in a slice, or values in a map, based on the type of the provided pointer
func SetDefaults(ptr interface{}) (err error) {
	// obtain the reflect.Value of the provided pointer
	v := reflect.ValueOf(ptr)
	// check if the provided value is a pointer
	if v.Kind() != reflect.Ptr {
		return
	}

	// obtain the kind of the value that the pointer points to
	switch v.Elem().Kind() {
	case reflect.Struct:
		// set defaults for struct fields
		err = setDefaultsStruct(ptr)
	case reflect.Slice:
		// set defaults for slice elements
		err = setDefaultsSlice(ptr)
	case reflect.Map:
		// set defaults for map values
		err = setDefaultsMap(ptr)
	}

	return err
}

// setDefaultsSlice sets default values for elements in a struct
func setDefaultsStruct(ptr interface{}) (err error) {
	// check if the passed value is a pointer
	if reflect.TypeOf(ptr).Kind() != reflect.Ptr {
		return nil
	}

	// check if the pointer is nil
	if reflect.ValueOf(ptr).IsNil() {
		return nil
	}

	// check if the value the pointer points to is a struct
	if reflect.TypeOf(ptr).Elem().Kind() != reflect.Struct {
		return nil
	}

	// obtain the value and type of the pointer's target
	objValue := reflect.ValueOf(ptr).Elem()
	objType := objValue.Type()

	// iterate over all fields of the struct
	for i := 0; i < objType.NumField(); i++ {
		// Get field and its value
		field := objType.Field(i)
		fieldValue := objValue.Field(i)
		fieldType := fieldValue.Type()
		defaultTag := field.Tag.Get("default")

		// determine the kind of field regardless of whether it is a pointer or not
		kind := fieldType.Kind()
		if kind == reflect.Ptr {
			kind = fieldType.Elem().Kind()
		}

		// set or call recursively based on field type
		switch kind {
		case reflect.Invalid:
			// do nothing for invalid type
		case reflect.Struct:
			err = setDefaultsStruct(getPtrInterface(fieldValue))
		case reflect.Slice:
			err = setDefaultsSlice(getPtrInterface(fieldValue))
		case reflect.Map:
			err = setDefaultsMap(getPtrInterface(fieldValue))
		default:
			// handle scalar data types
			if defaultTag != "" {
				defaultValue, err := parseDefaultValue(defaultTag, fieldType)
				if err != nil {
					return fmt.Errorf("failed to parse default tag for field %s: %s", field.Name, err)
				}

				// overwrite the value with the default value
				set(fieldValue, defaultValue)
			}
		}
	}

	return
}

// setDefaultsSlice sets default values for elements in a slice
func setDefaultsSlice(ptr interface{}) (err error) {
	// check if the passed value is a pointer
	if reflect.TypeOf(ptr).Kind() != reflect.Ptr {
		return nil
	}

	// check if the pointer is nil
	if reflect.ValueOf(ptr).IsNil() {
		return nil
	}

	// check if the value the pointer points to is a slice
	if reflect.TypeOf(ptr).Elem().Kind() != reflect.Slice {
		return nil
	}

	// obtain the reflect.Value of the pointer's target
	v := reflect.ValueOf(ptr).Elem()

	// determine the slice element type
	sliceType := v.Type().Elem()
	if sliceType.Kind() == reflect.Ptr {
		sliceType = sliceType.Elem()
	}

	// iterate through each element in the slice
	for i := 0; i < v.Len(); i++ {
		elemValue := v.Index(i)

		// if the element is a pointer, dereference it
		if elemValue.Kind() == reflect.Ptr {
			elemValue = elemValue.Elem()
		}

		// process different kinds of elements
		switch elemValue.Kind() {
		case reflect.Struct:
			// recursively set defaults for struct elements
			err = setDefaultsStruct(getPtrInterface(elemValue))
		case reflect.Slice:
			// recursively set defaults for slice elements
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
	// check if the passed value is a pointer
	if reflect.TypeOf(ptr).Kind() != reflect.Ptr {
		return nil
	}

	// check if the pointer is nil
	if reflect.ValueOf(ptr).IsNil() {
		return nil
	}

	// check if the value the pointer points to is a map
	fmt.Println(reflect.TypeOf(ptr).Elem().Kind())

	if reflect.TypeOf(ptr).Elem().Kind() != reflect.Map {
		return nil
	}

	// Obtain the reflect.Value of the pointer's target
	v := reflect.ValueOf(ptr).Elem()

	// determine the map element type
	mapType := v.Type().Elem()
	if mapType.Kind() == reflect.Ptr {
		mapType = mapType.Elem()
	}

	// iterate through keys of the map
	for _, key := range v.MapKeys() {
		elemValue := v.MapIndex(key)
		elemPtr := reflect.New(elemValue.Type()).Elem()
		elemPtr.Set(elemValue)

		// determine the type of the map element
		elemValueType := elemValue.Type()
		if elemValueType.Kind() == reflect.Ptr {
			elemValueType = elemValueType.Elem()
		}

		// process different kinds of map elements
		switch elemValueType.Kind() {
		case reflect.Struct:
			// recursively set defaults for struct elements
			err = setDefaultsStruct(getPtrInterface(elemPtr))
		case reflect.Slice:
			// recursively set defaults for slice elements
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
		v.SetMapIndex(key, elemPtr)
	}

	return nil
}

// parseDefaultValue parses the default tag and converts it to a value for scalar data types
func parseDefaultValue(defaultTag string, fieldType reflect.Type) (reflect.Value, error) {
	kind := fieldType.Kind()

	// if the field type is a pointer, process the pointed-to type recursively
	if kind == reflect.Ptr {
		elemType := fieldType.Elem()
		defaultValue, err := parseDefaultValue(defaultTag, elemType)
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
		// for integer fields, parse the defaultTag as an integer and convert it to the field type
		defaultValue, err := strconv.ParseInt(defaultTag, 10, fieldType.Bits())
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(defaultValue).Convert(fieldType), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		// for unsigned integer fields, parse the defaultTag as an unsigned integer and convert it to the field type
		defaultValue, err := strconv.ParseUint(defaultTag, 10, fieldType.Bits())
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(defaultValue).Convert(fieldType), nil
	case reflect.Float32, reflect.Float64:
		// for floating-point fields, parse the defaultTag as a float and convert it to the field type
		defaultValue, err := strconv.ParseFloat(defaultTag, fieldType.Bits())
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(defaultValue).Convert(fieldType), nil
	case reflect.Bool:
		// for boolean fields, parse the defaultTag as a boolean value
		defaultValue, err := strconv.ParseBool(defaultTag)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(defaultValue), nil
	case reflect.Uintptr:
		defaultValue, err := strconv.ParseUint(defaultTag, 0, strconv.IntSize)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(defaultValue), nil
	default:
		// for unsupported field types, return an error
		return reflect.Value{}, fmt.Errorf("unsupported field type: %s", fieldType.Kind())
	}
}

// set updates the value of a given field with the provided value.
// It constructs a new reflect.Value of the same type as the field at the memory address of the field,
// then sets the new value to the provided value.
func set(field reflect.Value, value reflect.Value) {
	// Create a new reflect.Value at the memory address of the field
	// with the same type as the field, then set its value to the provided value.
	reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Set(value)
}

// getPtrInterface converts a reflect.Value to an interface value. If the field is a pointer,
// it dereferences the pointer and creates a new interface value at the same address
func getPtrInterface(field reflect.Value) interface{} {
	// If the field is a pointer, dereference it
	if field.Kind() == reflect.Ptr {
		field = field.Elem()
	}

	// Obtain the pointer address and create a new reflect.Value at that address
	ptr := reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Pointer()

	// Create a new interface value using the type and pointer obtained above
	return reflect.NewAt(field.Type(), unsafe.Pointer(ptr)).Interface()
}



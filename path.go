package piranhas

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unsafe"
)

var (
	errObjNotExists                      = errors.New("object belonging to the path could not be retrieved")
	errInvalidInput                      = errors.New("invalid input format")
	errPathToShort                       = errors.New("path is too short")
	errPathToLong                        = errors.New("path is too long")
	errWrongElementType                  = errors.New("element is of the wrong type")
	errIsNotInterfaceable                = errors.New("value is'nt usable for interfaces possibly due to not exported symbols")
	errEndSquareBracketsOpen             = errors.New("element ends without square brackets being closed")
	errNesstedSquareBracketsNotPermitted = errors.New("nested square brackets are not permitted")
	errLostCloseSquareBracket            = errors.New("closed square brackets found without being opened before")
	errUnknownEscChar                    = errors.New("unknown escape character")
	errControlChar                       = errors.New("characters smaller than space are not permitted in an element")
	errEscape2Chars                      = errors.New("escape mode requires a second character")
	errSpaceInElement                    = errors.New("space characters are not permitted in element")
	errEndQuotsOpen                      = errors.New("element ends without quotes being closed")
)

// parsePath parses a given path string and returns a slice of path elements
func parsePath(path string) ([]string, error) {
	// trim common prefixes and replace slashes/backslashes with dots
	path = strings.TrimSpace(path)
	path = strings.TrimPrefix(path, "$..")
	path = strings.TrimPrefix(path, "$.")
	patht := strings.TrimSpace(path)
	if patht == "" {
		return nil, nil
	}

	// initialize a slice to store path elements
	pathelements := make([]string, 0)
	element := ""
	inEscapeMode := false
	inQuotes := false
	inSquareBrackets := false

	// iterate over characters in the path
	for _, c := range path {
		switch c {
		case '.':
			if inEscapeMode {
				element += string(c)
				inEscapeMode = false
			} else if inQuotes {
				element += string(c)
			} else if inSquareBrackets {
				return nil, errEndSquareBracketsOpen
			} else if element != "" {
				pathelements = append(pathelements, element)
				element = ""
			}

		case '\\':
			if inEscapeMode {
				element += string(c)
				inEscapeMode = false
			} else if inQuotes {
				inEscapeMode = true
			} else if inSquareBrackets {
				return nil, errEndSquareBracketsOpen
			} else if element != "" {
				pathelements = append(pathelements, element)
				element = ""
			}
		case '/':
			if inEscapeMode {
				element += string(c)
				inEscapeMode = false
			} else if inQuotes {
				element += string(c)
			} else if inSquareBrackets {
				return nil, errEndSquareBracketsOpen
			} else if element != "" {
				pathelements = append(pathelements, element)
				element = ""
			}

		case '[':
			if inEscapeMode {
				element += string(c)
				inEscapeMode = false
			} else if inQuotes {
				element += string(c)
			} else if inSquareBrackets {
				return nil, errNesstedSquareBracketsNotPermitted
			} else {
				if element != "" {
					pathelements = append(pathelements, element)
					element = ""
				}
				inSquareBrackets = true
			}

		case ']':
			if inEscapeMode {
				element += string(c)
				inEscapeMode = false
			} else if inQuotes {
				element += string(c)
			} else if inSquareBrackets {
				if element != "" {
					pathelements = append(pathelements, element)
					element = ""
				}
				inSquareBrackets = false
			} else {
				return nil, errLostCloseSquareBracket
			}

		case '"':
			if inEscapeMode {
				element += string(c)
				inEscapeMode = false
			} else if inQuotes {
				inQuotes = false
			} else {
				inQuotes = true
			}

		default:
			if inEscapeMode {
				return nil, errUnknownEscChar
			} else if inQuotes {
				if c < ' ' {
					return nil, errControlChar
				} else {
					element += string(c)
				}
			} else {
				if c < ' ' {
					return nil, errControlChar
				} else if unicode.IsSpace(c) {
					return nil, errSpaceInElement
				} else {
					element += string(c)
				}
			}
		}
	}

	if inEscapeMode {
		return nil, errEscape2Chars

	} else if inQuotes {
		return nil, errEndQuotsOpen

	} else if inSquareBrackets {
		return nil, errEndSquareBracketsOpen
	} else {
		if element != "" {
			pathelements = append(pathelements, element)
			element = ""
		}
	}

	// remove empty elements
	nonEmptyElements := make([]string, 0)
	for _, e := range pathelements {
		if e != "" {
			nonEmptyElements = append(nonEmptyElements, e)
		}
	}

	return nonEmptyElements, nil
}

// returnPathElement processes a given reflect.Value and a slice of path elements.
// It traverses through the path elements, handling pointers, and extracts the requested value from the reflect.Value.
// It returns the extracted value or an error if the path is too long or if the value is not found.
func returnPathElement(objValue reflect.Value, pathelements []string) (interface{}, error) {
	// read all pointers away
	for {
		if objValue.Kind() == reflect.Ptr {
			if objValue.IsNil() {
				// if there are no more path elements, return nil (value not found)
				if len(pathelements) == 0 {
					return nil, nil
				}
				// otherwise, return an error (path is too long)
				return nil, errPathToLong
			}
			objValue = objValue.Elem()
		} else {
			// break the loop if objValue is not a pointer
			break
		}
	}

	// if there are no more path elements, return the interface representation of value
	if len(pathelements) == 0 {
		return getInterfaceOfValue(objValue)
	}

	// process the objValue based on its kind
	switch objValue.Kind() {
	case reflect.Struct, reflect.Slice, reflect.Array, reflect.Map:
		return getPathContainer(objValue, pathelements)

	default:
		return nil, errPathToLong
	}
}

// getPathContainer retrieves the container for a given path in the project
// It takes a path string as input and returns the corresponding container object
func getPathContainer(objValue reflect.Value, pathelements []string) (interface{}, error) {
	// check input
	if len(pathelements) == 0 {
		return nil, errPathToShort
	}

	// read all pointers away
	for {
		if objValue.Kind() == reflect.Ptr {
			if objValue.IsNil() {
				// if there are no more path elements, return nil (value not found)
				if len(pathelements) > 0 {
					return nil, nil
				}
				// otherwise, return an error (path is too long)
				return nil, errPathToLong
			}
			objValue = objValue.Elem()
		} else {
			// break the loop if objValue is not a pointer
			break
		}
	}

	var elemValue reflect.Value
	switch objValue.Kind() {
	case reflect.Struct:
		// search the specific field
		elemValue = objValue.FieldByName(pathelements[0])
		if !elemValue.IsValid() {
			return nil, errObjNotExists
		}

	case reflect.Slice, reflect.Array:
		// determine and check the index
		index, err := strconv.Atoi(pathelements[0])
		if err != nil || index < 0 || index >= objValue.Len() {
			return nil, errObjNotExists
		}

		// array element to index determine
		elemValue = objValue.Index(index)

	case reflect.Map:
		// determine the value for the key
		keyType := objValue.Type().Key()
		switch keyType.Kind() {
		case reflect.String:
			elemValue = objValue.MapIndex(reflect.ValueOf(pathelements[0]))

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			key, err := strconv.ParseInt(pathelements[0], 10, keyType.Bits())
			if err != nil {
				return reflect.Value{}, errObjNotExists
			}
			elemValue = objValue.MapIndex(reflect.ValueOf(key).Convert(keyType))

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			key, err := strconv.ParseUint(pathelements[0], 10, keyType.Bits())
			if err != nil {
				return reflect.Value{}, errObjNotExists
			}
			elemValue = objValue.MapIndex(reflect.ValueOf(key).Convert(keyType))

		case reflect.Float32, reflect.Float64:
			key, err := strconv.ParseFloat(pathelements[0], keyType.Bits())
			if err != nil {
				return reflect.Value{}, errObjNotExists
			}
			elemValue = objValue.MapIndex(reflect.ValueOf(key).Convert(keyType))

		case reflect.Bool:
			key, err := strconv.ParseBool(pathelements[0])
			if err != nil {
				return reflect.Value{}, errObjNotExists
			}
			elemValue = objValue.MapIndex(reflect.ValueOf(key))

		default:
			return reflect.Value{}, fmt.Errorf("unsupported key type: %s", keyType.Kind())
		}
		if !elemValue.IsValid() {
			return nil, errObjNotExists
		}

	default:
		return nil, errWrongElementType
	}

	// if there are more pathelements, then deepen, otherwise return this interface{}
	if len(pathelements) > 1 {
		return returnPathElement(elemValue, pathelements[1:])
	}

	return getInterfaceOfValue(elemValue)
}

// getInterfaceOfValue takes a reflect.Value and returns its corresponding interface{} value
// If the value is a function without parameters, it returns the function itself
// For other types, it attempts to convert the value to an interface{}
func getInterfaceOfValue(objValue reflect.Value) (interface{}, error) {
	// read all pointers away
	for {
		if objValue.Kind() == reflect.Ptr {
			if objValue.IsNil() {
				return nil, nil
			}
			objValue = objValue.Elem()
		} else {
			// break the loop if objValue is not a pointer
			break
		}
	}

	// handle the different kind
	switch objValue.Kind() {
	case reflect.Invalid:
		return nil, nil

	case reflect.String:
		return objValue.String(), nil

	case reflect.Bool:
		return objValue.Bool(), nil

	case reflect.Int:
		return int(objValue.Int()), nil

	case reflect.Int16:
		return int16(objValue.Int()), nil

	case reflect.Int32:
		return int32(objValue.Int()), nil

	case reflect.Int64:
		if objValue.Type().String() == "time.Duration" {
			return time.Duration(objValue.Int()), nil
		}
		return int64(objValue.Int()), nil

	case reflect.Uint:
		return uint(objValue.Uint()), nil

	case reflect.Uint8:
		return uint8(objValue.Uint()), nil

	case reflect.Uint16:
		return uint16(objValue.Uint()), nil

	case reflect.Uint32:
		return uint32(objValue.Uint()), nil

	case reflect.Uint64:
		return uint64(objValue.Uint()), nil

	case reflect.Float32:
		return float32(objValue.Float()), nil

	case reflect.Float64:
		return float64(objValue.Float()), nil

	case reflect.Complex64:
		return complex64(objValue.Complex()), nil

	case reflect.Complex128:
		return complex128(objValue.Complex()), nil

	case reflect.Struct:
		if objValue.Type().String() == "time.Time" {
			// get internal variables of time.Time
			wall := uint64(objValue.FieldByName("wall").Uint())
			ext := int64(objValue.FieldByName("ext").Int())
			location := objValue.FieldByName("loc")

			// create a new time.Time object and return it
			if location.IsNil() {
				return createTimeFromWallExtLoc(wall, ext, nil), nil
			} else {
				return createTimeFromWallExtLoc(wall, ext, (*time.Location)(unsafe.Pointer(location.Elem().UnsafeAddr()))), nil
			}
		}

	case reflect.Slice:
		// []byte means a byte slice
		if objValue.Type().Elem().Kind() == reflect.Uint8 {
			return objValue.Bytes(), nil
		}

	}

	// for everything that has not been dealt with up to this point
	if objValue.CanInterface() {
		return objValue.Interface(), nil
	} else {
		return nil, errIsNotInterfaceable
	}
}

// GetPathInterface retrieves the interface for a given path in the project
func GetPathInterface(obj interface{}, path string) (interface{}, error) {
	// convert the path into a list of path elements
	pathelements, err := parsePath(path)
	if err != nil {
		return nil, err
	}

	return returnPathElement(reflect.ValueOf(obj), pathelements)
}

// GetPathString returns the object addressed by the path as string
func GetPathString(ptr interface{}, path string) (string, error) {
	obj, err := GetPathInterface(ptr, path)
	if err != nil {
		return "", err
	}
	if obj == nil {
		return "", errObjNotExists
	}
	sobj, ok := obj.(string)
	if ok {
		return sobj, nil
	}

	return "", errors.New("object is not a string")
}

// GetPathBool returns the object addressed by the path as bool
func GetPathBool(ptr interface{}, path string) (bool, error) {
	obj, err := GetPathInterface(ptr, path)
	if err != nil {
		return false, err
	}
	if obj == nil {
		return false, errObjNotExists
	}
	bobj, ok := obj.(bool)
	if ok {
		return bobj, nil
	}

	return false, errors.New("object is not a bool")
}

// GetPathInt returns the object addressed by the path as int
func GetPathInt(ptr interface{}, path string) (int, error) {
	obj, err := GetPathInterface(ptr, path)
	if err != nil {
		return 0, err
	}
	if obj == nil {
		return 0, errObjNotExists
	}
	iobj, ok := obj.(int)
	if ok {
		return iobj, nil
	}

	return 0, errors.New("object is not a int")
}

// GetPathInt16 returns the object addressed by the path as int16
func GetPathInt16(ptr interface{}, path string) (int16, error) {
	obj, err := GetPathInterface(ptr, path)
	if err != nil {
		return 0, err
	}

	if obj == nil {
		return 0, errObjNotExists
	}
	iobj, ok := obj.(int16)
	if ok {
		return iobj, nil
	}

	return 0, errors.New("object is not a int16")
}

// GetPathInt32 returns the object addressed by the path as int32
func GetPathInt32(ptr interface{}, path string) (int32, error) {
	obj, err := GetPathInterface(ptr, path)
	if err != nil {
		return 0, err
	}

	if obj == nil {
		return 0, errObjNotExists
	}
	iobj, ok := obj.(int32)
	if ok {
		return iobj, nil
	}

	return 0, errors.New("object is not a int32")
}

// GetPathInt64 returns the object addressed by the path as int64
func GetPathInt64(ptr interface{}, path string) (int64, error) {
	obj, err := GetPathInterface(ptr, path)
	if err != nil {
		return 0, err
	}
	if obj == nil {
		return 0, errObjNotExists
	}
	iobj, ok := obj.(int64)
	if ok {
		return iobj, nil
	}

	return 0, errors.New("object is not a int64")
}

// GetPathUint returns the object addressed by the path as uint
func GetPathUint(ptr interface{}, path string) (uint, error) {
	obj, err := GetPathInterface(ptr, path)
	if err != nil {
		return 0, err
	}
	if obj == nil {
		return 0, errObjNotExists
	}
	iobj, ok := obj.(uint)
	if ok {
		return iobj, nil
	}

	return 0, errors.New("object is not a uint")
}

// GetPathUint8 returns the object addressed by the path as uint8
func GetPathUint8(ptr interface{}, path string) (uint8, error) {
	obj, err := GetPathInterface(ptr, path)
	if err != nil {
		return 0, err
	}

	if obj == nil {
		return 0, errObjNotExists
	}
	iobj, ok := obj.(uint8)
	if ok {
		return iobj, nil
	}

	return 0, errors.New("object is not a uint8")
}

// GetPathUint16 returns the object addressed by the path as uint16
func GetPathUint16(ptr interface{}, path string) (uint16, error) {
	obj, err := GetPathInterface(ptr, path)
	if err != nil {
		return 0, err
	}

	if obj == nil {
		return 0, errObjNotExists
	}
	iobj, ok := obj.(uint16)
	if ok {
		return iobj, nil
	}

	return 0, errors.New("object is not a uint16")
}

// GetPathUint32 returns the object addressed by the path as uint32
func GetPathUint32(ptr interface{}, path string) (uint32, error) {
	obj, err := GetPathInterface(ptr, path)
	if err != nil {
		return 0, err
	}

	if obj == nil {
		return 0, errObjNotExists
	}
	iobj, ok := obj.(uint32)
	if ok {
		return iobj, nil
	}

	return 0, errors.New("object is not a uint32")
}

// GetPathUint64 returns the object addressed by the path as uint64
func GetPathUint64(ptr interface{}, path string) (uint64, error) {
	obj, err := GetPathInterface(ptr, path)
	if err != nil {
		return 0, err
	}
	if obj == nil {
		return 0, errObjNotExists
	}
	iobj, ok := obj.(uint64)
	if ok {
		return iobj, nil
	}

	return 0, errors.New("object is not a uint64")
}

// GetPathFloat32 returns the object addressed by the path as float32
func GetPathFloat32(ptr interface{}, path string) (float32, error) {
	obj, err := GetPathInterface(ptr, path)
	if err != nil {
		return 0, err
	}

	if obj == nil {
		return 0, errObjNotExists
	}
	iobj, ok := obj.(float32)
	if ok {
		return iobj, nil
	}

	return 0, errors.New("object is not a float32")
}

// GetPathFloat64 returns the object addressed by the path as float64
func GetPathFloat64(ptr interface{}, path string) (float64, error) {
	obj, err := GetPathInterface(ptr, path)
	if err != nil {
		return 0, err
	}
	if obj == nil {
		return 0, errObjNotExists
	}
	iobj, ok := obj.(float64)
	if ok {
		return iobj, nil
	}

	return 0, errors.New("object is not a float64")
}

// GetPathComplex64 returns the object addressed by the path as complex64
func GetPathComplex64(ptr interface{}, path string) (complex64, error) {
	obj, err := GetPathInterface(ptr, path)
	if err != nil {
		return 0, err
	}
	if obj == nil {
		return 0, errObjNotExists
	}
	iobj, ok := obj.(complex64)
	if ok {
		return iobj, nil
	}

	return 0, errors.New("object is not a complex64")
}

// GetPathComplex128 returns the object addressed by the path as complex128
func GetPathComplex128(ptr interface{}, path string) (complex128, error) {
	obj, err := GetPathInterface(ptr, path)
	if err != nil {
		return 0, err
	}
	if obj == nil {
		return 0, errObjNotExists
	}
	iobj, ok := obj.(complex128)
	if ok {
		return iobj, nil
	}

	return 0, errors.New("object is not a complex128")
}

// GetPathByteSlice returns the object addressed by the path as []byte
func GetPathByteSlice(ptr interface{}, path string) ([]byte, error) {
	obj, err := GetPathInterface(ptr, path)
	if err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, errObjNotExists
	}
	bsobj, ok := obj.([]byte)
	if ok {
		return bsobj, nil
	}

	return nil, errors.New("object is not a []byte")
}

// GetPathTime returns the object addressed by the path as time.Time
func GetPathTime(ptr interface{}, path string) (time.Time, error) {
	obj, err := GetPathInterface(ptr, path)
	if err != nil {
		return time.Time{}, err
	}
	if obj == nil {
		return time.Time{}, errObjNotExists
	}
	bsobj, ok := obj.(time.Time)
	if ok {
		return bsobj, nil
	}

	return time.Time{}, errors.New("object is not a time.Time")
}

// GetPathDuration returns the object addressed by the path as time.Duration
func GetPathDuration(ptr interface{}, path string) (time.Duration, error) {
	obj, err := GetPathInterface(ptr, path)
	if err != nil {
		return 0, err
	}
	if obj == nil {
		return 0, errObjNotExists
	}
	bsobj, ok := obj.(time.Duration)
	if ok {
		return bsobj, nil
	}

	return 0, errors.New("object is not a time.Duration")
}

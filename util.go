package piranhas

import (
	"errors"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

// createTimeFromWallExtLoc creates a new time.Time object from the values wall, ext and loc
func createTimeFromWallExtLoc(wall uint64, ext int64, loc *time.Location) time.Time {
	var t time.Time
	*(*uint64)(unsafe.Pointer(&t)) = wall
	*(*int64)(unsafe.Pointer(uintptr(unsafe.Pointer(&t)) + unsafe.Sizeof(wall))) = ext
	*(*uintptr)(unsafe.Pointer(uintptr(unsafe.Pointer(&t)) + unsafe.Sizeof(wall) + unsafe.Sizeof(ext))) = uintptr(unsafe.Pointer(loc))
	return t
}

// parseComplex parses a string representation of a complex number and returns the corresponding complex128 value
// The string should be in the format "real+imagi" or "real-imagi", where "real" and "imag" are the real and imaginary parts of the complex number, respectively.
// Example "3.5+2.7i"
func parseComplex(s string) (complex128, error) {
	// remove spaces and check for empty string
	s = strings.ReplaceAll(s, " ", "")
	if s == "" {
		return 0, errors.New("empty string")
	}

	// separate the real and imaginary parts by the "+" sign
	parts := strings.Split(s, "+")
	if len(parts) != 2 {
		return 0, errors.New("invalid format, expects 'a+bi'")
	}

	// extract and analyze the real and imaginary parts
	realPart, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, err
	}

	// remove the "i" at the end of the imaginary part string
	imagPartStr := strings.TrimSuffix(parts[1], "i")
	imagPart, err := strconv.ParseFloat(imagPartStr, 64)
	if err != nil {
		return 0, err
	}

	// create the complex value
	result := complex(realPart, imagPart)
	return result, nil
}

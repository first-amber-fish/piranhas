package piranhas

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestSetDefaults(t *testing.T) {

	type person struct {
		passport
		address address
		name    string `default:"John"`
		age     int    `default:"30"`
		isMale  bool   `default:"true"`
	}

	type person1ok struct {
		name   string `default:"John"`
		age    int    `default:"30"`
		isMale bool   `default:"true"`
	}

	type person1 struct {
		name   string `default:"John"`
		age    int    `default:"30"`
		isMale bool   `default:"error"`
	}

	type address struct {
		street string `default:"Müllerstraße"`
		number int    `default:"400"`
		city   string `default:"Berlin"`
		ZIP    string `default:"10000"`
	}

	type passport struct {
		number string `default:"KI123"`
	}

	type person2 struct {
		passport
		address address
	}

	type person3 struct {
		addresssSlice []address
		addressmap    map[string]address
	}

	tests := []struct {
		name        string
		input       interface{}
		expected    interface{}
		expectedErr error
	}{

		{
			name: "Struct with slice and map",
			input: &person3{
				addresssSlice: []address{
					{},
					{},
				},
				addressmap: map[string]address{
					"Hauptwohnsitz":  {},
					"Nebenwohnsitz1": {},
				},
			},
			expected: &person3{
				addresssSlice: []address{
					{"Müllerstraße", 400, "Berlin", "10000"},
					{"Müllerstraße", 400, "Berlin", "10000"},
				},
				addressmap: map[string]address{
					"Hauptwohnsitz":  {"Müllerstraße", 400, "Berlin", "10000"},
					"Nebenwohnsitz1": {"Müllerstraße", 400, "Berlin", "10000"},
				},
			},
			expectedErr: nil,
		},

		{
			name:  "Struct with struct and embedded struct",
			input: &person2{},
			expected: &person2{
				passport: passport{number: "KI123"},
				address:  address{"Müllerstraße", 400, "Berlin", "10000"},
			},
			expectedErr: nil,
		},

		{
			name:        "Struct with string int and bool fields",
			input:       &person1ok{},
			expected:    &person1ok{"John", 30, true},
			expectedErr: nil,
		},
		{
			name:        "Struct with wrong bool value",
			input:       &person1{},
			expected:    &person1{"John", 30, false},
			expectedErr: errors.New("failed to parse default tag for field isMale: invalid syntax"),
		},
		{
			name:        "Empty slice",
			input:       &[]int{},
			expected:    &[]int{},
			expectedErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := SetDefaults(test.input)

			if !reflect.DeepEqual(test.input, test.expected) {
				t.Errorf("Expected: %+v, but got: %+v", test.expected, test.input)
			}

			if (err == nil && test.expectedErr != nil) || (err != nil && test.expectedErr == nil) {
				t.Errorf("Missmatch on expected error")
			}

			if err != nil && test.expectedErr != nil {
				if err.Error() != test.expectedErr.Error() {
					fmt.Printf("'%s'", err.Error())

					t.Errorf("Expected error: %v, but got: %v", test.expectedErr, err)
				}
			}
		})
	}
}

func TestSetDefaultsTimeDuration(t *testing.T) {
	dur, _ := time.ParseDuration("2h30m")

	type person struct {
		birthDate            time.Time     `default:"04.09.1990" layout:"02.01.2006"`
		concentrationAbility time.Duration `default:"2h30m"`
	}

	tests := []struct {
		name        string
		input       interface{}
		expected    interface{}
		expectedErr error
	}{
		{
			name:  "duration & time",
			input: &person{},
			expected: &person{
				birthDate:            time.Date(1990, time.September, 4, 0, 0, 0, 0, time.UTC),
				concentrationAbility: dur,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := SetDefaults(test.input)

			if !reflect.DeepEqual(test.input, test.expected) {
				t.Errorf("Expected: %+v, but got: %+v", test.expected, test.input)
			}

			if (err == nil && test.expectedErr != nil) || (err != nil && test.expectedErr == nil) {
				t.Errorf("Missmatch on expected error")
			}

			if err != nil && test.expectedErr != nil {
				if err.Error() != test.expectedErr.Error() {
					fmt.Printf("'%s'", err.Error())

					t.Errorf("Expected error: %v, but got: %v", test.expectedErr, err)
				}
			}
		})
	}
}

func TestSetDefaultsComplext(t *testing.T) {
	cmplx128 := complex(3.5, 2.7)
	cmplx64 := complex64(cmplx128)

	type structur struct {
		c128 complex128 `default:"3.5+2.7i"`
		c64  complex64  `default:"3.5+2.7i"`
	}

	tests := []struct {
		name        string
		input       interface{}
		expected    interface{}
		expectedErr error
	}{
		{
			name:  "duration & time",
			input: &structur{},
			expected: &structur{
				c128: cmplx128,
				c64:  cmplx64,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := SetDefaults(test.input)

			if !reflect.DeepEqual(test.input, test.expected) {
				t.Errorf("Expected: %+v, but got: %+v", test.expected, test.input)
			}

			if (err == nil && test.expectedErr != nil) || (err != nil && test.expectedErr == nil) {
				t.Errorf("Missmatch on expected error")
			}

			if err != nil && test.expectedErr != nil {
				if err.Error() != test.expectedErr.Error() {
					fmt.Printf("'%s'", err.Error())

					t.Errorf("Expected error: %v, but got: %v", test.expectedErr, err)
				}
			}
		})
	}
}

func TestSetDefaultsJson(t *testing.T) {
	type structurSlice struct {
		stringSlice    []string  `default:"[\"a\",\"b\"]"`
		stringSlicePtr *[]string `default:"[\"a\",\"b\"]"`
	}

	type structurMap struct {
		stringMapOfInt    map[string]int  `default:"{\"a\": 5,\"b\": 6}"`
		stringMapOfIntPtr *map[string]int `default:"{\"a\": 5,\"b\": 6}"`
	}

	type structurMapError struct {
		stringMapOfInt map[string]int `default:"{\"a\": 5},\"b\": 6}}"`
	}

	sl := []string{"a", "b"}
	ml := map[string]int{"a": 5, "b": 6}
	tests := []struct {
		name        string
		input       interface{}
		expected    interface{}
		expectedErr error
	}{
		{
			name:  "string slice defined by json default",
			input: &structurSlice{},
			expected: &structurSlice{
				stringSlice:    []string{"a", "b"},
				stringSlicePtr: &sl,
			},
		},
		{
			name:  "string map of int defined by json default",
			input: &structurMap{},
			expected: &structurMap{
				stringMapOfInt:    map[string]int{"a": 5, "b": 6},
				stringMapOfIntPtr: &ml,
			},
		},
		{
			name:  "string map of int defined with by json error",
			input: &structurMapError{},
			expected: &structurMapError{
				stringMapOfInt: make(map[string]int),
			},
			expectedErr: errors.New("failed to parse default tag for field stringMapOfInt: invalid character ',' after top-level value"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := SetDefaults(test.input)

			if err != nil {
				fmt.Printf("'%s'\n", err)
			}

			if err == nil && !reflect.DeepEqual(test.input, test.expected) {
				t.Errorf("Expected: %+v, but got: %+v", test.expected, test.input)
			}

			if (err == nil && test.expectedErr != nil) || (err != nil && test.expectedErr == nil) {
				t.Errorf("Missmatch on expected error")
			}

			if err != nil && test.expectedErr != nil {
				if err.Error() != test.expectedErr.Error() {
					fmt.Printf("'%s'", err.Error())

					t.Errorf("Expected error: %v, but got: %v", test.expectedErr, err)
				}
			}
		})
	}
}

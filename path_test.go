package piranhas

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestParsePath(t *testing.T) {
	tests := []struct {
		path     string
		expected []string
		err      error
	}{
		{
			path:     "$..foo.bar",
			expected: []string{"foo", "bar"},
			err:      nil,
		},
		{
			path:     "$.baz[0].qux",
			expected: []string{"baz", "0", "qux"},
			err:      nil,
		},
		{
			path:     "$..[[0].foo",
			expected: nil,
			err:      errNesstedSquareBracketsNotPermitted,
		},
		{
			path:     "$..0].foo",
			expected: nil,
			err:      errLostCloseSquareBracket,
		},
		{
			path:     "address[\"\\nstreet\"]",
			expected: nil,
			err:      errUnknownEscChar,
		},

		{
			path:     "\"foo\\",
			expected: nil,
			err:      errEscape2Chars,
		},
		{
			path:     "\"foo",
			expected: nil,
			err:      errEndQuotsOpen,
		},
	}

	for _, test := range tests {
		result, err := parsePath(test.path)

		if err != test.err {
			t.Errorf("Expected error: %v, but got: %v", test.err, err)
		}

		if !sliceEqual(result, test.expected) {
			t.Errorf("Expected %v, but got %v", test.expected, result)
		}
	}
}

func TestGetInterfaceOfValue(t *testing.T) {
	testCases := []struct {
		name     string
		input    reflect.Value
		expected interface{}
		err      error
	}{
		{
			name:     "String",
			input:    reflect.ValueOf("Hello"),
			expected: "Hello",
			err:      nil,
		},
		{
			name:     "Bool",
			input:    reflect.ValueOf(true),
			expected: true,
			err:      nil,
		},
		{
			name:     "Int",
			input:    reflect.ValueOf(42),
			expected: 42,
			err:      nil,
		},
		{
			name:     "ByteSlice",
			input:    reflect.ValueOf([]byte{65, 66, 67}),
			expected: []byte{65, 66, 67},
			err:      nil,
		},
		{
			name:     "SliceOfStruct",
			input:    reflect.ValueOf([]struct{ Name string }{{Name: "Alice"}, {Name: "Bob"}}),
			expected: []struct{ Name string }{{Name: "Alice"}, {Name: "Bob"}},
			err:      nil,
		},
		{
			name:     "Array",
			input:    reflect.ValueOf([3]int{1, 2, 3}),
			expected: [3]int{1, 2, 3},
			err:      nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := getInterfaceOfValue(tc.input)
			if err != nil && tc.err == nil {
				t.Errorf("Expected no error, but got: %v", err)
			}
			if err == nil && tc.err != nil {
				t.Errorf("Expected error: %v, but got none", tc.err)
			}
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected interface value: %v, but got: %v", tc.expected, result)
			}
		})
	}
}

type address struct {
	street string
	number int
	city   string
	ZIP    string
}

type passport struct {
	number string
}

type person struct {
	passport
	firstName            string
	lastName             *string
	age                  int
	developer            bool
	address              address
	adresses1            []address
	hobbys               map[string]int
	fingerprint          []byte
	birthDate            time.Time
	concentrationAbility time.Duration

	vint16      int16
	vint32      int32
	vint64      int64
	vuint       uint
	vuint8      uint8
	vuint16     uint16
	vuint32     uint32
	vuint64     uint64
	vfloat32    float32
	vfloat64    float64
	vcomplex64  complex64
	vcomplex128 complex128
}

func buildPersonData() *person {
	lastName := "Ranseier"
	cetLocation := time.FixedZone("CET", 1*60*60)
	p := person{
		passport: passport{"KI123"},
		firstName: "Karl",
		lastName:  &lastName,
		age:       58,
		developer: true,

		address: address{
			street: "Tellerstraße",
			number: 29,
			city:   "Berlin",
			ZIP:    "10553",
		},
		adresses1: []address{
			{
				street: "Müllerstr",
				number: 129,
				city:   "Berlin",
				ZIP:    "10487",
			},
			{
				street: "Kanzlerpaltz",
				number: 1,
				city:   "Berlin",
				ZIP:    "10000",
			},
		},
		hobbys:               map[string]int{"Motorcycle": 10, "Skydiving": 9, "Crochet": 0},
		fingerprint:          []byte{72, 101, 108, 108, 111},
		birthDate:            time.Date(1965, time.June, 9, 3, 0, 0, 0, cetLocation),
		concentrationAbility: 2*time.Hour + 35*time.Minute,

		vint16:      16,
		vint32:      15,
		vint64:      223,
		vuint:       789,
		vuint8:      8,
		vuint16:     16,
		vuint32:     32,
		vuint64:     64,
		vfloat32:    32.05,
		vfloat64:    64.05,
		vcomplex64:  complex(float32(3.2), float32(4.3)),
		vcomplex128: complex(3.2, 4.3),
	}

	return &p
}
func TestGetPathInterface(t *testing.T) {
	data := buildPersonData()

	tests := []struct {
		path     string
		expected interface{}
	}{

		{"number", "KI123"},
		{"concentrationAbility", 2*time.Hour + 35*time.Minute},
		{"birthDate", time.Date(1965, time.June, 9, 3, 0, 0, 0, time.FixedZone("CET", 1*60*60))},
		{"fingerprint", []byte{72, 101, 108, 108, 111}},
		{"firstName", "Karl"},
		{"lastName", "Ranseier"},
		{"developer", true},
		{"address.city", "Berlin"},
		{"address.street", "Tellerstraße"},
		{"adresses1.0.street", "Müllerstr"},
		{"adresses1.1.ZIP", "10000"},
		{"hobbys.Motorcycle", 10},
	}

	for _, test := range tests {
		result, err := GetPathInterface(data, test.path)
		if err != nil {
			t.Errorf("Error for path %s: %v", test.path, err)
			continue
		}

		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("For path %s, expected: %v, got: %v", test.path, test.expected, result)
		}
	}
}

func TestGetPathString(t *testing.T) {
	data := buildPersonData()

	tests := []struct {
		name     string
		ptr      interface{}
		path     string
		expected string
		err      error
	}{
		{
			name:     "Object is a string",
			ptr:      data,
			path:     "firstName",
			expected: "Karl",
			err:      nil,
		},
		{
			name:     "Object is not a string",
			ptr:      data,
			path:     "fingerprint",
			expected: "",
			err:      errors.New("object is not a string"),
		},
		{
			name:     "Object does not exist",
			ptr:      nil,
			path:     "",
			expected: "",
			err:      errObjNotExists,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := GetPathString(test.ptr, test.path)
			if err != nil && err.Error() != test.err.Error() {
				t.Errorf("Expected error: %v, but got: %v", test.err, err)
			}
			if result != test.expected {
				t.Errorf("Expected %s, but got %s", test.expected, result)
			}
		})
	}
}

func TestGetPathBool(t *testing.T) {
	data := buildPersonData()

	tests := []struct {
		name     string
		ptr      interface{}
		path     string
		expected bool
		err      error
	}{
		{
			name:     "Object is a bool",
			ptr:      data,
			path:     "developer",
			expected: true,
			err:      nil,
		},
		{
			name:     "Object is not a bool",
			ptr:      data,
			path:     "fingerprint",
			expected: false,
			err:      errors.New("object is not a bool"),
		},
		{
			name:     "Object does not exist",
			ptr:      nil,
			path:     "",
			expected: false,
			err:      errObjNotExists,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := GetPathBool(test.ptr, test.path)
			if err != nil && err.Error() != test.err.Error() {
				t.Errorf("Expected error: %v, but got: %v", test.err, err)
			}
			if result != test.expected {
				t.Errorf("Expected %v, but got %v", test.expected, result)
			}
		})
	}
}

func TestGetPathInt(t *testing.T) {
	data := buildPersonData()

	tests := []struct {
		name     string
		ptr      interface{}
		path     string
		expected int
		err      error
	}{
		{
			name:     "Object is a int",
			ptr:      data,
			path:     "age",
			expected: 58,
			err:      nil,
		},
		{
			name:     "Object is not a int",
			ptr:      data,
			path:     "fingerprint",
			expected: 0,
			err:      errors.New("object is not a int"),
		},
		{
			name:     "Object does not exist",
			ptr:      nil,
			path:     "",
			expected: 0,
			err:      errObjNotExists,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := GetPathInt(test.ptr, test.path)
			if err != nil && err.Error() != test.err.Error() {
				t.Errorf("Expected error: %v, but got: %v", test.err, err)
			}
			if result != test.expected {
				t.Errorf("Expected %v, but got %v", test.expected, result)
			}
		})
	}
}

func TestGetPathInt16(t *testing.T) {
	data := buildPersonData()

	tests := []struct {
		name     string
		ptr      interface{}
		path     string
		expected int16
		err      error
	}{
		{
			name:     "Object is a int16",
			ptr:      data,
			path:     "vint16",
			expected: 16,
			err:      nil,
		},
		{
			name:     "Object is not a int16",
			ptr:      data,
			path:     "fingerprint",
			expected: 0,
			err:      errors.New("object is not a int16"),
		},
		{
			name:     "Object does not exist",
			ptr:      nil,
			path:     "",
			expected: 0,
			err:      errObjNotExists,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := GetPathInt16(test.ptr, test.path)
			if err != nil && err.Error() != test.err.Error() {
				t.Errorf("Expected error: %v, but got: %v", test.err, err)
			}
			if result != test.expected {
				t.Errorf("Expected %v, but got %v", test.expected, result)
			}
		})
	}
}

func TestGetPathInt32(t *testing.T) {
	data := buildPersonData()

	tests := []struct {
		name     string
		ptr      interface{}
		path     string
		expected int32
		err      error
	}{
		{
			name:     "Object is a int32",
			ptr:      data,
			path:     "vint32",
			expected: 15,
			err:      nil,
		},
		{
			name:     "Object is not a int32",
			ptr:      data,
			path:     "fingerprint",
			expected: 0,
			err:      errors.New("object is not a int32"),
		},
		{
			name:     "Object does not exist",
			ptr:      nil,
			path:     "",
			expected: 0,
			err:      errObjNotExists,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := GetPathInt32(test.ptr, test.path)
			if err != nil && err.Error() != test.err.Error() {
				t.Errorf("Expected error: %v, but got: %v", test.err, err)
			}
			if result != test.expected {
				t.Errorf("Expected %v, but got %v", test.expected, result)
			}
		})
	}
}

func TestGetPathInt64(t *testing.T) {
	data := buildPersonData()

	tests := []struct {
		name     string
		ptr      interface{}
		path     string
		expected int64
		err      error
	}{
		{
			name:     "Object is a int64",
			ptr:      data,
			path:     "vint64",
			expected: 223,
			err:      nil,
		},
		{
			name:     "Object is not a int64",
			ptr:      data,
			path:     "fingerprint",
			expected: 0,
			err:      errors.New("object is not a int64"),
		},
		{
			name:     "Object does not exist",
			ptr:      nil,
			path:     "",
			expected: 0,
			err:      errObjNotExists,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := GetPathInt64(test.ptr, test.path)
			if err != nil && err.Error() != test.err.Error() {
				t.Errorf("Expected error: %v, but got: %v", test.err, err)
			}
			if result != test.expected {
				t.Errorf("Expected %v, but got %v", test.expected, result)
			}
		})
	}
}

func TestGetPathUint(t *testing.T) {
	data := buildPersonData()

	tests := []struct {
		name     string
		ptr      interface{}
		path     string
		expected uint
		err      error
	}{
		{
			name:     "Object is a uint",
			ptr:      data,
			path:     "vuint",
			expected: 789,
			err:      nil,
		},
		{
			name:     "Object is not a uint",
			ptr:      data,
			path:     "fingerprint",
			expected: 0,
			err:      errors.New("object is not a uint"),
		},
		{
			name:     "Object does not exist",
			ptr:      nil,
			path:     "",
			expected: 0,
			err:      errObjNotExists,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := GetPathUint(test.ptr, test.path)
			if err != nil && err.Error() != test.err.Error() {
				t.Errorf("Expected error: %v, but got: %v", test.err, err)
			}
			if result != test.expected {
				t.Errorf("Expected %v, but got %v", test.expected, result)
			}
		})
	}
}

func TestGetPathUint8(t *testing.T) {
	data := buildPersonData()

	tests := []struct {
		name     string
		ptr      interface{}
		path     string
		expected uint8
		err      error
	}{
		{
			name:     "Object is a uint8",
			ptr:      data,
			path:     "vuint8",
			expected: 8,
			err:      nil,
		},
		{
			name:     "Object is not a uint8",
			ptr:      data,
			path:     "fingerprint",
			expected: 0,
			err:      errors.New("object is not a uint8"),
		},
		{
			name:     "Object does not exist",
			ptr:      nil,
			path:     "",
			expected: 0,
			err:      errObjNotExists,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := GetPathUint8(test.ptr, test.path)
			if err != nil && err.Error() != test.err.Error() {
				t.Errorf("Expected error: %v, but got: %v", test.err, err)
			}
			if result != test.expected {
				t.Errorf("Expected %v, but got %v", test.expected, result)
			}
		})
	}
}

func TestGetPathUint16(t *testing.T) {
	data := buildPersonData()

	tests := []struct {
		name     string
		ptr      interface{}
		path     string
		expected uint16
		err      error
	}{
		{
			name:     "Object is a uint16",
			ptr:      data,
			path:     "vuint16",
			expected: 16,
			err:      nil,
		},
		{
			name:     "Object is not a uint16",
			ptr:      data,
			path:     "fingerprint",
			expected: 0,
			err:      errors.New("object is not a uint16"),
		},
		{
			name:     "Object does not exist",
			ptr:      nil,
			path:     "",
			expected: 0,
			err:      errObjNotExists,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := GetPathUint16(test.ptr, test.path)
			if err != nil && err.Error() != test.err.Error() {
				t.Errorf("Expected error: %v, but got: %v", test.err, err)
			}
			if result != test.expected {
				t.Errorf("Expected %v, but got %v", test.expected, result)
			}
		})
	}
}

func TestGetPathUint32(t *testing.T) {
	data := buildPersonData()

	tests := []struct {
		name     string
		ptr      interface{}
		path     string
		expected uint32
		err      error
	}{
		{
			name:     "Object is a uint32",
			ptr:      data,
			path:     "vuint32",
			expected: 32,
			err:      nil,
		},
		{
			name:     "Object is not a uint32",
			ptr:      data,
			path:     "fingerprint",
			expected: 0,
			err:      errors.New("object is not a uint32"),
		},
		{
			name:     "Object does not exist",
			ptr:      nil,
			path:     "",
			expected: 0,
			err:      errObjNotExists,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := GetPathUint32(test.ptr, test.path)
			if err != nil && err.Error() != test.err.Error() {
				t.Errorf("Expected error: %v, but got: %v", test.err, err)
			}
			if result != test.expected {
				t.Errorf("Expected %v, but got %v", test.expected, result)
			}
		})
	}
}

func TestGetPathUint64(t *testing.T) {
	data := buildPersonData()

	tests := []struct {
		name     string
		ptr      interface{}
		path     string
		expected uint64
		err      error
	}{
		{
			name:     "Object is a uint64",
			ptr:      data,
			path:     "vuint64",
			expected: 64,
			err:      nil,
		},
		{
			name:     "Object is not a uint64",
			ptr:      data,
			path:     "fingerprint",
			expected: 0,
			err:      errors.New("object is not a uint64"),
		},
		{
			name:     "Object does not exist",
			ptr:      nil,
			path:     "",
			expected: 0,
			err:      errObjNotExists,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := GetPathUint64(test.ptr, test.path)
			if err != nil && err.Error() != test.err.Error() {
				t.Errorf("Expected error: %v, but got: %v", test.err, err)
			}
			if result != test.expected {
				t.Errorf("Expected %v, but got %v", test.expected, result)
			}
		})
	}
}

func TestGetPathFloat32(t *testing.T) {
	data := buildPersonData()

	tests := []struct {
		name     string
		ptr      interface{}
		path     string
		expected float32
		err      error
	}{
		{
			name:     "Object is a float32",
			ptr:      data,
			path:     "vfloat32",
			expected: 32.05,
			err:      nil,
		},
		{
			name:     "Object is not a float32",
			ptr:      data,
			path:     "fingerprint",
			expected: 0,
			err:      errors.New("object is not a float32"),
		},
		{
			name:     "Object does not exist",
			ptr:      nil,
			path:     "",
			expected: 0,
			err:      errObjNotExists,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := GetPathFloat32(test.ptr, test.path)
			if err != nil && err.Error() != test.err.Error() {
				t.Errorf("Expected error: %v, but got: %v", test.err, err)
			}
			if result != test.expected {
				t.Errorf("Expected %v, but got %v", test.expected, result)
			}
		})
	}
}

func TestGetPathFloat64(t *testing.T) {
	data := buildPersonData()

	tests := []struct {
		name     string
		ptr      interface{}
		path     string
		expected float64
		err      error
	}{
		{
			name:     "Object is a float64",
			ptr:      data,
			path:     "vfloat64",
			expected: 64.05,
			err:      nil,
		},
		{
			name:     "Object is not a float64",
			ptr:      data,
			path:     "fingerprint",
			expected: 0,
			err:      errors.New("object is not a float64"),
		},
		{
			name:     "Object does not exist",
			ptr:      nil,
			path:     "",
			expected: 0,
			err:      errObjNotExists,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := GetPathFloat64(test.ptr, test.path)
			if err != nil && err.Error() != test.err.Error() {
				t.Errorf("Expected error: %v, but got: %v", test.err, err)
			}
			if result != test.expected {
				t.Errorf("Expected %v, but got %v", test.expected, result)
			}
		})
	}
}

func TestGetPathComplex64(t *testing.T) {
	data := buildPersonData()

	tests := []struct {
		name     string
		ptr      interface{}
		path     string
		expected complex64
		err      error
	}{
		{
			name:     "Object is a complex64",
			ptr:      data,
			path:     "vcomplex64",
			expected: complex(float32(3.2), float32(4.3)),
			err:      nil,
		},
		{
			name:     "Object is not a complex64",
			ptr:      data,
			path:     "fingerprint",
			expected: 0,
			err:      errors.New("object is not a complex64"),
		},
		{
			name:     "Object does not exist",
			ptr:      nil,
			path:     "",
			expected: 0,
			err:      errObjNotExists,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := GetPathComplex64(test.ptr, test.path)
			if err != nil && err.Error() != test.err.Error() {
				t.Errorf("Expected error: %v, but got: %v", test.err, err)
			}
			if result != test.expected {
				t.Errorf("Expected %v, but got %v", test.expected, result)
			}
		})
	}
}

func TestGetPathComplex128(t *testing.T) {
	data := buildPersonData()

	tests := []struct {
		name     string
		ptr      interface{}
		path     string
		expected complex128
		err      error
	}{
		{
			name:     "Object is a complex128",
			ptr:      data,
			path:     "vcomplex128",
			expected: complex(3.2, 4.3),
			err:      nil,
		},
		{
			name:     "Object is not a complex128",
			ptr:      data,
			path:     "fingerprint",
			expected: 0,
			err:      errors.New("object is not a complex128"),
		},
		{
			name:     "Object does not exist",
			ptr:      nil,
			path:     "",
			expected: 0,
			err:      errObjNotExists,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := GetPathComplex128(test.ptr, test.path)
			if err != nil && err.Error() != test.err.Error() {
				t.Errorf("Expected error: %v, but got: %v", test.err, err)
			}
			if result != test.expected {
				t.Errorf("Expected %v, but got %v", test.expected, result)
			}
		})
	}
}

func TestGetPathByteSlice(t *testing.T) {
	data := buildPersonData()

	tests := []struct {
		name     string
		ptr      interface{}
		path     string
		expected []byte
		err      error
	}{
		{
			name:     "Object is a []byte",
			ptr:      data,
			path:     "fingerprint",
			expected: []byte{72, 101, 108, 108, 111},
			err:      nil,
		},
		{
			name:     "Object is not a []byte",
			ptr:      data,
			path:     "age",
			expected: nil,
			err:      errors.New("object is not a []byte"),
		},
		{
			name:     "Object does not exist",
			ptr:      nil,
			path:     "",
			expected: nil,
			err:      errObjNotExists,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := GetPathByteSlice(test.ptr, test.path)
			if err != nil && err.Error() != test.err.Error() {
				t.Errorf("Expected error: %v, but got: %v", test.err, err)
			}
			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("Expected %v, but got %v", test.expected, result)
			}
		})
	}
}

func TestGetPathTime(t *testing.T) {
	data := buildPersonData()

	tests := []struct {
		name     string
		ptr      interface{}
		path     string
		expected time.Time
		err      error
	}{
		{
			name:     "Object is a time.Time",
			ptr:      data,
			path:     "birthDate",
			expected: time.Date(1965, time.June, 9, 3, 0, 0, 0, time.FixedZone("CET", 1*60*60)),
			err:      nil,
		},
		{
			name:     "Object is not a time.Time",
			ptr:      data,
			path:     "fingerprint",
			expected: time.Time{},
			err:      errors.New("object is not a time.Time"),
		},
		{
			name:     "Object does not exist",
			ptr:      nil,
			path:     "",
			expected: time.Time{},
			err:      errObjNotExists,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := GetPathTime(test.ptr, test.path)
			if err != nil && err.Error() != test.err.Error() {
				t.Errorf("Expected error: %v, but got: %v", test.err, err)
			}
			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("Expected %v, but got %v", test.expected, result)
			}
		})
	}
}

func TestGetPathDuration(t *testing.T) {
	data := buildPersonData()

	tests := []struct {
		name     string
		ptr      interface{}
		path     string
		expected time.Duration
		err      error
	}{
		{
			name:     "Object is a time.Duration",
			ptr:      data,
			path:     "concentrationAbility",
			expected: 2*time.Hour + 35*time.Minute,
			err:      nil,
		},
		{
			name:     "Object is not a time.Duration",
			ptr:      data,
			path:     "fingerprint",
			expected: 0,
			err:      errors.New("object is not a time.Duration"),
		},
		{
			name:     "Object does not exist",
			ptr:      nil,
			path:     "",
			expected: 0,
			err:      errObjNotExists,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := GetPathDuration(test.ptr, test.path)
			if err != nil && err.Error() != test.err.Error() {
				t.Errorf("Expected error: %v, but got: %v", test.err, err)
			}
			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("Expected %v, but got %v", test.expected, result)
			}
		})
	}
}

// helper function to compare slices
func sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

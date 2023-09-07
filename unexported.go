package piranhas

import (
	"reflect"
	"unsafe"
)

// getUnexportedField retrieves the value of an unexported field from a struct using reflection.
// It takes a reflect.Value representing the unexported field and returns its value as an interface{}.
// Note: This function works with unexported fields, which are fields with names starting with a lowercase letter,
//
//	and uses unsafe.Pointer to access memory directly. Use with caution and only when necessary.
func getUnexportedField(field reflect.Value) interface{} {
	// create a new reflect.Value pointing to the same memory address as the given field
	// this is done using the unsafe package to obtain a pointer to the field's memory.
	ptr := unsafe.Pointer(field.UnsafeAddr())
	newField := reflect.NewAt(field.Type(), ptr).Elem()

	// retrieve the value stored in the newly created reflect.Value and convert it to an interface{}.
	return newField.Interface()
}

// setUnexportedField updates the value of a given field with the provided value.
// It constructs a new reflect.Value of the same type as the field at the memory address of the field,
// then sets the new value to the provided value.
func setUnexportedField(field reflect.Value, value reflect.Value) {
	// create a new reflect.Value at the memory address of the field
	// with the same type as the field, then set its value to the provided value.

	reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Set(value)
}

// getPtrInterface converts a reflect.Value to an interface value. If the field is a pointer,
// it dereferences the pointer and creates a new interface value at the same address
func getPtrInterface(field reflect.Value) interface{} {
	// if the field is a pointer, dereference it
	if field.Kind() == reflect.Ptr {
		field = field.Elem()
	}

	// obtain the pointer address and create a new reflect.Value at that address
	ptr := reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Pointer()

	// create a new interface value using the type and pointer obtained above
	return reflect.NewAt(field.Type(), unsafe.Pointer(ptr)).Interface()
}

package types

import "context"

// Value is JSON value
type Value interface {
	// Invert returns inverted value/operation.
	// All type values are invertible.
	// Literals like string,number,... return itself as-is.
	// If argument type or arity is different, this returns non-nil error.
	Invert() (Value, error)

	// Eval returns a evaluated value.
	// Literals like string,number,... return itself as-is.
	Eval(ctx context.Context) (val Value, rollback func(), err error)

	// Type returns the type of this value.
	Type() Type
}

// Type is a type of expression
type Type uint

// ================ Null ================

// Null is JSON null struct
type Null struct{}

// NullType is JSON null type
const NullType Type = 1

// NullValue is the JSON null value
var NullValue = &Null{}

// Invert returns itself as-is.
func (v *Null) Invert() (Value, error) {
	return v, nil
}

// Eval returns itself as-is.
func (v *Null) Eval(context.Context) (val Value, rollback func(), err error) {
	return v, func() {}, nil
}

// Type returns the type.
func (v *Null) Type() Type {
	return NullType
}

// ================ Bool ================

// Bool is JSON boolean struct
type Bool struct {
	Value bool
}

// BoolType is JSON boolean type
const BoolType Type = 2

// Invert returns itself as-is. All literal types of JSON values are the same.
func (v *Bool) Invert() (Value, error) {
	return v, nil
}

// Eval returns itself as-is.
func (v *Bool) Eval(context.Context) (val Value, rollback func(), err error) {
	return v, func() {}, nil
}

// Type returns the type.
func (v *Bool) Type() Type {
	return BoolType
}

// TrueValue is the JSON true value
var TrueValue = &Bool{true}

// FalseValue is the JSON false value
var FalseValue = &Bool{false}

// ================ Number ================

// Number is JSON number struct
type Number struct {
	Value float64
}

// NumberType is JSON number struct
const NumberType Type = 3

// Invert returns itself as-is. All literal types of JSON values are the same.
func (v *Number) Invert() (Value, error) {
	return v, nil
}

// Eval returns itself as-is.
func (v *Number) Eval(context.Context) (val Value, rollback func(), err error) {
	return v, func() {}, nil
}

// Type returns the type.
func (v *Number) Type() Type {
	return NumberType
}

// ================ String ================

// String is JSON string struct
type String struct {
	Value string
}

// StringType is JSON string type
const StringType Type = 4

// Invert returns itself as-is. All literal types of JSON values are the same.
func (v *String) Invert() (Value, error) {
	return v, nil
}

// Eval returns itself as-is.
func (v *String) Eval(context.Context) (val Value, rollback func(), err error) {
	return v, func() {}, nil
}

// Type returns the type.
func (v *String) Type() Type {
	return StringType
}

// ================ Array ================

// Array is JSON array struct
type Array struct {
	Value []Value
}

// ArrayType is JSON array type
const ArrayType Type = 5

// Invert returns itself as-is. All literal types of JSON values are the same.
func (v *Array) Invert() (Value, error) {
	return v, nil
}

// Eval returns itself as-is.
func (v *Array) Eval(context.Context) (val Value, rollback func(), err error) {
	return v, func() {}, nil
}

// Type returns the type.
func (v *Array) Type() Type {
	return ArrayType
}

// ================ Object ================

// Object is JSON object struct
type Object struct {
	Value map[string]Value
}

// ObjectType is JSON object type
const ObjectType Type = 6

// Invert returns itself as-is. All literal types of JSON values are the same.
func (v *Object) Invert() (Value, error) {
	return v, nil
}

// Eval returns itself as-is.
func (v *Object) Eval(context.Context) (val Value, rollback func(), err error) {
	return v, func() {}, nil
}

// Type returns the type.
func (v *Object) Type() Type {
	return ObjectType
}

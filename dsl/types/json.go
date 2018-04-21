package types

import "context"

// ================ Null ================

// Null is JSON null struct
type Null struct{}

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
	return &NullType{}
}

// ================ Bool ================

// Bool is JSON boolean struct
type Bool struct {
	value bool
}

// NewBool creates Bool instance
func NewBool(value bool) *Bool {
	if value {
		return TrueValue
	}
	return FalseValue
}

// Value returns the holding internal value
func (v *Bool) Value() bool {
	return v.value
}

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
	return &BoolType{}
}

// TrueValue is the JSON true value
var TrueValue = &Bool{true}

// FalseValue is the JSON false value
var FalseValue = &Bool{false}

// ================ Number ================

// Number is JSON number struct
type Number struct {
	value float64
}

// Value returns the holding internal value
func (v *Number) Value() float64 {
	return v.value
}

// NewNumber creates Number instance
func NewNumber(value float64) *Number {
	return &Number{value: value}
}

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
	return &NumberType{}
}

// ================ String ================

// String is JSON string struct
type String struct {
	value string
}

// Value returns the holding internal value
func (v *String) Value() string {
	return v.value
}

// NewString creates String instance
func NewString(value string) *String {
	return &String{value: value}
}

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
	return &StringType{}
}

// ================ Array ================

// Array is JSON array struct
type Array struct {
	value   []Value
	argType Type
}

// Value returns the holding internal value.
// DO NOT CHANGE THE RETURN VALUE DIRECTLY!
// Copy the slice before changing the value.
func (v *Array) Value() []Value {
	return v.value
}

// NewArray creates Array instance
func NewArray(value []Value, argType Type) *Array {
	return &Array{value: value, argType: argType}
}

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
	return NewArrayType(v.argType)
}

// ================ Object ================

// Object is JSON object struct
type Object struct {
	value   map[string]Value
	argType Type
}

// Value returns the holding internal value.
// DO NOT CHANGE THE RETURN VALUE DIRECTLY!
// Copy the map instance before changing the value.
func (v *Object) Value() map[string]Value {
	return v.value
}

// NewObject creates Object instance
func NewObject(value map[string]Value, argType Type) *Object {
	return &Object{value: value, argType: argType}
}

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
	return NewObjectType(v.argType)
}

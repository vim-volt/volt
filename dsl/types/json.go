package types

import "context"

// ================ Null ================

// NullValue is the JSON null value
var NullValue = &nullT{}

type nullT struct{}

func (*nullT) Invert(context.Context) (Value, error) {
	return NullValue, nil
}

func (v *nullT) Eval(context.Context) (val Value, rollback func(), err error) {
	return v, func() {}, nil
}

func (*nullT) Type() Type {
	return NullType
}

// ================ Bool ================

// TrueValue is the JSON true value
var TrueValue = &boolT{true}

// FalseValue is the JSON false value
var FalseValue = &boolT{false}

// Bool is JSON boolean value
type Bool interface {
	Value

	// Value returns the holding internal value
	Value() bool
}

// NewBool creates Bool instance
func NewBool(value bool) Bool {
	if value {
		return TrueValue
	}
	return FalseValue
}

type boolT struct {
	value bool
}

func (v *boolT) Value() bool {
	return v.value
}

func (v *boolT) Invert(context.Context) (Value, error) {
	return v, nil
}

func (v *boolT) Eval(context.Context) (val Value, rollback func(), err error) {
	return v, func() {}, nil
}

func (*boolT) Type() Type {
	return BoolType
}

// ================ Number ================

// Number is JSON number value
type Number interface {
	Value

	// Value returns the holding internal value
	Value() float64
}

// NewNumber creates Number instance
func NewNumber(value float64) Number {
	return &numberT{value: value}
}

type numberT struct {
	value float64
}

func (v *numberT) Value() float64 {
	return v.value
}

func (v *numberT) Invert(context.Context) (Value, error) {
	return v, nil
}

func (v *numberT) Eval(context.Context) (val Value, rollback func(), err error) {
	return v, func() {}, nil
}

func (*numberT) Type() Type {
	return NumberType
}

// ================ String ================

// String is JSON string value
type String interface {
	Value

	// Value returns the holding internal value
	Value() string
}

// NewString creates String instance
func NewString(value string) String {
	return &stringT{value: value}
}

type stringT struct {
	value string
}

func (v *stringT) Value() string {
	return v.value
}

func (v *stringT) Invert(context.Context) (Value, error) {
	return v, nil
}

func (v *stringT) Eval(context.Context) (val Value, rollback func(), err error) {
	return v, func() {}, nil
}

func (*stringT) Type() Type {
	return StringType
}

// ================ Array ================

// Array is JSON array value
type Array interface {
	Value

	// Value returns the holding internal value.
	// DO NOT CHANGE THE RETURN VALUE DIRECTLY!
	// Copy the slice before changing the value.
	Value() []Value
}

// NewArray creates Array instance
func NewArray(value []Value, argType Type) Array {
	return &arrayT{value: value, argType: argType}
}

type arrayT struct {
	value   []Value
	argType Type
}

func (v *arrayT) Value() []Value {
	return v.value
}

func (v *arrayT) Invert(context.Context) (Value, error) {
	return v, nil
}

func (v *arrayT) Eval(context.Context) (val Value, rollback func(), err error) {
	return v, func() {}, nil
}

func (v *arrayT) Type() Type {
	return NewArrayType(v.argType)
}

// ================ Object ================

// Object is JSON object value
type Object interface {
	Value

	// Value returns the holding internal value.
	// DO NOT CHANGE THE RETURN VALUE DIRECTLY!
	// Copy the map instance before changing the value.
	Value() map[string]Value
}

// NewObject creates Object instance
func NewObject(value map[string]Value, argType Type) Object {
	return &objectT{value: value, argType: argType}
}

type objectT struct {
	value   map[string]Value
	argType Type
}

func (v *objectT) Value() map[string]Value {
	return v.value
}

func (v *objectT) Invert(context.Context) (Value, error) {
	return v, nil
}

func (v *objectT) Eval(context.Context) (val Value, rollback func(), err error) {
	return v, func() {}, nil
}

func (v *objectT) Type() Type {
	return NewObjectType(v.argType)
}

package types

// Type is a type of a value
type Type interface {
	// String returns a string like "<type %s>"
	String() string

	// InstanceOf checks has-a relation with its argument type
	InstanceOf(Type) bool
}

// ===================== Void type ===================== //

// VoidType is a void type
var VoidType = &voidType{}

type voidType struct{}

func (*voidType) String() string {
	return "Void"
}

func (*voidType) InstanceOf(t Type) bool {
	if _, ok := t.(*voidType); ok {
		return true
	}
	return false
}

// ===================== Null type ===================== //

// NullType is a null type
var NullType = &nullType{}

type nullType struct{}

func (*nullType) String() string {
	return "Null"
}

func (*nullType) InstanceOf(t Type) bool {
	if _, ok := t.(*nullType); ok {
		return true
	}
	return false
}

// ===================== Bool type ===================== //

// BoolType is a null type
var BoolType = &boolType{}

type boolType struct{}

func (*boolType) String() string {
	return "Bool"
}

func (*boolType) InstanceOf(t Type) bool {
	if _, ok := t.(*boolType); ok {
		return true
	}
	return false
}

// ===================== Number type ===================== //

// NumberType is a null type
var NumberType = &numberType{}

type numberType struct{}

func (*numberType) String() string {
	return "Number"
}

func (*numberType) InstanceOf(t Type) bool {
	if _, ok := t.(*numberType); ok {
		return true
	}
	return false
}

// ===================== String type ===================== //

// StringType is a null type
var StringType = &stringType{}

type stringType struct{}

func (*stringType) String() string {
	return "String"
}

func (*stringType) InstanceOf(t Type) bool {
	if _, ok := t.(*stringType); ok {
		return true
	}
	return false
}

// ===================== Array type ===================== //

// NewArrayType creates array type instance
func NewArrayType(arg Type) Type {
	return &arrayType{arg: arg}
}

type arrayType struct {
	arg Type
}

func (t *arrayType) String() string {
	return "Array[" + t.arg.String() + "]"
}

func (t *arrayType) InstanceOf(t2 Type) bool {
	if array, ok := t2.(*arrayType); ok {
		return t.arg.InstanceOf(array.arg)
	}
	return false
}

// ===================== Object type ===================== //

// NewObjectType creates object type instance
func NewObjectType(arg Type) Type {
	return &objectType{arg: arg}
}

type objectType struct {
	arg Type
}

func (t *objectType) String() string {
	return "Object[" + t.arg.String() + "]"
}

func (t *objectType) InstanceOf(t2 Type) bool {
	if array, ok := t2.(*objectType); ok {
		return t.arg.InstanceOf(array.arg)
	}
	return false
}

// ===================== Any type ===================== //

// AnyValue allows any type
var AnyValue = &anyType{}

type anyType struct{}

func (*anyType) String() string {
	return "Any"
}

func (*anyType) InstanceOf(_ Type) bool {
	return true
}

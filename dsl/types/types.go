package types

// Type is a type of expression
type Type uint

const (
	// VoidType is void type
	VoidType Type = iota
	// NullType is JSON null type
	NullType
	// BoolType is JSON boolean type
	BoolType
	// NumberType is JSON number struct
	NumberType
	// StringType is JSON string type
	StringType
	// ArrayType is JSON array type
	ArrayType
	// ObjectType is JSON object type
	ObjectType
)

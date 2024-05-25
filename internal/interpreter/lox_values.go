package interpreter

import (
	"github.com/leonardinius/golox/internal/parser"
)

// Value alias, not type redefinition.
type Value = parser.Value

type (
	ValueNil      struct{}
	ValueBool     bool
	ValueFloat    float64
	ValueString   string
	ValueCallable struct {
		Callable
	}
	ValueClass struct {
		*LoxClass
	}
	ValueObject struct {
		LoxObject
	}
)

var (
	NilValue               = ValueNil{}
	EmptyStringValue       = ValueString("")
	NilStringValue         = ValueString("nil")
	ErrNilNil        error = nil
)

// Type implements parser.Value.
func (v ValueNil) Type() parser.ValueType {
	return parser.ValueNilType
}

// Type implements parser.Value.
func (v ValueBool) Type() parser.ValueType {
	return parser.ValueBoolType
}

// Type implements parser.Value.
func (v ValueFloat) Type() parser.ValueType {
	return parser.ValueFloatType
}

// Type implements parser.Value.
func (v ValueString) Type() parser.ValueType {
	return parser.ValueStringType
}

// Type implements parser.Value.
func (v ValueCallable) Type() parser.ValueType {
	return parser.ValueCallableType
}

// Type implements parser.Value.
func (v ValueObject) Type() parser.ValueType {
	return parser.ValueObjectType
}

// Type implements parser.Value.
func (v ValueClass) Type() parser.ValueType {
	return parser.ValueClassType
}

var (
	_ Value = (*ValueNil)(nil)
	_ Value = ValueCallable{Callable: nil}
	_ Value = ValueBool(false)
	_ Value = ValueFloat(0)
	_ Value = ValueString("")
	_ Value = ValueClass{LoxClass: nil}
	_ Value = ValueObject{LoxObject: nil}
)

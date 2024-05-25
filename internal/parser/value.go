package parser

type ValueType uint

const (
	ValueNilType ValueType = iota
	ValueBoolType
	ValueFloatType
	ValueStringType
	ValueCallableType
	ValueClassType
	ValueObjectType
)

type Value interface {
	Type() ValueType
}

package interpreter

import (
	"fmt"
	"time"

	"github.com/leonardinius/golox/internal/loxerrors"
	"github.com/leonardinius/golox/internal/parser"
	"github.com/leonardinius/golox/internal/token"
)

func StdFnTime(interpeter *interpreter) (Value, error) {
	return ValueFloat(time.Now().UnixMilli()) / 1000.0, nil
}

func StdFnPPrint(interpeter *interpreter, args ...Value) (Value, error) {
	interpeter.print(args...)
	return nil, ErrNilNil
}

func StdFnCreateArray(interpeter *interpreter, arg Value) (Value, error) {
	var size int
	switch arg.Type() {
	case parser.ValueFloatType:
		size = int(arg.(ValueFloat))
	default:
		return nil, loxerrors.ErrRuntimeArrayInvalidArraySize
	}

	values := make([]Value, size)
	return ValueObject{NewStdArray(values)}, nil
}

type StdArray struct {
	values []Value
}

func NewStdArray(values []Value) *StdArray {
	return &StdArray{values: values}
}

// Get implements LoxInstance.
func (s *StdArray) Get(name *token.Token) (Value, error) {
	switch name.Lexeme {
	case "length":
		return ValueFloat(len(s.values)), nil
	case "get":
		return ValueCallable{NativeFunction1(func(interpeter *interpreter, arg1 Value) (Value, error) {
			return s.getAt(name, arg1)
		})}, nil
	case "set":
		return ValueCallable{NativeFunction2(func(interpeter *interpreter, arg1, arg2 Value) (Value, error) {
			return s.setAt(name, arg1, arg2)
		})}, nil
	}

	return nil, loxerrors.NewRuntimeError(name, loxerrors.ErrRuntimeUndefinedProperty(name.Lexeme))
}

// Set implements LoxInstance.
func (s *StdArray) Set(name *token.Token, value Value) (Value, error) {
	return nil, loxerrors.NewRuntimeError(name, loxerrors.ErrRuntimeArraysCantSetProperties)
}

func (s *StdArray) getAt(name *token.Token, index Value) (Value, error) {
	i, err := s.indexToInt(name, index)
	if err != nil {
		return nil, err
	}

	if i < 0 || i >= len(s.values) {
		return nil, loxerrors.NewRuntimeError(name, loxerrors.ErrRuntimeArrayIndexOutOfRange)
	}

	return s.values[i], nil
}

func (s *StdArray) setAt(name *token.Token, index, value Value) (Value, error) {
	i, err := s.indexToInt(name, index)
	if err != nil {
		return nil, err
	}

	if i < 0 || i >= len(s.values) {
		return nil, loxerrors.NewRuntimeError(name, loxerrors.ErrRuntimeArrayIndexOutOfRange)
	}

	s.values[i] = value
	return nil, ErrNilNil
}

func (s *StdArray) indexToInt(name *token.Token, index Value) (int, error) {
	if index.Type() == parser.ValueFloatType {
		return int(index.(ValueFloat)), nil
	}

	return 0, loxerrors.NewRuntimeError(name, loxerrors.ErrRuntimeArrayInvalidArrayIndex)
}

func (s *StdArray) String() string {
	return fmt.Sprintf("%v", s.values)
}

func (s *StdArray) GoString() string {
	return s.String()
}

var (
	_ LoxObject      = (*StdArray)(nil)
	_ fmt.Stringer   = (*StdArray)(nil)
	_ fmt.GoStringer = (*StdArray)(nil)
)

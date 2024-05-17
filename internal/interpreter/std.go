package interpreter

import (
	"context"
	"fmt"
	"time"

	"github.com/leonardinius/golox/internal/loxerrors"
	"github.com/leonardinius/golox/internal/token"
)

func StdFnTime(ctx context.Context, interpeter *interpreter) (any, error) {
	return float64(time.Now().UnixMilli()), nil
}

func StdFnPPrint(ctx context.Context, interpeter *interpreter, args ...any) (any, error) {
	interpeter.print(args...)
	return nil, nil
}

func StdFnCreateArray(ctx context.Context, interpeter *interpreter, arg any) (any, error) {
	var size int
	switch arg := arg.(type) {
	case int:
		size = arg
	case float64:
		size = int(arg)
	default:
		return nil, loxerrors.ErrRuntimeArrayInvalidArraySize
	}

	values := make([]any, size)
	return NewStdArray(values), nil
}

type StdArray struct {
	values []any
}

func NewStdArray(values []any) *StdArray {
	return &StdArray{values: values}
}

// Get implements LoxInstance.
func (s *StdArray) Get(ctx context.Context, name *token.Token) (any, error) {
	switch name.Lexeme {
	case "length":
		return float64(len(s.values)), nil
	case "get":
		return NativeFunction1(func(ctx context.Context, interpeter *interpreter, arg1 any) (any, error) {
			return s.getAt(name, arg1)
		}), nil
	case "set":
		return NativeFunction2(func(ctx context.Context, interpeter *interpreter, arg1, arg2 any) (any, error) {
			return s.setAt(name, arg1, arg2)
		}), nil
	}

	return nil, loxerrors.NewRuntimeError(name, loxerrors.ErrRuntimeUndefinedProperty(name.Lexeme))
}

// Set implements LoxInstance.
func (s *StdArray) Set(ctx context.Context, name *token.Token, value any) (any, error) {
	return nil, loxerrors.NewRuntimeError(name, loxerrors.ErrRuntimeArraysCantSetProperties)
}

func (s *StdArray) getAt(name *token.Token, index any) (any, error) {
	i, err := s.indexToInt(name, index)
	if err != nil {
		return nil, err
	}

	if i < 0 || i >= len(s.values) {
		return nil, loxerrors.NewRuntimeError(name, loxerrors.ErrRuntimeArrayIndexOutOfRange)
	}

	return s.values[i], nil
}

func (s *StdArray) setAt(name *token.Token, index, value any) (any, error) {
	i, err := s.indexToInt(name, index)
	if err != nil {
		return nil, err
	}

	if i < 0 || i >= len(s.values) {
		return nil, loxerrors.NewRuntimeError(name, loxerrors.ErrRuntimeArrayIndexOutOfRange)
	}

	s.values[i] = value
	return nil, nil
}

func (s *StdArray) indexToInt(name *token.Token, index any) (int, error) {
	switch index := index.(type) {
	case int:
		return index, nil
	case float64:
		return int(index), nil
	}

	return 0, loxerrors.NewRuntimeError(name, loxerrors.ErrRuntimeArrayInvalidArrayIndex)
}

func (s *StdArray) String() string {
	return fmt.Sprintf("%v", s.values)
}

func (s *StdArray) GoString() string {
	return s.String()
}

var _ LoxInstance = (*StdArray)(nil)
var _ fmt.Stringer = (*StdArray)(nil)
var _ fmt.GoStringer = (*StdArray)(nil)

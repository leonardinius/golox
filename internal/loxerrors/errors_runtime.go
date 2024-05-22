package loxerrors

import (
	"errors"
	"fmt"

	"github.com/leonardinius/golox/internal/token"
)

var (
	ErrRuntimeOperandMustBeNumber          = errors.New("Operand must be a number.")
	ErrRuntimeOperandsMustBeNumbers        = errors.New("Operands must be numbers.")
	ErrRuntimeOperandsMustNumbersOrStrings = errors.New("Operands must be two numbers or two strings.")
	ErrRuntimeUndefinedVariable            = errors.New("Undefined variable")
	ErrRuntimeCalleeMustBeCallable         = errors.New("Can only call functions and classes.")
	ErrRuntimeOnlyInstancesHaveProperties  = errors.New("Only instances have properties.")
	ErrRuntimeOnlyInstancesHaveFields      = errors.New("Only instances have fields.")
	ErrRuntimeSuperClassMustBeClass        = errors.New("Superclass must be a class.")
	ErrRuntimeArraysCantSetProperties      = errors.New("Can't set properties on arrays.")
	ErrRuntimeArrayIndexOutOfRange         = errors.New("Array index out of range.")
	ErrRuntimeArrayInvalidArrayIndex       = errors.New("Invalid array index, must be number.")
	ErrRuntimeArrayInvalidArraySize        = errors.New("Invalid array size, must be number.")
)

func ErrRuntimeCalleeArityError(expectedArity int, actualArity int) error {
	return fmt.Errorf("Expected %d arguments but got %d.", expectedArity, actualArity)
}

func ErrRuntimeUndefinedProperty(name string) error {
	return fmt.Errorf("Undefined property '%s'.", name)
}

func NewRuntimeError(tok *token.Token, cause error) error {
	return &RuntimeError{tok, cause}
}

type RuntimeError struct {
	tok   *token.Token
	cause error
}

// Error implements error.
func (r *RuntimeError) Error() string {
	return fmt.Sprintf("%v\n[line %d] in script", r.cause, r.tok.Line)
}

func (r *RuntimeError) Unwrap() error {
	return r.cause
}

var _ error = (*RuntimeError)(nil)
var _ unwrapInterface = (*RuntimeError)(nil)

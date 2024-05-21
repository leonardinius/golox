package loxerrors

import (
	"errors"
	"fmt"

	"github.com/leonardinius/golox/internal/token"
)

var (
	ErrParseError                                 = errors.New("Parse error.")
	ErrParseUnexpectedToken                       = errors.New("Expect expression.")
	ErrParseUnexpectedVariableName                = errors.New("Expect variable name.")
	ErrParseCantInitVarSelfReference              = errors.New("Can't read local variable in its own initializer.")
	ErrParseCantDuplicateVariableDefinition       = errors.New("Already a variable with this name in this scope.")
	ErrParseInvalidAssignmentTarget               = errors.New("Invalid assignment target.")
	ErrParseExpectedRightParenToken               = errors.New("Expected ')' after expression.")
	ErrParseExpectedLeftParentIfToken             = errors.New("Expected '(' after if.")
	ErrParseExpectedRightParentIfToken            = errors.New("Expected ')' after if condition.")
	ErrParseExpectedLeftParentWhileToken          = errors.New("Expected '(' after while.")
	ErrParseExpectedRightParentWhileToken         = errors.New("Expected ')' after condition.")
	ErrParseExpectedLeftParentForToken            = errors.New("Expected '(' after for.")
	ErrParseExpectedRightParentForToken           = errors.New("Expected ')' after for clauses.")
	ErrParseExpectedRightCurlyBlockToken          = errors.New("Expect '}' after block.")
	ErrParseExpectedSemicolonTokenAfterPrintValue = errors.New("Expect ';' after print value.")
	ErrParseExpectedSemicolonTokenAfterVar        = errors.New("Expect ';' after variable declaration.")
	ErrParseExpectedSemicolonAfterForLoopCond     = errors.New("Expect ';' after loop condition.")
	ErrParseExpectedSemicolonTokenAfterBreak      = errors.New("Expect ';' after 'break'.")
	ErrParseExpectedSemicolonTokenAfterContinue   = errors.New("Expect ';' after 'continue'.")
	ErrParseExpectedSemicolonTokenAfterReturn     = errors.New("Expect ';' after return value.")
	ErrParseReturnOutsideFunction                 = errors.New("Can't return from top-level code.")
	ErrParseUnexpectedParameterName               = errors.New("Expect parameter name.")
	ErrParseExpectedRightParentFunToken           = errors.New("Expect ')' after parameters.")
	ErrParseClassCantInheritFromItself            = errors.New("A class can't inherit from itself.")
	ErrParseBreakOutsideLoop                      = errors.New("Must be inside a loop to use 'break'.")
	ErrParseContinueOutsideLoop                   = errors.New("Must be inside a loop to use 'continue'.")
	ErrParseTooManyArguments                      = errors.New("Can't have more than 255 arguments.")
	ErrParseTooManyParameters                     = errors.New("Can't have more than 255 parameters.")
	ErrParseLocalVariableNotUsed                  = errors.New("Local variable is not used.")
	ErrParseExpectClassName                       = errors.New("Expect class name.")
	ErrParseExpectSuperClassName                  = errors.New("Expect super class name.")
	ErrParseExpectLeftCurlyBeforeClassBody        = errors.New("Expect '{' before class body.")
	ErrParseExpectRightCurlyAfterClassBody        = errors.New("Expect '}' after class body.")
	ErrParseExpectedPropertyNameAfterDot          = errors.New("Expect property name after '.'.")
	ErrParseThisOutsideClass                      = errors.New("Can't use 'this' outside of a class.")
	ErrParseCantReturnValueFromInitializer        = errors.New("Can't return a value from an initializer.")
	ErrParseExpectedDotAfterSuper                 = errors.New("Expect '.' after 'super'.")
	ErrParseExpectedSuperClassMethodName          = errors.New("Expect superclass method name.")
	ErrParseCantUseSuperOutsideClass              = errors.New("Can't use 'super' outside of a class.")
	ErrParseCantUseSuperInClassWithNoSuperclass   = errors.New("Can't use 'super' in a class with no superclass.")
	ErrParseCantUseSuperInClassMethod             = errors.New("Can't use 'super' in a static class method.")
)

func ErrParseExpectedIdentifierKindError(kind string) error {
	return fmt.Errorf("expect %s name.", kind)
}

func ErrParseExpectedLeftParenError(kind string) error {
	return fmt.Errorf("expect '(' after %s name.", kind)
}

func ErrParseExpectedLeftBraceFunToken(kind string) error {
	return fmt.Errorf("expect '{' before %s body.", kind)
}

func NewParseError(tok *token.Token, cause error) error {
	return &ParserError{tok: tok, cause: cause}
}

type ParserError struct {
	tok   *token.Token
	cause error
}

// Error implements error.
func (p *ParserError) Error() string {
	where := "at end"
	if p.tok.Type != token.EOF {
		where = fmt.Sprintf("at '%s'", p.tok.Lexeme)
	}
	return fmt.Sprintf("[line %d] Error %s: %v", p.tok.Line, where, p.cause)
}

func (p *ParserError) Unwrap() error {
	return p.cause
}

var _ error = (*ParserError)(nil)
var _ unwrapInterface = (*ParserError)(nil)

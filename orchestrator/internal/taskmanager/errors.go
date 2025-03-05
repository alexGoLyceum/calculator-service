package taskmanager

import "errors"

var (
	ErrEmptyExpression           = errors.New("invalid expression: empty input")
	ErrMissingOperator           = errors.New("invalid expression: must contain at least one operator")
	ErrInvalidCharacter          = errors.New("invalid character")
	ErrParenthesisIssue          = errors.New("invalid expression: mismatched or improperly placed parentheses")
	ErrNumberFormatIssue         = errors.New("invalid number: incorrect format")
	ErrOperatorIssue             = errors.New("invalid expression: consecutive or misplaced operators")
	ErrDivisionByZero            = errors.New("division by zero is not allowed")
	ErrInvalidExpressionStartEnd = errors.New("expression cannot start or end with an operator")
	ErrOperatorInsideParenthesis = errors.New("invalid expression: unary operators inside parentheses are not allowed")
)

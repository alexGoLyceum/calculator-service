package services

import "errors"

var (
	ErrEmptyExpression           = errors.New("empty input")
	ErrMissingOperator           = errors.New("must contain at least one operator")
	ErrInvalidCharacter          = errors.New("invalid character")
	ErrParenthesisIssue          = errors.New("mismatched or improperly placed parentheses")
	ErrNumberFormatIssue         = errors.New("incorrect number format")
	ErrOperatorIssue             = errors.New("consecutive or misplaced operators")
	ErrDivisionByZero            = errors.New("division by zero is not allowed")
	ErrInvalidExpressionStartEnd = errors.New("expression cannot start or end with an operator")
	ErrUnaryOperatorNotSupported = errors.New("unary operators are not supported")
	ErrInvalidExpression         = errors.New("invalid expression")

	ErrUnknownUserID        = errors.New("unknown user id")
	ErrUnknownExpressionsID = errors.New("unknown expressions id")
	ErrUnknownTaskID        = errors.New("unknown task id")
	ErrForbidden            = errors.New("you do not have access to this resource")

	ErrUserWithLoginAlreadyExists = errors.New("user with this login already exists")
	ErrUserNotFoundByLogin        = errors.New("user with this login does not exist")
	ErrInvalidLogin               = errors.New("login must be 3–32 characters long")
	ErrInvalidPassword            = errors.New("invalid password")
	ErrWeakPassword               = errors.New("password must contain upper and lower case letters, a digit, a special character, and be 8-20 characters long")
	ErrDatabaseUnavailable        = errors.New("database is unavailable")
)

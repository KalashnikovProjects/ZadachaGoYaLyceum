package my_errors

import "errors"

var (
	ExpressionValidateError = errors.New("expression validate error")
	StrangeSymbolsError     = errors.New("strange symbols error")
	AuthenticationError     = errors.New("authentication error")
)

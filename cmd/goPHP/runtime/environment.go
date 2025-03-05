package runtime

import (
	"GoPHP/cmd/goPHP/phpError"
	"GoPHP/cmd/goPHP/runtime/values"
)

type Environment interface {
	LookupConstant(constantName string) (values.RuntimeValue, phpError.Error)
	LookupVariable(variableName string) (values.RuntimeValue, phpError.Error)
}

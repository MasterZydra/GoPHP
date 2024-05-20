package interpreter

import (
	"GoPHP/cmd/goPHP/ast"
	"fmt"
	"math"
	"regexp"
	"slices"
	"strings"
)

func (interpreter *Interpreter) print(str string) {
	interpreter.result += str
}

func (interpreter *Interpreter) println(str string) {
	interpreter.print(str + "\n")
}

func (interpreter *Interpreter) processCondition(expr ast.IExpression, env *Environment) (IRuntimeValue, bool, error) {
	runtimeValue, err := interpreter.processStmt(expr, env)
	if err != nil {
		return runtimeValue, false, err
	}

	boolean, err := lib_boolval(runtimeValue)
	return runtimeValue, boolean, err
}

func (interpreter *Interpreter) lookupVariable(expr ast.IExpression, env *Environment) (IRuntimeValue, error) {
	variableName, err := interpreter.varExprToVarName(expr, env)
	if err != nil {
		return NewVoidRuntimeValue(), err
	}

	return env.lookupVariable(variableName)
}

// Convert a variable expression into the interpreted variable name
func (interpreter *Interpreter) varExprToVarName(expr ast.IExpression, env *Environment) (string, error) {
	switch expr.GetKind() {
	case ast.SimpleVariableExpr:
		variableNameExpr := ast.ExprToSimpleVarExpr(expr).GetVariableName()

		if variableNameExpr.GetKind() == ast.VariableNameExpr {
			return ast.ExprToVarNameExpr(variableNameExpr).GetVariableName(), nil
		}

		if variableNameExpr.GetKind() == ast.SimpleVariableExpr {
			variableName, err := interpreter.varExprToVarName(variableNameExpr, env)
			if err != nil {
				return "", err
			}
			runtimeValue, err := env.lookupVariable(variableName)
			if err != nil {
				interpreter.println(err.Error())
			}
			valueStr, err := lib_strval(runtimeValue)
			if err != nil {
				return "", err
			}
			return "$" + valueStr, nil
		}

		return "", fmt.Errorf("varExprToVarName - SimpleVariableExpr: Unsupported expression: %s", expr)
	default:
		return "", fmt.Errorf("varExprToVarName: Unsupported expression: %s", expr)
	}
}

// ------------------- MARK: RuntimeValue -------------------

func (interpreter *Interpreter) exprToRuntimeValue(expr ast.IExpression, env *Environment) (IRuntimeValue, error) {
	switch expr.GetKind() {
	case ast.ArrayLiteralExpr:
		elements := map[IRuntimeValue]IRuntimeValue{}
		for key, element := range ast.ExprToArrayLitExpr(expr).GetElements() {
			keyValue, err := interpreter.processStmt(key, env)
			if err != nil {
				return NewVoidRuntimeValue(), err
			}
			elementValue, err := interpreter.processStmt(element, env)
			if err != nil {
				return NewVoidRuntimeValue(), err
			}
			elements[keyValue] = elementValue
		}
		return NewArrayRuntimeValue(elements), nil
	case ast.BooleanLiteralExpr:
		return NewBooleanRuntimeValue(ast.ExprToBoolLitExpr(expr).GetValue()), nil
	case ast.IntegerLiteralExpr:
		return NewIntegerRuntimeValue(ast.ExprToIntLitExpr(expr).GetValue()), nil
	case ast.FloatingLiteralExpr:
		return NewFloatingRuntimeValue(ast.ExprToFloatLitExpr(expr).GetValue()), nil
	case ast.StringLiteralExpr:
		str := ast.ExprToStrLitExpr(expr).GetValue()
		// variable substitution
		if ast.ExprToStrLitExpr(expr).GetStringType() == ast.DoubleQuotedString {
			r, _ := regexp.Compile(`{\$[A-Za-z1-9_][A-Za-z0-9_]*(\[['A-Za-z]*\])?}`)
			matches := r.FindAllString(str, -1)
			for _, match := range matches {
				exprStr := "<?= " + match[1:len(match)-1] + ";"
				result, err := NewInterpreter(interpreter.request).process(exprStr, env)
				if err != nil {
					return NewVoidRuntimeValue(), err
				}
				str = strings.Replace(str, match, result, 1)
			}
		}
		return NewStringRuntimeValue(str), nil
	case ast.NullLiteralExpr:
		return NewNullRuntimeValue(), nil
	default:
		return NewVoidRuntimeValue(), fmt.Errorf("exprToRuntimeValue: Unsupported expression: %s", expr)
	}
}

func runtimeValueToValueType(valueType ValueType, runtimeValue IRuntimeValue) (IRuntimeValue, error) {
	switch valueType {
	case BooleanValue:
		boolean, err := lib_boolval(runtimeValue)
		return NewBooleanRuntimeValue(boolean), err
	case FloatingValue:
		floating, err := lib_floatval(runtimeValue)
		return NewFloatingRuntimeValue(floating), err
	case IntegerValue:
		integer, err := lib_intval(runtimeValue)
		return NewIntegerRuntimeValue(integer), err
	case StringValue:
		str, err := lib_strval(runtimeValue)
		return NewStringRuntimeValue(str), err
	default:
		return NewVoidRuntimeValue(), fmt.Errorf("runtimeValueToValueType: Unsupported runtime value: %s", valueType)
	}
}

// ------------------- MARK: calculation -------------------

func calculate(operand1 IRuntimeValue, operator string, operand2 IRuntimeValue) (IRuntimeValue, error) {
	resultType := VoidValue
	if slices.Contains([]string{"."}, operator) {
		resultType = StringValue
	} else {
		resultType = IntegerValue
		if operand1.GetType() == FloatingValue || operand2.GetType() == FloatingValue {
			resultType = FloatingValue
		}
	}

	var err error
	operand1, err = runtimeValueToValueType(resultType, operand1)
	if err != nil {
		return NewVoidRuntimeValue(), err
	}
	operand2, err = runtimeValueToValueType(resultType, operand2)
	if err != nil {
		return NewVoidRuntimeValue(), err
	}
	// TODO testing how PHP behavious: var_dump(1.0 + 2); var_dump(1 + 2.0); var_dump("1" + 2);
	// var_dump("1" + "2"); => int
	// var_dump("1" . 2); => str
	// type order "string" - "int" - "float"

	// Testen
	//   true + 2
	//   true && 3

	switch resultType {
	case IntegerValue:
		return calculateInteger(runtimeValToIntRuntimeVal(operand1), operator, runtimeValToIntRuntimeVal(operand2))
	case FloatingValue:
		return calculateFloating(runtimeValToFloatRuntimeVal(operand1), operator, runtimeValToFloatRuntimeVal(operand2))
	case StringValue:
		return calculateString(runtimeValToStrRuntimeVal(operand1), operator, runtimeValToStrRuntimeVal(operand2))
	default:
		return NewVoidRuntimeValue(), fmt.Errorf("calculate: Type \"%s\" not implemented", operator)
	}
}

func calculateFloating(operand1 IFloatingRuntimeValue, operator string, operand2 IFloatingRuntimeValue) (IFloatingRuntimeValue, error) {
	switch operator {
	case "+":
		return NewFloatingRuntimeValue(operand1.GetValue() + operand2.GetValue()), nil
	case "-":
		return NewFloatingRuntimeValue(operand1.GetValue() - operand2.GetValue()), nil
	case "*":
		return NewFloatingRuntimeValue(operand1.GetValue() * operand2.GetValue()), nil
	case "/":
		return NewFloatingRuntimeValue(operand1.GetValue() / operand2.GetValue()), nil
	case "**":
		return NewFloatingRuntimeValue(math.Pow(operand1.GetValue(), operand2.GetValue())), nil
	default:
		return NewFloatingRuntimeValue(0), fmt.Errorf("calculateInteger: Operator \"%s\" not implemented", operator)
	}
}

func calculateInteger(operand1 IIntegerRuntimeValue, operator string, operand2 IIntegerRuntimeValue) (IIntegerRuntimeValue, error) {
	switch operator {
	case "|":
		return NewIntegerRuntimeValue(operand1.GetValue() | operand2.GetValue()), nil
	case "&":
		return NewIntegerRuntimeValue(operand1.GetValue() & operand2.GetValue()), nil
	case "+":
		return NewIntegerRuntimeValue(operand1.GetValue() + operand2.GetValue()), nil
	case "-":
		return NewIntegerRuntimeValue(operand1.GetValue() - operand2.GetValue()), nil
	case "*":
		return NewIntegerRuntimeValue(operand1.GetValue() * operand2.GetValue()), nil
	case "/":
		return NewIntegerRuntimeValue(operand1.GetValue() / operand2.GetValue()), nil
	case "%":
		return NewIntegerRuntimeValue(operand1.GetValue() % operand2.GetValue()), nil
	case "**":
		return NewIntegerRuntimeValue(int64(math.Pow(float64(operand1.GetValue()), float64(operand2.GetValue())))), nil
	default:
		return NewIntegerRuntimeValue(0), fmt.Errorf("calculateInteger: Operator \"%s\" not implemented", operator)
	}
}

func calculateString(operand1 IStringRuntimeValue, operator string, operand2 IStringRuntimeValue) (IStringRuntimeValue, error) {
	switch operator {
	case ".":
		return NewStringRuntimeValue(operand1.GetValue() + operand2.GetValue()), nil
	default:
		return NewStringRuntimeValue(""), fmt.Errorf("calculateString: Operator \"%s\" not implemented", operator)
	}
}

// ------------------- MARK: comparison -------------------

func compare(lhs IRuntimeValue, operator string, rhs IRuntimeValue) (IBooleanRuntimeValue, error) {
	// Spec: https://phplang.org/spec/10-expressions.html#grammar-equality-expression

	// TODO compare - "==", "!=", "<>"
	// Spec: https://phplang.org/spec/10-expressions.html#grammar-equality-expression
	// Operator == represents value equality, operators != and <> are equivalent and represent value inequality.
	// For operators ==, !=, and <>, the operands of different types are converted and compared according to the same rules as in relational operators. Two objects of different types are always not equal.

	// Spec: https://phplang.org/spec/10-expressions.html#grammar-equality-expression
	// Operator === represents same type and value equality, or identity, comparison,
	// and operator !== represents the opposite of ===.
	// The values are considered identical if they have the same type and compare as equal, with the additional conditions below:
	//    When comparing two objects, identity operators check to see if the two operands are the exact same object,
	//    not two different objects of the same type and value.
	//    Arrays must have the same elements in the same order to be considered identical.
	//    Strings are identical if they contain the same characters, unlike value comparison operators no conversions are performed for numeric strings.
	if operator == "===" || operator == "!==" {
		result := lhs.GetType() == rhs.GetType()
		if result {
			switch lhs.GetType() {
			case BooleanValue:
				result = runtimeValToBoolRuntimeVal(lhs).GetValue() == runtimeValToBoolRuntimeVal(rhs).GetValue()
			case FloatingValue:
				result = runtimeValToFloatRuntimeVal(lhs).GetValue() == runtimeValToFloatRuntimeVal(rhs).GetValue()
			case IntegerValue:
				result = runtimeValToIntRuntimeVal(lhs).GetValue() == runtimeValToIntRuntimeVal(rhs).GetValue()
			case NullValue:
				result = true
			case StringValue:
				result = runtimeValToStrRuntimeVal(lhs).GetValue() == runtimeValToStrRuntimeVal(rhs).GetValue()
			default:
				return NewBooleanRuntimeValue(false), fmt.Errorf("compare: Runtime type %s for operator \"===\" not implemented", lhs.GetType())
			}
		}

		if operator == "!==" {
			return NewBooleanRuntimeValue(!result), nil
		} else {
			return NewBooleanRuntimeValue(result), nil
		}
	}

	return NewBooleanRuntimeValue(false), fmt.Errorf("compare: Operator \"%s\" not implemented", operator)
}

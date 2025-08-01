package interpreter

import (
	"GoPHP/cmd/goPHP/ast"
	"GoPHP/cmd/goPHP/common"
	"GoPHP/cmd/goPHP/common/os"
	"GoPHP/cmd/goPHP/phpError"
	"GoPHP/cmd/goPHP/runtime"
	"GoPHP/cmd/goPHP/runtime/stdlib/outputControl"
	"GoPHP/cmd/goPHP/runtime/stdlib/variableHandling"
	"GoPHP/cmd/goPHP/runtime/values"
	"math"
	GoOs "os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

// -------------------------------------- Common -------------------------------------- MARK: Common

func (interpreter *Interpreter) Print(str string) {
	if interpreter.outputBufferStack.Len() > 0 {
		interpreter.outputBufferStack.Get(interpreter.outputBufferStack.Len() - 1).Content += str
	} else {
		interpreter.WriteResult(str)
	}
}

func (interpreter *Interpreter) Println(str string) {
	interpreter.Print(str + os.EOL)
}

func (interpreter *Interpreter) WriteResult(str string) {
	interpreter.result += str
}

func (interpreter *Interpreter) flushOutputBuffers() {
	if interpreter.outputBufferStack.Len() == 0 {
		return
	}

	for interpreter.outputBufferStack.Len() > 0 {
		outputControl.ObEndFlush(runtime.NewContext(interpreter, nil, nil))
	}
}

func (interpreter *Interpreter) processCondition(expr ast.IExpression, env *Environment) (values.RuntimeValue, bool, phpError.Error) {
	runtimeValue, err := interpreter.processStmt(expr, env)
	if err != nil {
		return runtimeValue, false, err
	}

	boolean, err := variableHandling.BoolVal(runtimeValue)
	return runtimeValue, boolean, err
}

func (interpreter *Interpreter) lookupVariable(expr ast.IExpression, env *Environment) (values.RuntimeValue, phpError.Error) {
	variableName, err := interpreter.varExprToVarName(expr, env)
	if err != nil {
		return values.NewVoid(), err
	}

	runtimeValue, err := env.LookupVariable(variableName)
	if !interpreter.suppressWarning && err != nil {
		interpreter.PrintError(phpError.NewWarning("%s in %s", strings.TrimPrefix(err.Error(), "Warning: "), expr.GetPosition().ToPosString()))
	}
	return runtimeValue, nil
}

// Convert a variable expression into the interpreted variable name
func (interpreter *Interpreter) varExprToVarName(expr ast.IExpression, env *Environment) (string, phpError.Error) {
	switch expr.GetKind() {
	case ast.SimpleVariableExpr:
		variableNameExpr := expr.(*ast.SimpleVariableExpression).VariableName

		if variableNameExpr.GetKind() == ast.VariableNameExpr {
			return variableNameExpr.(*ast.VariableNameExpression).VariableName, nil
		}

		if variableNameExpr.GetKind() == ast.SimpleVariableExpr {
			variableName, err := interpreter.varExprToVarName(variableNameExpr, env)
			if err != nil {
				return "", err
			}
			runtimeValue, err := env.LookupVariable(variableName)
			if err != nil {
				interpreter.PrintError(err)
			}
			valueStr, err := variableHandling.StrVal(runtimeValue)
			if err != nil {
				return "", err
			}
			return "$" + valueStr, nil
		}

		variableName, err := interpreter.processStmt(variableNameExpr, env)
		if err != nil {
			return "", err
		}
		valueStr, err := variableHandling.StrVal(variableName)
		if err != nil {
			return "", err
		}
		return "$" + valueStr, nil
	case ast.SubscriptExpr:
		return interpreter.varExprToVarName(expr.(*ast.SubscriptExpression).Variable, env)
	default:
		return "", phpError.NewError("varExprToVarName: Unsupported expression: %s", ast.ToString(expr))
	}
}

func (interpreter *Interpreter) ErrorToString(err phpError.Error) string {
	if (err.GetErrorType() == phpError.WarningPhpError && interpreter.ini.GetInt("error_reporting")&phpError.E_WARNING == 0) ||
		(err.GetErrorType() == phpError.ErrorPhpError && interpreter.ini.GetInt("error_reporting")&phpError.E_ERROR == 0) ||
		(err.GetErrorType() == phpError.NoticePhpError && interpreter.ini.GetInt("error_reporting")&phpError.E_NOTICE == 0) ||
		(err.GetErrorType() == phpError.ParsePhpError && interpreter.ini.GetInt("error_reporting")&phpError.E_PARSE == 0) {
		return ""
	}
	return err.GetMessage()
}

func (interpreter *Interpreter) PrintError(err phpError.Error) {
	if errStr := interpreter.ErrorToString(err); errStr == "" {
		return
	} else {
		interpreter.Println("")
		interpreter.Println(errStr)
	}
}

// Scan and process program for function definitions on root level and in compound statements
func (interpreter *Interpreter) scanForFunctionDefinition(statements []ast.IStatement, env *Environment) phpError.Error {
	for _, stmt := range statements {
		if stmt.GetKind() == ast.CompoundStmt {
			interpreter.scanForFunctionDefinition(stmt.(*ast.CompoundStatement).Statements, env)
			continue
		}

		if stmt.GetKind() != ast.FunctionDefinitionStmt {
			continue
		}

		_, err := interpreter.processStmt(stmt, env)
		if err != nil {
			return err
		}
	}
	return nil
}

var literalExprTypeRuntimeValue = map[ast.NodeType]string{
	ast.ArrayLiteralExpr:   "array",
	ast.IntegerLiteralExpr: "int",
	ast.StringLiteralExpr:  "string",
}

func literalExprTypeToRuntimeValue(expr ast.IExpression) (string, phpError.Error) {
	typeStr, found := literalExprTypeRuntimeValue[expr.GetKind()]
	if !found {
		return "", phpError.NewError("literalExprTypeToRuntimeValue: No mapping for type %s", expr.GetKind())
	}
	return typeStr, nil
}

func checkParameterTypes(runtimeValue values.RuntimeValue, expectedTypes []string) phpError.Error {
	typeStr := values.ToPhpType(runtimeValue)
	if typeStr == "" {
		return phpError.NewError("checkParameterTypes: No mapping for type %s", runtimeValue.GetType())
	}

	for _, expectedType := range expectedTypes {
		if expectedType == "mixed" {
			return nil
		}

		if typeStr == expectedType {
			return nil
		}
	}
	return phpError.NewError("Types do not match")
}

func (interpreter *Interpreter) includeFile(filepathExpr ast.IExpression, env *Environment, include bool, once bool) (values.RuntimeValue, phpError.Error) {
	runtimeValue, err := interpreter.processStmt(filepathExpr, env)
	if err != nil {
		return runtimeValue, err
	}
	if runtimeValue.GetType() == values.NullValue {
		return runtimeValue, phpError.NewError("Uncaught ValueError: Path cannot be empty in %s", filepathExpr.GetPosition().ToPosString())
	}

	filename, err := variableHandling.StrVal(runtimeValue)
	if err != nil {
		return runtimeValue, err
	}

	// Spec: https://phplang.org/spec/10-expressions.html#the-require-operator
	// Once an include file has been included, a subsequent use of require_once on that include file
	// results in a return value of TRUE but nothing else happens.
	if once && slices.Contains(interpreter.includedFiles, filename) && !os.IS_WIN {
		return values.NewBool(true), nil
	}
	if once && slices.Contains(interpreter.includedFiles, strings.ToLower(filename)) && os.IS_WIN {
		return values.NewBool(true), nil
	}

	absFilename := filename
	if !common.IsAbsPath(filename) {
		absFilename = common.GetAbsPathForWorkingDir(common.ExtractPath(filepathExpr.GetPosition().File.Filename), filename)
	}

	var functionName string
	if include {
		functionName = "include"
	} else {
		functionName = "require"
	}

	// Spec: https://phplang.org/spec/10-expressions.html#the-require-operator
	// This operator is identical to operator include except that in the case of require,
	// failure to find/open the designated include file produces a fatal error.
	getError := func() (values.RuntimeValue, phpError.Error) {
		if include {
			return values.NewVoid(), phpError.NewWarning(
				"%s(): Failed opening '%s' for inclusion (include_path='%s') in %s",
				functionName, filename, common.ExtractPath(filepathExpr.GetPosition().File.Filename), filepathExpr.GetPosition().ToPosString(),
			)
		} else {
			return values.NewVoid(), phpError.NewError(
				"Uncaught Error: Failed opening required '%s' (include_path='%s') in %s",
				filename, common.ExtractPath(filepathExpr.GetPosition().File.Filename), filepathExpr.GetPosition().ToPosString(),
			)
		}
	}

	if !common.PathExists(absFilename) {
		interpreter.PrintError(phpError.NewWarning(
			"%s(%s): Failed to open stream: No such file or directory in %s",
			functionName, filename, filepathExpr.GetPosition().ToPosString(),
		))
		return getError()
	}

	content, fileErr := GoOs.ReadFile(absFilename)
	if fileErr != nil {
		return getError()
	}
	program, parserErr := interpreter.parser.ProduceAST(string(content), absFilename)

	if !os.IS_WIN {
		interpreter.includedFiles = append(interpreter.includedFiles, absFilename)
	} else {
		interpreter.includedFiles = append(interpreter.includedFiles, strings.ToLower(absFilename))
	}
	if parserErr != nil {
		return runtimeValue, parserErr
	}
	return interpreter.processProgram(program, env)
}

func GetExecutableCreationDate() time.Time {
	exePath, err := GoOs.Executable()
	if err != nil {
		return time.Now()
	}
	info, err := GoOs.Stat(filepath.Clean(exePath))
	if err != nil {
		// println(err.Error())

		return time.Now()
	}
	return info.ModTime()
}

func (interpreter *Interpreter) GetClass(class string) (*ast.ClassDeclarationStatement, bool) {
	classDeclaration, found := interpreter.classDeclarations[class]
	if !found {
		return nil, false
	}
	return classDeclaration, true
}

// -------------------------------------- Caching -------------------------------------- MARK: Caching

func (interpreter *Interpreter) isCached(stmt ast.IStatement) bool {
	_, found := interpreter.cache[stmt.GetId()]
	return found
}

func (interpreter *Interpreter) writeCache(stmt ast.IStatement, value values.RuntimeValue) values.RuntimeValue {
	interpreter.cache[stmt.GetId()] = value
	return value
}

// -------------------------------------- RuntimeValue -------------------------------------- MARK: RuntimeValue

func (interpreter *Interpreter) exprToRuntimeValue(expr ast.IExpression, env *Environment) (values.RuntimeValue, phpError.Error) {
	switch expr.GetKind() {
	case ast.ArrayLiteralExpr:
		Array := values.NewArray()
		for _, key := range expr.(*ast.ArrayLiteralExpression).Keys {
			var keyValue values.RuntimeValue
			var err phpError.Error
			if key.GetKind() != ast.ArrayNextKeyExpr {
				keyValue, err = interpreter.processStmt(key, env)
				if err != nil {
					return values.NewVoid(), err
				}
			}
			elementValue, err := interpreter.processStmt(expr.(*ast.ArrayLiteralExpression).Elements[key], env)
			if err != nil {
				return values.NewVoid(), err
			}
			if err = Array.SetElement(keyValue, elementValue); err != nil {
				return values.NewVoid(), err
			}
		}
		return Array, nil
	case ast.IntegerLiteralExpr:
		return values.NewInt(expr.(*ast.IntegerLiteralExpression).Value), nil
	case ast.FloatingLiteralExpr:
		return values.NewFloat(expr.(*ast.FloatingLiteralExpression).Value), nil
	case ast.StringLiteralExpr:
		str := expr.(*ast.StringLiteralExpression).Value
		strType := expr.(*ast.StringLiteralExpression).StringType
		if strType == ast.DoubleQuotedString || strType == ast.HeredocString {
			// Supported expression: variable substitution: `echo "{$a}";`
			// variable substitution
			// TODO improve variable substitution: Regex and replace will not work for every case here. A parser is required that searches for variables, subscriptExpr, ... and resolves them.
			// TODO improve variable substitution to detect if a $ is escaped. E.g. "\$i"
			// TODO improve variable substitution to accept nested arrays "$a[$b['c']][0]"
			r, _ := regexp.Compile(`({\$[A-Za-z_][A-Za-z0-9_]*['A-Za-z0-9\[\]]*[^}]*})|(\$[A-Za-z_][A-Za-z0-9_]*['A-Za-z0-9\[\]]*)`)
			matches := r.FindAllString(str, -1)
			for _, match := range matches {
				varExpr := match
				if match[0] == '{' {
					// Remove curly braces
					varExpr = match[1 : len(match)-1]
				}
				// TODO Find better solution for code evaluation
				exprStr := "<?= " + varExpr + ";"
				interp, err := NewInterpreter(interpreter.ini, interpreter.request, "__file_name__")
				if err != nil {
					return values.NewVoid(), err
				}
				result, err := interp.process(exprStr, env, true)
				if err != nil {
					return values.NewVoid(), err
				}
				// TODO Improve bad fix so that the warning is above the possible string output
				if strings.Contains(result, "Warning: Undefined variable") {
					filenameRegex := regexp.MustCompile(`__file_name__:\d+:\d+`)
					interpreter.PrintError(
						phpError.NewWarning(
							"%s", strings.TrimPrefix(
								strings.TrimSpace(
									filenameRegex.ReplaceAllString(result, expr.GetPosition().ToPosString())),
								"Warning: ")))
					result = ""
				}
				str = strings.Replace(str, match, result, 1)
			}

			// unicode escape sequence
			r, _ = regexp.Compile(`\\u\{[0-9a-fA-F]+\}`)
			matches = r.FindAllString(str, -1)
			for _, match := range matches {
				unicodeChar, err := strconv.ParseInt(match[3:len(match)-1], 16, 32)
				if err != nil {
					return values.NewVoid(), phpError.NewError("%s", err.Error())
				}
				str = strings.Replace(str, match, string(rune(unicodeChar)), 1)
			}
		}
		return values.NewStr(str), nil
	default:
		return values.NewVoid(), phpError.NewError("exprToRuntimeValue: Unsupported expression: %s", expr)
	}
}

// -------------------------------------- inc-dec-calculation -------------------------------------- MARK: inc-dec-calculation

func calculateIncDec(operator string, operand values.RuntimeValue) (values.RuntimeValue, phpError.Error) {
	switch operand.GetType() {
	case values.BoolValue:
		// Spec: https://phplang.org/spec/10-expressions.html#prefix-increment-and-decrement-operators
		// For a prefix ++ or -- operator used with a Boolean-valued operand, there is no side effect, and the result is the operand’s value.
		return operand, nil
	case values.FloatValue:
		return calculateIncDecFloating(operator, operand.(*values.Float))
	case values.IntValue:
		return calculateIncDecInteger(operator, operand.(*values.Int))
	case values.NullValue:
		return calculateIncDecNull(operator)
	case values.StrValue:
		return calculateIncDecString(operator, operand.(*values.Str))
	default:
		return values.NewVoid(), phpError.NewError("calculateIncDec: Type \"%s\" not implemented", operand.GetType())
	}

	// TODO calculateIncDec - object
	// Spec: https://phplang.org/spec/10-expressions.html#prefix-increment-and-decrement-operators
	// If the operand has an object type supporting the operation, then the object semantics defines the result. Otherwise, the operation has no effect and the result is the operand.
}

func calculateIncDecInteger(operator string, operand *values.Int) (values.RuntimeValue, phpError.Error) {
	switch operator {
	case "++":
		// Spec: https://phplang.org/spec/10-expressions.html#prefix-increment-and-decrement-operators
		//For a prefix "++" operator used with an arithmetic operand, the side effect of the operator is to increment the value of the operand by 1.
		// The result is the value of the operand after it has been incremented.
		// If an int operand’s value is the largest representable for that type, the operand is incremented as if it were float.

		// Spec: https://phplang.org/spec/10-expressions.html#prefix-increment-and-decrement-operators
		// For a prefix ++ or -- operator used with an operand having the value INF, -INF, or NAN, there is no side effect, and the result is the operand’s value.
		return calculateInteger(operand, "+", values.NewInt(1))

	case "--":
		// Spec: https://phplang.org/spec/10-expressions.html#prefix-increment-and-decrement-operators
		// For a prefix "--" operator used with an arithmetic operand, the side effect of the operator is to decrement the value of the operand by 1.
		// The result is the value of the operand after it has been decremented.
		// If an int operand’s value is the smallest representable for that type, the operand is decremented as if it were float.

		// Spec: https://phplang.org/spec/10-expressions.html#prefix-increment-and-decrement-operators
		// For a prefix ++ or -- operator used with an operand having the value INF, -INF, or NAN, there is no side effect, and the result is the operand’s value.
		return calculateInteger(operand, "-", values.NewInt(1))

	default:
		return values.NewInt(0), phpError.NewError("calculateIncDecInteger: Operator \"%s\" not implemented", operator)
	}
}

func calculateIncDecFloating(operator string, operand *values.Float) (values.RuntimeValue, phpError.Error) {
	switch operator {
	case "++":
		// Spec: https://phplang.org/spec/10-expressions.html#prefix-increment-and-decrement-operators
		//For a prefix "++" operator used with an arithmetic operand, the side effect of the operator is to increment the value of the operand by 1.
		// The result is the value of the operand after it has been incremented.
		// If an int operand’s value is the largest representable for that type, the operand is incremented as if it were float.

		// Spec: https://phplang.org/spec/10-expressions.html#prefix-increment-and-decrement-operators
		// For a prefix ++ or -- operator used with an operand having the value INF, -INF, or NAN, there is no side effect, and the result is the operand’s value.
		return calculateFloating(operand, "+", values.NewFloat(1))

	case "--":
		// Spec: https://phplang.org/spec/10-expressions.html#prefix-increment-and-decrement-operators
		// For a prefix "--" operator used with an arithmetic operand, the side effect of the operator is to decrement the value of the operand by 1.
		// The result is the value of the operand after it has been decremented.
		// If an int operand’s value is the smallest representable for that type, the operand is decremented as if it were float.

		// Spec: https://phplang.org/spec/10-expressions.html#prefix-increment-and-decrement-operators
		// For a prefix ++ or -- operator used with an operand having the value INF, -INF, or NAN, there is no side effect, and the result is the operand’s value.
		return calculateFloating(operand, "-", values.NewFloat(1))

	default:
		return values.NewInt(0), phpError.NewError("calculateIncDecFloating: Operator \"%s\" not implemented", operator)
	}
}

func calculateIncDecNull(operator string) (values.RuntimeValue, phpError.Error) {
	switch operator {
	case "++":
		// Spec: https://phplang.org/spec/10-expressions.html#prefix-increment-and-decrement-operators
		// For a prefix ++ operator used with a NULL-valued operand, the side effect is that the operand’s type is changed to int,
		// the operand’s value is set to zero, and that value is incremented by 1.
		// The result is the value of the operand after it has been incremented.
		return values.NewInt(1), nil

	case "--":
		// Spec: https://phplang.org/spec/10-expressions.html#prefix-increment-and-decrement-operators
		// For a prefix – operator used with a NULL-valued operand, there is no side effect, and the result is the operand’s value.
		return values.NewNull(), nil

	default:
		return values.NewInt(0), phpError.NewError("calculateIncDecNull: Operator \"%s\" not implemented", operator)
	}
}

func calculateIncDecString(operator string, operand *values.Str) (values.RuntimeValue, phpError.Error) {
	switch operator {
	case "++":
		// Spec: https://phplang.org/spec/10-expressions.html#prefix-increment-and-decrement-operators
		// For a prefix "++" operator used with an operand whose value is an empty string,
		// the side effect is that the operand’s value is changed to the string “1”. The type of the operand is unchanged.
		// The result is the new value of the operand.
		if operand.Value == "" {
			return values.NewStr("1"), nil
		}
		return values.NewVoid(), phpError.NewError("TODO calculateIncDecString")

	case "--":
		// Spec: https://phplang.org/spec/10-expressions.html#prefix-increment-and-decrement-operators
		// For a prefix "--" operator used with an operand whose value is an empty string,
		// the side effect is that the operand’s type is changed to int, the operand’s value is set to zero,
		// and that value is decremented by 1. The result is the value of the operand after it has been incremented.
		if operand.Value == "" {
			return values.NewInt(-1), nil
		}
		return values.NewVoid(), phpError.NewError("TODO calculateIncDecString")

	default:
		return values.NewInt(0), phpError.NewError("calculateIncDecNull: Operator \"%s\" not implemented", operator)
	}

	// TODO calculateIncDecString
	// Spec: https://phplang.org/spec/10-expressions.html#prefix-increment-and-decrement-operators
	/*
		String Operands

		For a prefix -- or ++ operator used with a numeric string, the numeric string is treated as the corresponding int or float value.

		For a prefix -- operator used with a non-numeric string-valued operand, there is no side effect, and the result is the operand’s value.

		For a non-numeric string-valued operand that contains only alphanumeric characters, for a prefix ++ operator, the operand is considered to be a representation of a base-36 number (i.e., with digits 0–9 followed by A–Z or a–z) in which letter case is ignored for value purposes. The right-most digit is incremented by 1. For the digits 0–8, that means going to 1–9. For the letters “A”–“Y” (or “a”–“y”), that means going to “B”–“Z” (or “b”–“z”). For the digit 9, the digit becomes 0, and the carry is added to the next left-most digit, and so on. For the digit “Z” (or “z”), the resulting string has an extra digit “A” (or “a”) appended. For example, when incrementing, “a” -> “b”, “Z” -> “AA”, “AA” -> “AB”, “F29” -> “F30”, “FZ9” -> “GA0”, and “ZZ9” -> “AAA0”. A digit position containing a number wraps modulo-10, while a digit position containing a letter wraps modulo-26.

		For a non-numeric string-valued operand that contains any non-alphanumeric characters, for a prefix ++ operator, all characters up to and including the right-most non-alphanumeric character is passed through to the resulting string, unchanged. Characters to the right of that right-most non-alphanumeric character are treated like a non-numeric string-valued operand that contains only alphanumeric characters, except that the resulting string will not be extended. Instead, a digit position containing a number wraps modulo-10, while a digit position containing a letter wraps modulo-26.
	*/
}

// -------------------------------------- unary-op-calculation -------------------------------------- MARK: unary-op-calculation

func calculateUnary(operator string, operand values.RuntimeValue) (values.RuntimeValue, phpError.Error) {
	switch operand.GetType() {
	case values.BoolValue:
		return calculateUnaryBoolean(operator, operand.(*values.Bool))
	case values.IntValue:
		return calculateUnaryInteger(operator, operand.(*values.Int))
	case values.FloatValue:
		return calculateUnaryFloating(operator, operand.(*values.Float))
	case values.NullValue:
		// Spec: https://phplang.org/spec/10-expressions.html#unary-arithmetic-operators
		// For a unary + or unary - operator used with a NULL-valued operand, the value of the result is zero and the type is int.
		return values.NewInt(0), nil
	default:
		return values.NewVoid(), phpError.NewError("calculateUnary: Type \"%s\" not implemented", operand.GetType())
	}

	// TODO calculateUnary - string
	// Spec: https://phplang.org/spec/10-expressions.html#unary-arithmetic-operators
	// For a unary + or - operator used with a numeric string or a leading-numeric string, the string is first converted to an int or float, as appropriate, after which it is handled as an arithmetic operand. The trailing non-numeric characters in leading-numeric strings are ignored. With a non-numeric string, the result has type int and value 0. If the string was leading-numeric or non-numeric, a non-fatal error MUST be produced.
	// For a unary ~ operator used with a string, the result is the string with each byte being bitwise complement of the corresponding byte of the source string.

	// TODO calculateUnary - object
	// If the operand has an object type supporting the operation, then the object semantics defines the result. Otherwise, for ~ the fatal error is issued and for + and - the object is converted to int.
}

func calculateUnaryBoolean(operator string, operand *values.Bool) (*values.Int, phpError.Error) {
	switch operator {
	case "+":
		// Spec: https://phplang.org/spec/10-expressions.html#unary-arithmetic-operators
		// For a unary "+" operator used with a TRUE-valued operand, the value of the result is 1 and the type is int.
		// When used with a FALSE-valued operand, the value of the result is zero and the type is int.
		if operand.Value {
			return values.NewInt(1), nil
		}
		return values.NewInt(0), nil

	case "-":
		// Spec: https://phplang.org/spec/10-expressions.html#unary-arithmetic-operators
		// For a unary "-" operator used with a TRUE-valued operand, the value of the result is -1 and the type is int.
		// When used with a FALSE-valued operand, the value of the result is zero and the type is int.
		if operand.Value {
			return values.NewInt(-1), nil
		}
		return values.NewInt(0), nil

	default:
		return values.NewInt(0), phpError.NewError("calculateUnaryBoolean: Operator \"%s\" not implemented", operator)
	}
}

func calculateUnaryFloating(operator string, operand *values.Float) (values.RuntimeValue, phpError.Error) {
	switch operator {
	case "+":
		// Spec: https://phplang.org/spec/10-expressions.html#unary-arithmetic-operators
		// For a unary "+" operator used with an arithmetic operand, the type and value of the result is the type and value of the operand.
		return operand, nil

	case "-":
		// Spec: https://phplang.org/spec/10-expressions.html#unary-arithmetic-operators
		// For a unary - operator used with an arithmetic operand, the value of the result is the negated value of the operand.
		// However, if an int operand’s original value is the smallest representable for that type,
		// the operand is treated as if it were float and the result will be float.
		return values.NewFloat(-operand.Value), nil

	case "~":
		// Spec: https://phplang.org/spec/10-expressions.html#unary-arithmetic-operators
		// For a unary ~ operator used with a float operand, the value of the operand is first converted to int before the bitwise complement is computed.
		intRuntimeValue, err := variableHandling.ToValueType(values.IntValue, operand, false)
		if err != nil {
			return values.NewFloat(0), err
		}
		return calculateUnaryInteger(operator, intRuntimeValue.(*values.Int))

	default:
		return values.NewFloat(0), phpError.NewError("calculateUnaryFloating: Operator \"%s\" not implemented", operator)
	}
}

func calculateUnaryInteger(operator string, operand *values.Int) (*values.Int, phpError.Error) {
	switch operator {
	case "+":
		// Spec: https://phplang.org/spec/10-expressions.html#unary-arithmetic-operators
		// For a unary "+" operator used with an arithmetic operand, the type and value of the result is the type and value of the operand.
		return operand, nil

	case "-":
		// Spec: https://phplang.org/spec/10-expressions.html#unary-arithmetic-operators
		// For a unary - operator used with an arithmetic operand, the value of the result is the negated value of the operand.
		// However, if an int operand’s original value is the smallest representable for that type,
		// the operand is treated as if it were float and the result will be float.
		return values.NewInt(-operand.Value), nil

	case "~":
		// Spec: https://phplang.org/spec/10-expressions.html#unary-arithmetic-operators
		// For a unary ~ operator used with an int operand, the type of the result is int.
		// The value of the result is the bitwise complement of the value of the operand
		// (that is, each bit in the result is set if and only if the corresponding bit in the operand is clear).
		return values.NewInt(^operand.Value), nil
	default:
		return values.NewInt(0), phpError.NewError("calculateUnaryInteger: Operator \"%s\" not implemented", operator)
	}
}

// -------------------------------------- binary-op-calculation -------------------------------------- MARK: binary-op-calculation

func calculate(operand1 values.RuntimeValue, operator string, operand2 values.RuntimeValue) (values.RuntimeValue, phpError.Error) {
	resultType := values.VoidValue
	if slices.Contains([]string{"."}, operator) {
		resultType = values.StrValue
	} else if slices.Contains([]string{"&", "|", "^", "<<", ">>"}, operator) {
		resultType = values.IntValue
	} else {
		resultType = values.IntValue
		if operand1.GetType() == values.FloatValue || operand2.GetType() == values.FloatValue {
			resultType = values.FloatValue
		}
	}

	var err phpError.Error
	operand1, err = variableHandling.ToValueType(resultType, operand1, false)
	if err != nil {
		return values.NewVoid(), err
	}
	operand2, err = variableHandling.ToValueType(resultType, operand2, false)
	if err != nil {
		return values.NewVoid(), err
	}
	// TODO testing how PHP behavious: var_dump(1.0 + 2); var_dump(1 + 2.0); var_dump("1" + 2);
	// var_dump("1" + "2"); => int
	// var_dump("1" . 2); => str
	// type order "string" - "int" - "float"

	// Testen
	//   true + 2
	//   true && 3

	switch resultType {
	case values.IntValue:
		return calculateInteger(operand1.(*values.Int), operator, operand2.(*values.Int))
	case values.FloatValue:
		return calculateFloating(operand1.(*values.Float), operator, operand2.(*values.Float))
	case values.StrValue:
		return calculateString(operand1.(*values.Str), operator, operand2.(*values.Str))
	default:
		return values.NewVoid(), phpError.NewError("calculate: Type \"%s\" not implemented", resultType)
	}
}

func calculateFloating(operand1 *values.Float, operator string, operand2 *values.Float) (values.RuntimeValue, phpError.Error) {
	switch operator {
	case "+":
		return values.NewFloat(operand1.Value + operand2.Value), nil
	case "-":
		return values.NewFloat(operand1.Value - operand2.Value), nil
	case "*":
		return values.NewFloat(operand1.Value * operand2.Value), nil
	case "/":
		return values.NewFloat(operand1.Value / operand2.Value), nil
	case "**":
		return values.NewFloat(math.Pow(operand1.Value, operand2.Value)), nil
	case "%":
		op1, err := variableHandling.IntVal(operand1, false)
		if err != nil {
			return values.NewFloat(0), err
		}
		op2, err := variableHandling.IntVal(operand2, false)
		if err != nil {
			return values.NewFloat(0), err
		}
		return values.NewInt(op1 % op2), nil
	default:
		return values.NewFloat(0), phpError.NewError("calculateFloating: Operator \"%s\" not implemented", operator)
	}
}

func calculateInteger(operand1 *values.Int, operator string, operand2 *values.Int) (*values.Int, phpError.Error) {
	switch operator {
	case "<<":
		return values.NewInt(operand1.Value << operand2.Value), nil
	case ">>":
		return values.NewInt(operand1.Value >> operand2.Value), nil
	case "^":
		return values.NewInt(operand1.Value ^ operand2.Value), nil
	case "|":
		return values.NewInt(operand1.Value | operand2.Value), nil
	case "&":
		return values.NewInt(operand1.Value & operand2.Value), nil
	case "+":
		return values.NewInt(operand1.Value + operand2.Value), nil
	case "-":
		return values.NewInt(operand1.Value - operand2.Value), nil
	case "*":
		return values.NewInt(operand1.Value * operand2.Value), nil
	case "/":
		if operand2.Value == 0 {
			// TODO Add position in output: Fatal error: Uncaught DivisionByZeroError: Division by zero in /home/user/scripts/code.php:3
			return values.NewInt(0), phpError.NewError("Uncaught DivisionByZeroError: Division by zero")
		}
		return values.NewInt(operand1.Value / operand2.Value), nil
	case "%":
		return values.NewInt(operand1.Value % operand2.Value), nil
	case "**":
		return values.NewInt(int64(math.Pow(float64(operand1.Value), float64(operand2.Value)))), nil
	default:
		return values.NewInt(0), phpError.NewError("calculateInteger: Operator \"%s\" not implemented", operator)
	}
}

func calculateString(operand1 *values.Str, operator string, operand2 *values.Str) (*values.Str, phpError.Error) {
	switch operator {
	case ".":
		return values.NewStr(operand1.Value + operand2.Value), nil
	default:
		return values.NewStr(""), phpError.NewError("calculateString: Operator \"%s\" not implemented", operator)
	}
}

// -------------------------------------- class-object -------------------------------------- MARK: class-object

func (interpreter *Interpreter) CallMethod(object *values.Object, method string, args []ast.IExpression, env *Environment) (values.RuntimeValue, phpError.Error) {
	methodDefinition, found := object.GetMethod(method)
	if !found {
		return values.NewNull(), phpError.NewError("Class %s does not have a function \"%s\"", object.Class.Name, method)
	}

	methodEnv, err := NewEnvironment(env, nil, interpreter)
	if err != nil {
		return values.NewVoid(), err
	}
	methodEnv.CurrentObject = object
	methodEnv.CurrentMethod = methodDefinition

	if len(methodDefinition.Params) != len(args) {
		return values.NewVoid(), phpError.NewError(
			"Uncaught ArgumentCountError: %s::%s() expects exactly %d arguments, %d given",
			object.Class.BaseClass, methodDefinition.Name, len(methodDefinition.Params), len(args),
		)
	}

	for index, param := range methodDefinition.Params {
		runtimeValue := must(interpreter.processStmt(args[index], env))

		// Check if the parameter types match
		err = checkParameterTypes(runtimeValue, param.Type)
		if err != nil && err.GetMessage() == "Types do not match" {
			givenType, err := variableHandling.GetType(runtimeValue)
			if err != nil {
				return values.NewVoid(), err
			}
			return values.NewVoid(), phpError.NewError(
				"Uncaught TypeError: %s::%s(): Argument #%d (%s) must be of type %s, %s given",
				object.Class.BaseClass, methodDefinition.Name, index+1, param.Name, strings.Join(param.Type, "|"), givenType,
			)
		}
		// Declare parameter in method environment
		methodEnv.declareVariable(param.Name, values.DeepCopy(runtimeValue))
	}

	runtimeValue, err := interpreter.processStmt(methodDefinition.Body, methodEnv)
	if err != nil && !(err.GetErrorType() == phpError.EventError && err.GetMessage() == phpError.ReturnEvent) {
		return runtimeValue, err
	}
	err = checkParameterTypes(runtimeValue, methodDefinition.ReturnType)
	if err != nil && err.GetMessage() == "Types do not match" {
		givenType, err := variableHandling.GetType(runtimeValue)
		if runtimeValue.GetType() == values.VoidValue {
			givenType = "void"
		}
		if err != nil {
			return runtimeValue, err
		}
		return runtimeValue, phpError.NewError(
			"Uncaught TypeError: %s::%s(): Return value must be of type %s, %s given",
			object.Class.BaseClass, methodDefinition.Name, strings.Join(methodDefinition.ReturnType, "|"), givenType,
		)
	}

	return runtimeValue, nil
}

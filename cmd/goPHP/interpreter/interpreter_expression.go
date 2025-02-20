package interpreter

import (
	"GoPHP/cmd/goPHP/ast"
	"GoPHP/cmd/goPHP/common"
	"GoPHP/cmd/goPHP/phpError"
	"strings"
)

// ProcessTextExpr implements Visitor.
func (interpreter *Interpreter) ProcessTextExpr(expr *ast.TextExpression, _ any) (any, error) {
	interpreter.print(expr.Value)
	return NewVoidRuntimeValue(), nil
}

// ProcessExpr implements Visitor.
func (interpreter *Interpreter) ProcessExpr(stmt *ast.Expression, _ any) (any, error) {
	panic("ProcessExpr should never be called")
}

// ProcessVariableNameExpr implements Visitor.
func (interpreter *Interpreter) ProcessVariableNameExpr(expr *ast.VariableNameExpression, _ any) (any, error) {
	panic("ProcessVariableNameExpr should never be called")
}

// ProcessArrayNextKeyExpr implements Visitor.
func (visitor *Interpreter) ProcessArrayNextKeyExpr(stmt *ast.ArrayNextKeyExpression, _ any) (any, error) {
	panic("ProcessArrayNextKeyExpr should never be called")
}

// ProcessParenthesizedExpr implements Visitor.
func (interpreter *Interpreter) ProcessParenthesizedExpr(stmt *ast.ParenthesizedExpression, env any) (any, error) {
	return interpreter.processStmt(stmt.Expr, env)
}

// ProcessArrayLiteralExpr implements Visitor.
func (interpreter *Interpreter) ProcessArrayLiteralExpr(expr *ast.ArrayLiteralExpression, env any) (any, error) {
	return interpreter.exprToRuntimeValue(expr, env.(*Environment))
}

// ProcessFloatingLiteralExpr implements Visitor.
func (interpreter *Interpreter) ProcessFloatingLiteralExpr(expr *ast.FloatingLiteralExpression, env any) (any, error) {
	return interpreter.exprToRuntimeValue(expr, env.(*Environment))
}

// ProcessIntegerLiteralExpr implements Visitor.
func (interpreter *Interpreter) ProcessIntegerLiteralExpr(expr *ast.IntegerLiteralExpression, env any) (any, error) {
	return interpreter.exprToRuntimeValue(expr, env.(*Environment))
}

// ProcessStringLiteralExpr implements Visitor.
func (interpreter *Interpreter) ProcessStringLiteralExpr(expr *ast.StringLiteralExpression, env any) (any, error) {
	return interpreter.exprToRuntimeValue(expr, env.(*Environment))
}

// ProcessSimpleVariableExpr implements Visitor.
func (interpreter *Interpreter) ProcessSimpleVariableExpr(expr *ast.SimpleVariableExpression, env any) (any, error) {
	return interpreter.lookupVariable(expr, env.(*Environment), false)
}

// ProcessSimpleAssignmentExpr implements Visitor.
func (interpreter *Interpreter) ProcessSimpleAssignmentExpr(expr *ast.SimpleAssignmentExpression, env any) (any, error) {
	if !ast.IsVariableExpr(expr.Variable) {
		return NewVoidRuntimeValue(),
			phpError.NewError("processSimpleAssignmentExpr: Invalid variable: %s", expr.Variable)
	}

	variableName := mustOrVoid(interpreter.varExprToVarName(expr.Variable, env.(*Environment)))
	currentValue, _ := env.(*Environment).lookupVariable(variableName)

	if currentValue.GetType() == StringValue && expr.Variable.GetKind() == ast.SubscriptExpr {
		if expr.Variable.(*ast.SubscriptExpression).Index == nil {
			return NewVoidRuntimeValue(), phpError.NewError("[] operator not supported for strings in %s", expr.Variable.GetPosition().ToPosString())
		}
		if expr.Variable.(*ast.SubscriptExpression).Index.GetKind() != ast.IntegerLiteralExpr {
			indexType, err := literalExprTypeToRuntimeValue(expr.Variable.(*ast.SubscriptExpression).Index)
			if err != nil {
				return NewVoidRuntimeValue(), err
			}
			return NewVoidRuntimeValue(), phpError.NewError("Cannot access offset of type %s on string in %s", indexType, expr.Variable.(*ast.SubscriptExpression).Index.GetPosition().ToPosString())
		}

		key := expr.Variable.(*ast.SubscriptExpression).Index.(*ast.IntegerLiteralExpression).Value
		value := must(interpreter.processStmt(expr.Value, env))

		currentValue, _ = env.(*Environment).lookupVariable(variableName)
		str := currentValue.(*StringRuntimeValue).Value

		valueStr, err := lib_strval(value)
		if err != nil {
			return NewVoidRuntimeValue(), err
		}
		if valueStr == "" {
			return NewVoidRuntimeValue(),
				phpError.NewError("Cannot assign an empty string to a string offset in %s", expr.Value.GetPosition().ToPosString())
		}

		str = common.ExtendWithSpaces(str, int(key+1))
		str = common.ReplaceAtPos(str, valueStr, int(key))

		_, err = env.(*Environment).declareVariable(variableName, NewStringRuntimeValue(str))
		return value, err
	}

	if currentValue.GetType() == NullValue && expr.Variable.GetKind() == ast.SubscriptExpr {
		env.(*Environment).declareVariable(variableName, NewArrayRuntimeValue())
		currentValue, _ = env.(*Environment).lookupVariable(variableName)
	}

	if currentValue.GetType() == ArrayValue {
		if expr.Variable.GetKind() != ast.SubscriptExpr {
			return NewVoidRuntimeValue(), phpError.NewError("processSimpleAssignmentExpr: Unsupported variable type %s", expr.Variable.GetKind())
		}

		keys := []ast.IExpression{expr.Variable.(*ast.SubscriptExpression).Index}
		subarray := expr.Variable.(*ast.SubscriptExpression).Variable
		for subarray.GetKind() == ast.SubscriptExpr {
			keys = append(keys, subarray.(*ast.SubscriptExpression).Index)
			subarray = subarray.(*ast.SubscriptExpression).Variable
		}

		var value IRuntimeValue
		for i := len(keys) - 1; i >= 0; i-- {
			if currentValue.GetType() != ArrayValue {
				return NewVoidRuntimeValue(), phpError.NewError("processSimpleAssignmentExpr: Unexpected currentValue type %s", currentValue.GetType())
			}

			array := currentValue.(*ArrayRuntimeValue)
			var keyValue IRuntimeValue = nil
			if keys[i] != nil {
				keyValue = must(interpreter.processStmt(keys[i], env))
			}

			if i == 0 {
				value = must(interpreter.processStmt(expr.Value, env))
				if err := array.SetElement(keyValue, deepCopy(value)); err != nil {
					return NewVoidRuntimeValue(), err
				}
				break
			}

			_, found := array.findKey(keyValue)
			if found {
				currentValue, _ = array.GetElement(keyValue)
			} else {
				newArray := NewArrayRuntimeValue()
				if err := array.SetElement(keyValue, newArray); err != nil {
					return NewVoidRuntimeValue(), err
				}
				currentValue = newArray
			}
		}

		return value, nil
	}

	value := must(interpreter.processStmt(expr.Value, env))
	return env.(*Environment).declareVariable(variableName, value)
}

// ProcessSubscriptExpr implements Visitor.
func (interpreter *Interpreter) ProcessSubscriptExpr(expr *ast.SubscriptExpression, env any) (any, error) {
	// Spec: https://phplang.org/spec/10-expressions.html#grammar-subscript-expression

	variable := must(interpreter.lookupVariable(expr.Variable, env.(*Environment), false))

	if variable.GetType() == StringValue {
		if expr.Index == nil {
			return NewVoidRuntimeValue(), phpError.NewError("Cannot use [] for reading in %s", expr.Variable.GetPosition().ToPosString())
		}
		if expr.Index.GetKind() != ast.IntegerLiteralExpr {
			indexType, err := literalExprTypeToRuntimeValue(expr.Index)
			if err != nil {
				return NewVoidRuntimeValue(), err
			}
			return NewVoidRuntimeValue(), phpError.NewError("Cannot access offset of type %s on string in %s", indexType, expr.Index.GetPosition().ToPosString())
		}

		key := expr.Index.(*ast.IntegerLiteralExpression).Value
		str := variable.(*StringRuntimeValue).Value

		if len(str) <= int(key) {
			return NewVoidRuntimeValue(), phpError.NewError("Uninitialized string offset %d in %s", key, expr.Index.GetPosition().ToPosString())
		}

		return NewStringRuntimeValue(str[key : key+1]), nil
		// str = common.ExtendWithSpaces(str, int(key+1))
		// str = common.ReplaceAtPos(str, valueStr, int(key))

		// _, err = env.(*Environment).declareVariable(variableName, NewStringRuntimeValue(str))
		// return value, err
	}

	// Spec: https://phplang.org/spec/10-expressions.html#grammar-subscript-expression
	// dereferencable-expression designates an array
	if variable.GetType() == ArrayValue {
		array := variable.(*ArrayRuntimeValue)

		keys := []ast.IExpression{expr.Index}
		subarray := expr.Variable
		for subarray.GetKind() == ast.SubscriptExpr {
			keys = append(keys, subarray.(*ast.SubscriptExpression).Index)
			subarray = subarray.(*ast.SubscriptExpression).Variable
		}

		for i := len(keys) - 1; i >= 0; i-- {
			// TODO processSubscriptExpr - no key
			// Spec: https://phplang.org/spec/10-expressions.html#grammar-subscript-expression
			// If expression is omitted, a new element is inserted. Its key has type int and is one more than the highest, previously assigned int key for this array. If this is the first element with an int key, key 0 is used. If the largest previously assigned int key is the largest integer value that can be represented, the new element is not added. The result is the added new element, or NULL if the element was not added.

			keyValue := must(interpreter.processStmt(keys[i], env))
			_, exists := array.findKey(keyValue)

			if i == 0 {
				// Spec: https://phplang.org/spec/10-expressions.html#grammar-subscript-expression
				// If expression is present, if the designated element exists,
				// the type and value of the result is the type and value of that element;
				// otherwise, the result is NULL.
				if exists {
					element, _ := array.GetElement(keyValue)
					return element, nil
				} else {
					return NewNullRuntimeValue(), nil
				}
			}

			if exists {
				element, _ := array.GetElement(keyValue)
				if element.GetType() != ArrayValue {
					return NewNullRuntimeValue(), phpError.NewError("ProcessSubscriptExpr: Expected type Array. Got: %s", element.GetType())
				}
				array = element.(*ArrayRuntimeValue)
				continue
			}
			return NewNullRuntimeValue(), phpError.NewError("ProcessSubscriptExpr: Array does not contain key: %s", keyValue.GetType())

			// TODO processSubscriptExpr
			// If the usage context is as the left-hand side of a simple-assignment-expression, the value of the new element is the value of the right-hand side of that simple-assignment-expression.
			// If the usage context is as the left-hand side of a compound-assignment-expression: the expression e1 op= e2 is evaluated as e1 = NULL op (e2).
			// If the usage context is as the operand of a postfix- or prefix-increment or decrement operator, the value of the new element is considered to be NULL.
		}
	}

	return NewVoidRuntimeValue(), phpError.NewError("Unsupported subscript expression: %s", ast.ToString(expr))

	/*
	   If dereferencable-expression designates a string, expression must not designate a string.

	   expression can be omitted only if subscript-expression is used in a modifiable-lvalue context and dereferencable-expression does not designate a string. Exception from this is when dereferencable-expression is an empty string - then it is converted to an empty array.

	   If subscript-expression is used in a non-lvalue context, the element being designated must exist.

	   Semantics

	   A subscript-expression designates a (possibly non-existent) element of an array or string. When subscript-expression designates an object of a type that implements ArrayAccess, the minimal semantics are defined below; however, they can be augmented by that object’s methods offsetGet and offsetSet.

	   The element key is designated by expression. If the type of element-key is neither int nor string, keys with float or bool values, or strings whose contents match exactly the pattern of decimal-literal, are converted to integer, and key values of all other types are converted to string.

	   If both dereferencable-expression and expression designate strings, expression is treated as if it specified the int key zero instead and a non-fatal error is produces.

	   A subscript-expression designates a modifiable lvalue if and only if dereferencable-expression designates a modifiable lvalue.

	   dereferencable-expression designates a string

	   The expression is converted to int and the result is the character of the string at the position corresponding to that integer. If the integer is negative, the position is counted backwards from the end of the string. If the position refers to a non-existing offset, the result is an empty string.

	   If the operator is used as the left-hand side of a simple-assignment-expression,

	       If the assigned string is empty, or in case of non-existing negative offset (absolute value larger than string length), a warning is raised and no assignment is performed.
	       If the offset is larger than the current string length, the string is extended to a length equal to the offset value, using space (0x20) padding characters.
	       The value being assigned is converted to string and the character in the specified offset is replaced by the first character of the string.

	   The subscript operator can not be used on a string value in a byRef context or as the operand of the postfix- or prefix-increment or decrement operators or on the left side of compound-assignment-expression, doing so will result in a fatal error.

	   dereferencable-expression designates an object of a type that implements ArrayAccess

	   If expression is present,

	       If subscript-expression is used in a non-lvalue context, the object’s method offsetGet is called with an argument of expression. The return value of the offsetGet is the result.
	       If the usage context is as the left-hand side of a simple-assignment-expression, the object’s method offsetSet is called with a first argument of expression and a second argument that is the value of the right-hand side of that simple-assignment-expression. The value of the right-hand side is the result.
	       If the usage context is as the left-hand side of a compound-assignment-expression, the expression e1[e] op= e2 is evaluated as e1[e] = e1->offsetGet(e) op (e2), which is then processed according to the rules for simple assignment immediately above.
	       If the usage context is as the operand of the postfix- or prefix-increment or decrement operators, the object’s method offsetGet is called with an argument of expression. However, this method has no way of knowing if an increment or decrement operator was used, or whether it was a prefix or postfix operator. In order for the value to be modified by the increment/decrement, offsetGet must return byRef. The result of the subscript operator value returned by offsetGet.

	   If expression is omitted,

	       If the usage context is as the left-hand side of a simple-assignment-expression, the object’s method offsetSet is called with a first argument of NULL and a second argument that is the value of the right-hand side of that simple-assignment-expression. The type and value of the result is the type and value of the right-hand side of that simple-assignment-expression.
	       If the usage context is as the left-hand side of a compound-assignment-expression: The expression e1[] op= e2 is evaluated as e1[] = e1->offsetGet(NULL) op (e2), which is then processed according to the rules for simple assignment immediately above.
	       If the usage context is as the operand of the postfix- or prefix-increment or decrement operators, the object’s method offsetGet is called with an argument of NULL. However, this method has no way of knowing if an increment or decrement operator was used, or whether it was a prefix or postfix operator. In order for the value to be modified by the increment/decrement, offsetGet must return byRef. The result of the subscript operator value returned by offsetGet.

	   Note: The brace ({...}) form of this operator has been deprecated.
	*/

	// Examples
	/*
	   $v = array(10, 20, 30);
	   $v[1] = 1.234;    // change the value (and type) of element [1]
	   $v[-10] = 19;   // insert a new element with int key -10
	   $v["red"] = TRUE; // insert a new element with string key "red"
	   [[2,4,6,8], [5,10], [100,200,300]][0][2]  // designates element with value 6
	   ["black", "white", "yellow"][1][2]  // designates substring "i" in "white"
	   function f() { return [1000, 2000, 3000]; }
	   f()[2];      // designates element with value 3000
	   "red"[1.9];    // designates "e"
	   "red"[-2];    // designates "e"
	   "red"[0][0][0];    // designates "r"
	   // -----------------------------------------
	   class MyVector implements ArrayAccess { /* ... */ /*}
	$vect1 = new MyVector(array(10, 'A' => 2.3, "up"));
	$vect1[10] = 987; // calls Vector::offsetSet(10, 987)
	$vect1[] = "xxx"; // calls Vector::offsetSet(NULL, "xxx")
	$x = $vect1[1];   // calls Vector::offsetGet(1)
	*/
}

// ProcessFunctionCallExpr implements Visitor.
func (interpreter *Interpreter) ProcessFunctionCallExpr(expr *ast.FunctionCallExpression, env any) (any, error) {
	functionNameRuntime := must(interpreter.processStmt(expr.FunctionName, env))
	functionName := mustOrVoid(lib_strval(functionNameRuntime))

	// Lookup native function
	nativeFunction, err := env.(*Environment).lookupNativeFunction(functionName)
	if err == nil {
		functionArguments := make([]IRuntimeValue, len(expr.Arguments))
		for index, arg := range expr.Arguments {
			runtimeValue := must(interpreter.processStmt(arg, env))
			functionArguments[index] = deepCopy(runtimeValue)
		}
		return nativeFunction(functionArguments, interpreter)
	}

	// Lookup user function
	userFunction := mustOrVoid(env.(*Environment).lookupUserFunction(functionName))

	functionEnv := NewEnvironment(env.(*Environment), nil, interpreter.ini)
	functionEnv.CurrentFunction = userFunction

	if len(userFunction.Params) != len(expr.Arguments) {
		return NewVoidRuntimeValue(), phpError.NewError(
			"Uncaught ArgumentCountError: %s() expects exactly %d arguments, %d given",
			userFunction.FunctionName, len(userFunction.Params), len(expr.Arguments),
		)
	}
	for index, param := range userFunction.Params {
		runtimeValue := must(interpreter.processStmt(expr.Arguments[index], env))

		// Check if the parameter types match
		err = checkParameterTypes(runtimeValue, param.Type)
		if err != nil && err.GetMessage() == "Types do not match" {
			givenType, err := lib_gettype(runtimeValue)
			if err != nil {
				return NewVoidRuntimeValue(), err
			}
			return NewVoidRuntimeValue(), phpError.NewError(
				"Uncaught TypeError: %s(): Argument #%d (%s) must be of type %s, %s given",
				userFunction.FunctionName, index+1, param.Name, strings.Join(param.Type, "|"), givenType,
			)
		}
		// Declare parameter in function environment
		functionEnv.declareVariable(param.Name, deepCopy(runtimeValue))
	}

	runtimeValue, err := interpreter.processStmt(userFunction.Body, functionEnv)
	if err != nil && !(err.GetErrorType() == phpError.EventError && err.GetMessage() == phpError.ReturnEvent) {
		return runtimeValue, err
	}
	err = checkParameterTypes(runtimeValue, userFunction.ReturnType)
	if err != nil && err.GetMessage() == "Types do not match" {
		givenType, err := lib_gettype(runtimeValue)
		if runtimeValue.GetType() == VoidValue {
			givenType = "void"
		}
		if err != nil {
			return runtimeValue, err
		}
		return runtimeValue, phpError.NewError(
			"Uncaught TypeError: %s(): Return value must be of type %s, %s given",
			userFunction.FunctionName, strings.Join(userFunction.ReturnType, "|"), givenType,
		)
	}
	return runtimeValue, nil

}

// ProcessEmptyIntrinsicExpr implements Visitor.
func (interpreter *Interpreter) ProcessEmptyIntrinsicExpr(expr *ast.EmptyIntrinsicExpression, env any) (any, error) {
	// Spec: https://phplang.org/spec/10-expressions.html#grammar-empty-intrinsic

	// This intrinsic returns TRUE if the variable or value designated by expression is empty,
	// where empty means that the variable designated by it does not exist, or it exists and its value compares equal to FALSE.
	// Otherwise, the intrinsic returns FALSE.

	// The following values are considered empty:
	// FALSE, 0, 0.0, "" (empty string), "0", NULL, an empty array, and any uninitialized variable.

	// If this intrinsic is used with an expression that designates a dynamic property,
	// then if the class of that property has an __isset, that method is called.
	// If that method returns TRUE, the value of the property is retrieved (which may call __get if defined)
	// and compared to FALSE as described above. Otherwise, the result is FALSE.

	var runtimeValue IRuntimeValue
	var err phpError.Error
	if ast.IsVariableExpr(expr.Arguments[0]) {
		runtimeValue, err = interpreter.lookupVariable(expr.Arguments[0], env.(*Environment), true)
		if err != nil {
			return NewBooleanRuntimeValue(true), nil
		}
	} else {
		runtimeValue = must(interpreter.processStmt(expr.Arguments[0], env))
	}

	boolean := mustOrVoid(lib_boolval(runtimeValue))
	return NewBooleanRuntimeValue(!boolean), nil
}

// ProcessEvalIntrinsicExpr implements Visitor.
func (interpreter *Interpreter) ProcessEvalIntrinsicExpr(expr *ast.EvalIntrinsicExpression, env any) (any, error) {
	// Spec: https://phplang.org/spec/10-expressions.html#grammar-eval-intrinsic

	expression, err := interpreter.processStmt(expr.Arguments[0], env)
	if err != nil {
		return NewVoidRuntimeValue(), phpError.NewParseError(err.Error())
	}
	expressionStr, err := lib_strval(expression)
	if err != nil {
		return NewVoidRuntimeValue(), err
	}

	_, err = interpreter.process("<?php "+expressionStr+" ?>", env.(*Environment), false)
	if err != nil {
		if err.GetErrorType() == phpError.EventError && err.GetMessage() == phpError.ReturnEvent {
			return interpreter.resultRuntimeValue, nil
		}
		return NewBooleanRuntimeValue(false), err
	}

	return NewNullRuntimeValue(), nil
}

// ProcessExitIntrinsicExpr implements Visitor.
func (interpreter *Interpreter) ProcessExitIntrinsicExpr(expr *ast.ExitIntrinsicExpression, env any) (any, error) {
	// Spec: https://phplang.org/spec/10-expressions.html#grammar-exit-intrinsic

	// "exit" and "die" are equivalent.

	// This intrinsic terminates the current script.
	// If expression designates a string, that string is written to STDOUT.
	// If expression designates an integer, that represents the script’s exit status code. Code 255 is reserved by PHP.
	// Code 0 represents “success”. The exit status code is made available to the execution environment.
	// If expression is omitted or is a string, the exit status code is zero. exit does not have a resulting value.

	// "exit" performs the following operations, in order:
	//   1. Writes the optional string to STDOUT.
	//   2. Calls any functions registered via the library function register_shutdown_function in their order of registration.
	//   3. Invokes destructors for all remaining instances.

	expression := expr.Arguments[0]
	if expression != nil {
		exprValue := must(interpreter.processStmt(expression, env))
		if exprValue.GetType() == StringValue {
			interpreter.print(exprValue.(*StringRuntimeValue).Value)
		}
		if exprValue.GetType() == IntegerValue {
			exitCode := exprValue.(*IntegerRuntimeValue).Value
			if exitCode >= 0 && exitCode < 255 {
				interpreter.exitCode = exitCode
			}
		}
	}

	// TODO processExitIntrinsicExpr - call shutdown functions
	// TODO processExitIntrinsicExpr - call destructors

	return NewVoidRuntimeValue(), phpError.NewEvent(phpError.ExitEvent)
}

// ProcessIssetIntrinsicExpr implements Visitor.
func (interpreter *Interpreter) ProcessIssetIntrinsicExpr(expr *ast.IssetIntrinsicExpression, env any) (any, error) {
	// Spec: https://phplang.org/spec/10-expressions.html#grammar-isset-intrinsic

	// This intrinsic returns TRUE if all the variables designated by variabless are set and their values are not NULL.
	// Otherwise, it returns FALSE.

	// If this intrinsic is used with an expression that designate a dynamic property,
	// then if the class of that property has an __isset, that method is called.
	// If that method returns TRUE, the value of the property is retrieved (which may call __get if defined)
	// and if it is not NULL, the result is TRUE. Otherwise, the result is FALSE.

	for _, arg := range expr.Arguments {
		runtimeValue, _ := interpreter.lookupVariable(arg, env.(*Environment), true)
		if runtimeValue.GetType() == NullValue {
			return NewBooleanRuntimeValue(false), nil
		}
	}
	return NewBooleanRuntimeValue(true), nil
}

// ProcessUnsetIntrinsicExpr implements Visitor.
func (interpreter *Interpreter) ProcessUnsetIntrinsicExpr(expr *ast.UnsetIntrinsicExpression, env any) (any, error) {
	// Spec: https://phplang.org/spec/11-statements.html#grammar-unset-statement

	// This statement unsets the variables designated by each variable in variable-list. No value is returned.
	// An attempt to unset a non-existent variable (such as a non-existent element in an array) is ignored.

	for _, arg := range expr.Arguments {
		variableName := mustOrVoid(interpreter.varExprToVarName(arg, env.(*Environment)))
		env.(*Environment).unsetVariable(variableName)
	}
	return NewVoidRuntimeValue(), nil
}

// ProcessConstantAccessExpr implements Visitor.
func (interpreter *Interpreter) ProcessConstantAccessExpr(expr *ast.ConstantAccessExpression, env any) (any, error) {
	// Magic constants

	// Spec: https://www.php.net/manual/en/language.constants.magic.php
	// The directory of the file. If used inside an include, the directory of the included file is returned.
	// This is equivalent to dirname(__FILE__). This directory name does not have a trailing slash unless it is the root directory.
	if expr.ConstantName == "__DIR__" {
		// TODO Use lib function dirname
		return NewStringRuntimeValue(common.ExtractPath(expr.GetPosition().Filename)), nil
	}

	// Spec: https://www.php.net/manual/en/language.constants.magic.php
	// The full path and filename of the file with symlinks resolved.
	// If used inside an include, the name of the included file is returned.
	if expr.ConstantName == "__FILE__" {
		return NewStringRuntimeValue(expr.GetPosition().Filename), nil
	}

	// Spec: https://www.php.net/manual/en/language.constants.magic.php
	// The function name, or {closure} for anonymous functions.
	if expr.ConstantName == "__FUNCTION__" {
		if env.(*Environment).CurrentFunction != nil {
			return NewStringRuntimeValue(env.(*Environment).CurrentFunction.FunctionName), nil
		}
		return NewStringRuntimeValue(""), nil
	}

	// Spec: https://www.php.net/manual/en/language.constants.magic.php
	// The current line number of the file.
	if expr.ConstantName == "__LINE__" {
		return NewIntegerRuntimeValue(int64(expr.GetPosition().Line)), nil
	}

	// TODO __CLASS__ 	The class name. The class name includes the namespace it was declared in (e.g. Foo\Bar). When used inside a trait method, __CLASS__ is the name of the class the trait is used in.
	// TODO __TRAIT__ 	The trait name. The trait name includes the namespace it was declared in (e.g. Foo\Bar).
	// TODO __METHOD__ 	The class method name.
	// TODO __PROPERTY__ 	Only valid inside a property hook. It is equal to the name of the property.
	// TODO __NAMESPACE__ 	The name of the current namespace.

	return env.(*Environment).lookupConstant(expr.ConstantName)
}

// ProcessCompoundAssignmentExpr implements Visitor.
func (interpreter *Interpreter) ProcessCompoundAssignmentExpr(expr *ast.CompoundAssignmentExpression, env any) (any, error) {
	if !ast.IsVariableExpr(expr.Variable) {
		return NewVoidRuntimeValue(),
			phpError.NewError("processCompoundAssignmentExpr: Invalid variable: %s", expr.Variable)
	}

	operand1 := must(interpreter.processStmt(expr.Variable, env))
	operand2 := must(interpreter.processStmt(expr.Value, env))
	newValue := must(calculate(operand1, expr.Operator, operand2))

	variableName := mustOrVoid(interpreter.varExprToVarName(expr.Variable, env.(*Environment)))

	return env.(*Environment).declareVariable(variableName, newValue)
}

// ProcessConditionalExpr implements Visitor.
func (interpreter *Interpreter) ProcessConditionalExpr(expr *ast.ConditionalExpression, env any) (any, error) {
	// Spec: https://phplang.org/spec/10-expressions.html#grammar-conditional-expression

	// Spec: https://phplang.org/spec/10-expressions.html#grammar-conditional-expression
	// Given the expression "e1 ? e2 : e3", e1 is evaluated first and converted to bool if it has another type.
	runtimeValue, isConditionTrue, err := interpreter.processCondition(expr.Cond, env.(*Environment))
	if err != nil {
		return runtimeValue, err
	}

	if isConditionTrue {
		if expr.IfExpr != nil {
			// Spec: https://phplang.org/spec/10-expressions.html#grammar-conditional-expression
			// If the result is TRUE, then and only then is e2 evaluated, and the result and its type become the result
			// and type of the whole expression.
			return interpreter.processStmt(expr.IfExpr, env)
		} else {
			// Spec: https://phplang.org/spec/10-expressions.html#grammar-conditional-expression
			// There is a sequence point after the evaluation of e1.
			// If e2 is omitted, the result and type of the whole expression is the value
			// and type of e1 (before the conversion to bool).
			return runtimeValue, nil
		}
	} else {
		// Spec: https://phplang.org/spec/10-expressions.html#grammar-conditional-expression
		// Otherwise, then and only then is e3 evaluated, and the result and its type become the result
		// and type of the whole expression.
		return interpreter.processStmt(expr.ElseExpr, env)
	}
}

// ProcessCoalesceExpr implements Visitor.
func (interpreter *Interpreter) ProcessCoalesceExpr(expr *ast.CoalesceExpression, env any) (any, error) {
	// Spec: https://phplang.org/spec/10-expressions.html#grammar-coalesce-expression

	// Store current error reporting
	errorReporting, _ := interpreter.ini.Get("error_reporting")
	// Suppress all errors
	interpreter.ini.Set("error_reporting", "0")

	cond, err := interpreter.processStmt(expr.Cond, env)

	// Restore previous error reporting
	interpreter.ini.Set("error_reporting", errorReporting)

	if err != nil {
		return cond, err
	}
	// Spec: https://phplang.org/spec/10-expressions.html#grammar-coalesce-expression
	// Given the expression e1 ?? e2, if e1 is set and not NULL (i.e. TRUE for isset), then the result is e1.

	if cond.GetType() != NullValue {
		return cond, nil
	}

	// Spec: https://phplang.org/spec/10-expressions.html#grammar-coalesce-expression
	// Otherwise, then and only then is e2 evaluated, and the result becomes the result of the whole expression.
	// There is a sequence point after the evaluation of e1.
	return interpreter.processStmt(expr.ElseExpr, env)

	// TODO processCoalesceExpr - handle uninitialized variables
	// Spec: https://phplang.org/spec/10-expressions.html#grammar-coalesce-expression
	// Note that the semantics of ?? is similar to isset so that uninitialized variables will not produce warnings when used in e1.
	// TODO use isset here - Steps: Add caching of expression results - map[exprId]IRuntimeValue
}

// ProcessRelationalExpr implements Visitor.
func (interpreter *Interpreter) ProcessRelationalExpr(expr *ast.RelationalExpression, env any) (any, error) {
	lhs := must(interpreter.processStmt(expr.Lhs, env))
	rhs := must(interpreter.processStmt(expr.Rhs, env))
	return compareRelation(lhs, expr.Operator, rhs)
}

// ProcessEqualityExpr implements Visitor.
func (interpreter *Interpreter) ProcessEqualityExpr(expr *ast.EqualityExpression, env any) (any, error) {
	lhs := must(interpreter.processStmt(expr.Lhs, env))
	rhs := must(interpreter.processStmt(expr.Rhs, env))
	return compare(lhs, expr.Operator, rhs)
}

// ProcessBinaryOpExpr implements Visitor.
func (interpreter *Interpreter) ProcessBinaryOpExpr(expr *ast.BinaryOpExpression, env any) (any, error) {
	lhs := must(interpreter.processStmt(expr.Lhs, env))
	rhs := must(interpreter.processStmt(expr.Rhs, env))
	return calculate(lhs, expr.Operator, rhs)
}

// ProcessUnaryExpr implements Visitor.
func (interpreter *Interpreter) ProcessUnaryExpr(expr *ast.UnaryOpExpression, env any) (any, error) {
	operand := must(interpreter.processStmt(expr.Expr, env))
	return calculateUnary(expr.Operator, operand)
}

// ProcessCastExpr implements Visitor.
func (interpreter *Interpreter) ProcessCastExpr(expr *ast.CastExpression, env any) (any, error) {
	// Spec: https://phplang.org/spec/10-expressions.html#grammar-cast-type

	// A cast-type of "unset" is no longer supported and results in a compile-time error.
	// With the exception of the cast-type unset and binary (see below), the value of the operand cast-expression is converted to the type specified by cast-type, and that is the type and value of the result. This construct is referred to as a cast and is used as the verb, “to cast”. If no conversion is involved, the type and value of the result are the same as those of cast-expression.

	// A cast can result in a loss of information.

	// TODO processCastExpr - object
	// A cast-type of "object" results in a conversion to type "object".

	value := must(interpreter.processStmt(expr.Expr, env))

	switch expr.Operator {
	case "array":
		// Spec: https://phplang.org/spec/10-expressions.html#grammar-cast-type
		// A cast-type of "array" results in a conversion to type array.
		return lib_arrayval(value)
	case "binary", "string":
		// Spec: https://phplang.org/spec/10-expressions.html#grammar-cast-type
		// A cast-type of "binary" is reserved for future use in dealing with so-called binary strings. For now, it is fully equivalent to "string" cast.
		// A cast-type of "string" results in a conversion to type "string".
		return runtimeValueToValueType(StringValue, value)
	case "bool", "boolean":
		// Spec: https://phplang.org/spec/10-expressions.html#grammar-cast-type
		// A cast-type of "bool" or "boolean" results in a conversion to type "bool".
		return runtimeValueToValueType(BooleanValue, value)
	case "double", "float", "real":
		// Spec: https://phplang.org/spec/10-expressions.html#grammar-cast-type
		// A cast-type of "float", "double", or "real" results in a conversion to type "float".
		return runtimeValueToValueType(FloatingValue, value)
	case "int", "integer":
		// Spec: https://phplang.org/spec/10-expressions.html#grammar-cast-type
		// A cast-type of "int" or "integer" results in a conversion to type "int".
		return runtimeValueToValueType(IntegerValue, value)
	default:
		return NewVoidRuntimeValue(), phpError.NewError("processCastExpr: Unsupported cast type %s", expr.Operator)
	}
}

// ProcessLogicalExpr implements Visitor.
func (interpreter *Interpreter) ProcessLogicalExpr(expr *ast.LogicalExpression, env any) (any, error) {
	// Evaluate LHS first
	lhs := must(interpreter.processStmt(expr.Lhs, env))
	// Convert LHS to boolean value
	left := mustOrVoid(lib_boolval(lhs))

	// Check if condition is already short circuited
	if expr.Operator == "||" {
		// if LHS of "or" is true, the result is already true
		if left {
			return NewBooleanRuntimeValue(true), nil
		}
	} else if expr.Operator == "&&" {
		// if LHS of "and" is false, the result is already false
		if !left {
			return NewBooleanRuntimeValue(false), nil
		}
	}

	// Evaluate RHS after checking if condition is already short circuited
	rhs := must(interpreter.processStmt(expr.Rhs, env))
	// Convert RHS to boolean value
	right := mustOrVoid(lib_boolval(rhs))

	if expr.Operator == "xor" {
		return NewBooleanRuntimeValue(left != right), nil
	}

	return NewBooleanRuntimeValue(right), nil
}

// ProcessLogicalNotExpr implements Visitor.
func (interpreter *Interpreter) ProcessLogicalNotExpr(expr *ast.LogicalNotExpression, env any) (any, error) {
	runtimeValue := must(interpreter.processStmt(expr.Expr, env))
	boolValue := mustOrVoid(lib_boolval(runtimeValue))
	return NewBooleanRuntimeValue(!boolValue), nil
}

// ProcessPostfixIncExpr implements Visitor.
func (interpreter *Interpreter) ProcessPostfixIncExpr(expr *ast.PostfixIncExpression, env any) (any, error) {
	// Spec: https://phplang.org/spec/10-expressions.html#postfix-increment-and-decrement-operators
	// These operators behave like their prefix counterparts except that the value of a postfix ++ or – expression is the value
	// before any increment or decrement takes place.

	previous := must(interpreter.processStmt(expr.Expr, env))
	newValue := must(calculateIncDec(expr.Operator, previous))

	variableName := mustOrVoid(interpreter.varExprToVarName(expr.Expr, env.(*Environment)))
	env.(*Environment).declareVariable(variableName, newValue)

	return previous, nil
}

// ProcessPrefixIncExpr implements Visitor.
func (interpreter *Interpreter) ProcessPrefixIncExpr(expr *ast.PrefixIncExpression, env any) (any, error) {
	previous := must(interpreter.processStmt(expr.Expr, env))
	newValue := must(calculateIncDec(expr.Operator, previous))

	variableName := mustOrVoid(interpreter.varExprToVarName(expr.Expr, env.(*Environment)))
	env.(*Environment).declareVariable(variableName, newValue)

	return newValue, nil
}

// ProcessPrintExpr implements Visitor.
func (interpreter *Interpreter) ProcessPrintExpr(expr *ast.PrintExpression, env any) (any, error) {
	// Spec: https://phplang.org/spec/10-expressions.html#grammar-print-expression
	// After converting print-expression’s value into a string, if necessary, print writes the resulting string to STDOUT.
	// Unlike echo, print can be used in any context allowing an expression. It always returns the value 1.

	runtimeValue := must(interpreter.processStmt(expr.Expr, env))

	str, err := lib_strval(runtimeValue)
	if err == nil {
		interpreter.print(str)
	}
	return NewIntegerRuntimeValue(1), err
}

// ProcessRequireExpr implements Visitor.
func (interpreter *Interpreter) ProcessRequireExpr(expr *ast.RequireExpression, env any) (any, error) {
	return interpreter.includeFile(expr.Expr, env.(*Environment), false, false)
}

// ProcessRequireOnceExpr implements Visitor.
func (interpreter *Interpreter) ProcessRequireOnceExpr(expr *ast.RequireOnceExpression, env any) (any, error) {
	return interpreter.includeFile(expr.Expr, env.(*Environment), false, true)
}

// ProcessIncludeExpr implements Visitor.
func (interpreter *Interpreter) ProcessIncludeExpr(expr *ast.IncludeExpression, env any) (any, error) {
	return interpreter.includeFile(expr.Expr, env.(*Environment), true, false)
}

// ProcessIncludeOnceExpr implements Visitor.
func (interpreter *Interpreter) ProcessIncludeOnceExpr(expr *ast.IncludeOnceExpression, env any) (any, error) {
	return interpreter.includeFile(expr.Expr, env.(*Environment), true, true)
}

package parser

import (
	"GoPHP/cmd/goPHP/ast"
	"testing"
)

func TestText(t *testing.T) {
	program, err := NewParser().ProduceAST("<html>")
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected := ast.NewTextExpression("<html>")
	actual := ast.ExprToTextExpr(ast.StmtToExprStatement(program.GetStatements()[0]).GetExpression())
	if expected.String() != actual.String() || expected.GetValue() != actual.GetValue() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected, actual)
	}
}

func TestVariableName(t *testing.T) {
	program, err := NewParser().ProduceAST("<?php $myVar;")
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected := ast.NewSimpleVariableExpression(ast.NewVariableNameExpression("$myVar"))
	actual := ast.ExprToSimpleVarExpr(ast.StmtToExprStatement(program.GetStatements()[0]).GetExpression())
	if expected.String() != actual.String() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected, actual)
	}

	program, err = NewParser().ProduceAST("<?php $$myVar;")
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected = ast.NewSimpleVariableExpression(ast.NewSimpleVariableExpression(ast.NewVariableNameExpression("$myVar")))
	actual = ast.ExprToSimpleVarExpr(ast.StmtToExprStatement(program.GetStatements()[0]).GetExpression())
	if expected.String() != actual.String() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected, actual)
	}
}

func TestArrayMember(t *testing.T) {
	program, err := NewParser().ProduceAST("<?php $myVar[];")
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected := ast.NewSubscriptExpression(ast.NewSimpleVariableExpression(ast.NewVariableNameExpression("$myVar")), nil)
	actual := ast.ExprToSubscriptExpr(ast.StmtToExprStatement(program.GetStatements()[0]).GetExpression())
	if expected.String() != actual.String() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected, actual)
	}

	program, err = NewParser().ProduceAST("<?php $myVar[0];")
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected = ast.NewSubscriptExpression(
		ast.NewSimpleVariableExpression(ast.NewVariableNameExpression("$myVar")),
		ast.NewIntegerLiteralExpression(0),
	)
	actual = ast.ExprToSubscriptExpr(ast.StmtToExprStatement(program.GetStatements()[0]).GetExpression())
	if expected.String() != actual.String() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected, actual)
	}
}

func TestFunctionCall(t *testing.T) {
	// Without argument
	program, err := NewParser().ProduceAST("<?php func();")
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected := ast.NewFunctionCallExpression("func", []ast.IExpression{})
	actual := ast.ExprToFuncCallExpr(ast.StmtToExprStatement(program.GetStatements()[0]).GetExpression())
	if expected.String() != actual.String() || expected.GetFunctionName() != actual.GetFunctionName() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected, actual)
	}

	// With argument
	program, err = NewParser().ProduceAST("<?php func(42);")
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected = ast.NewFunctionCallExpression("func", []ast.IExpression{ast.NewIntegerLiteralExpression(42)})
	actual = ast.ExprToFuncCallExpr(ast.StmtToExprStatement(program.GetStatements()[0]).GetExpression())
	if expected.String() != actual.String() || expected.GetFunctionName() != actual.GetFunctionName() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected, actual)
	}
}

func TestArrayLiteral(t *testing.T) {
	program, err := NewParser().ProduceAST(`<?php [1, "a", false];`)
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected := ast.NewArrayLiteralExpression()
	expected.AddElement(ast.NewIntegerLiteralExpression(0), ast.NewIntegerLiteralExpression(1))
	expected.AddElement(ast.NewIntegerLiteralExpression(1), ast.NewStringLiteralExpression("a", ast.DoubleQuotedString))
	expected.AddElement(ast.NewIntegerLiteralExpression(2), ast.NewBooleanLiteralExpression(false))
	actual := ast.ExprToArrayLitExpr(ast.StmtToExprStatement(program.GetStatements()[0]).GetExpression())
	if expected.String() != actual.String() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected, actual)
	}

	program, err = NewParser().ProduceAST(`<?php array(1, "a", false,);`)
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected = ast.NewArrayLiteralExpression()
	expected.AddElement(ast.NewIntegerLiteralExpression(0), ast.NewIntegerLiteralExpression(1))
	expected.AddElement(ast.NewIntegerLiteralExpression(1), ast.NewStringLiteralExpression("a", ast.DoubleQuotedString))
	expected.AddElement(ast.NewIntegerLiteralExpression(2), ast.NewBooleanLiteralExpression(false))
	actual = ast.ExprToArrayLitExpr(ast.StmtToExprStatement(program.GetStatements()[0]).GetExpression())
	if expected.String() != actual.String() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected, actual)
	}
}

func TestBooleanLiteral(t *testing.T) {
	program, err := NewParser().ProduceAST("<?php true;")
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected := ast.NewBooleanLiteralExpression(true)
	actual := ast.ExprToConstAccessExpr(ast.StmtToExprStatement(program.GetStatements()[0]).GetExpression())
	if expected.String() != actual.String() || expected.GetConstantName() != actual.GetConstantName() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected, actual)
	}

	program, err = NewParser().ProduceAST("<?php false;")
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected = ast.NewBooleanLiteralExpression(false)
	actual = ast.ExprToConstAccessExpr(ast.StmtToExprStatement(program.GetStatements()[0]).GetExpression())
	if expected.String() != actual.String() || expected.GetConstantName() != actual.GetConstantName() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected, actual)
	}
}

func TestNullLiteral(t *testing.T) {
	program, err := NewParser().ProduceAST("<?php null;")
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected := ast.NewNullLiteralExpression()
	actual := ast.ExprToConstAccessExpr(ast.StmtToExprStatement(program.GetStatements()[0]).GetExpression())
	if expected.String() != actual.String() || expected.GetConstantName() != actual.GetConstantName() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected, actual)
	}
}

func TestIntegerLiteral(t *testing.T) {
	// decimal-literal
	program, err := NewParser().ProduceAST("<?php 42;")
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected := ast.NewIntegerLiteralExpression(42)
	actual := ast.ExprToIntLitExpr(ast.StmtToExprStatement(program.GetStatements()[0]).GetExpression())
	if expected.String() != actual.String() || expected.GetValue() != actual.GetValue() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected, actual)
	}

	// octal-literal
	program, err = NewParser().ProduceAST("<?php 042;")
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected = ast.NewIntegerLiteralExpression(34)
	actual = ast.ExprToIntLitExpr(ast.StmtToExprStatement(program.GetStatements()[0]).GetExpression())
	if expected.String() != actual.String() || expected.GetValue() != actual.GetValue() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected, actual)
	}

	// hexadecimal-literal
	program, err = NewParser().ProduceAST("<?php 0x42;")
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected = ast.NewIntegerLiteralExpression(66)
	actual = ast.ExprToIntLitExpr(ast.StmtToExprStatement(program.GetStatements()[0]).GetExpression())
	if expected.String() != actual.String() || expected.GetValue() != actual.GetValue() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected, actual)
	}

	// binary-literal
	program, err = NewParser().ProduceAST("<?php 0b110110101;")
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected = ast.NewIntegerLiteralExpression(437)
	actual = ast.ExprToIntLitExpr(ast.StmtToExprStatement(program.GetStatements()[0]).GetExpression())
	if expected.String() != actual.String() || expected.GetValue() != actual.GetValue() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected, actual)
	}
}

func TestFloatingLiteral(t *testing.T) {
	program, err := NewParser().ProduceAST("<?php .5;")
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected := ast.NewFloatingLiteralExpression(0.5)
	actual := ast.ExprToFloatLitExpr(ast.StmtToExprStatement(program.GetStatements()[0]).GetExpression())
	if expected.String() != actual.String() || expected.GetValue() != actual.GetValue() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected, actual)
	}

	program, err = NewParser().ProduceAST("<?php 1.2;")
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected = ast.NewFloatingLiteralExpression(1.2)
	actual = ast.ExprToFloatLitExpr(ast.StmtToExprStatement(program.GetStatements()[0]).GetExpression())
	if expected.String() != actual.String() || expected.GetValue() != actual.GetValue() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected, actual)
	}

	program, err = NewParser().ProduceAST("<?php .5e-4;")
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected = ast.NewFloatingLiteralExpression(0.5e-4)
	actual = ast.ExprToFloatLitExpr(ast.StmtToExprStatement(program.GetStatements()[0]).GetExpression())
	if expected.String() != actual.String() || expected.GetValue() != actual.GetValue() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected, actual)
	}

	program, err = NewParser().ProduceAST("<?php 2.5e+4;")
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected = ast.NewFloatingLiteralExpression(2.5e+4)
	actual = ast.ExprToFloatLitExpr(ast.StmtToExprStatement(program.GetStatements()[0]).GetExpression())
	if expected.String() != actual.String() || expected.GetValue() != actual.GetValue() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected, actual)
	}

	program, err = NewParser().ProduceAST("<?php 2e4;")
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected = ast.NewFloatingLiteralExpression(2e4)
	actual = ast.ExprToFloatLitExpr(ast.StmtToExprStatement(program.GetStatements()[0]).GetExpression())
	if expected.String() != actual.String() || expected.GetValue() != actual.GetValue() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected, actual)
	}
}

func TestStringLiteral(t *testing.T) {
	program, err := NewParser().ProduceAST(`<?php b'A "single quoted" \'string\'';`)
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected := ast.NewStringLiteralExpression(`A "single quoted" \'string\'`, ast.SingleQuotedString)
	actual := ast.ExprToStrLitExpr(ast.StmtToExprStatement(program.GetStatements()[0]).GetExpression())
	if expected.String() != actual.String() ||
		expected.GetValue() != actual.GetValue() || expected.GetStringType() != actual.GetStringType() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected, actual)
	}

	program, err = NewParser().ProduceAST(`<?php 'A "single quoted" \'string\'';`)
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected = ast.NewStringLiteralExpression(`A "single quoted" \'string\'`, ast.SingleQuotedString)
	actual = ast.ExprToStrLitExpr(ast.StmtToExprStatement(program.GetStatements()[0]).GetExpression())
	if expected.String() != actual.String() ||
		expected.GetValue() != actual.GetValue() || expected.GetStringType() != actual.GetStringType() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected, actual)
	}

	program, err = NewParser().ProduceAST(`<?php b"A \"single quoted\" 'string'";`)
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected = ast.NewStringLiteralExpression(`A \"single quoted\" 'string'`, ast.DoubleQuotedString)
	actual = ast.ExprToStrLitExpr(ast.StmtToExprStatement(program.GetStatements()[0]).GetExpression())
	if expected.String() != actual.String() ||
		expected.GetValue() != actual.GetValue() || expected.GetStringType() != actual.GetStringType() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected, actual)
	}

	program, err = NewParser().ProduceAST(`<?php "A \"single quoted\" 'string'";`)
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected = ast.NewStringLiteralExpression(`A \"single quoted\" 'string'`, ast.DoubleQuotedString)
	actual = ast.ExprToStrLitExpr(ast.StmtToExprStatement(program.GetStatements()[0]).GetExpression())
	if expected.String() != actual.String() ||
		expected.GetValue() != actual.GetValue() || expected.GetStringType() != actual.GetStringType() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected, actual)
	}
}

func TestEchoStatement(t *testing.T) {
	program, err := NewParser().ProduceAST(`<?php echo 12, "abc", $var;`)
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}

	expected := ast.NewIntegerLiteralExpression(12)
	actual := ast.ExprToIntLitExpr(ast.StmtToEchoStatement(program.GetStatements()[0]).GetExpressions()[0])
	if expected.String() != actual.String() || expected.GetValue() != actual.GetValue() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected, actual)
	}

	expected1 := ast.NewStringLiteralExpression("abc", ast.DoubleQuotedString)
	actual1 := ast.ExprToStrLitExpr(ast.StmtToEchoStatement(program.GetStatements()[0]).GetExpressions()[1])
	if expected1.String() != actual1.String() ||
		expected1.GetValue() != actual1.GetValue() || expected1.GetStringType() != actual1.GetStringType() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected1, actual1)
	}

	expected2 := ast.NewSimpleVariableExpression(ast.NewVariableNameExpression("$var"))
	actual2 := ast.ExprToSimpleVarExpr(ast.StmtToEchoStatement(program.GetStatements()[0]).GetExpressions()[2])
	if expected2.String() != actual2.String() || expected2.GetVariableName().GetKind() != actual2.GetVariableName().GetKind() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected2, actual2)
	}
}

func TestAssignmentExpression(t *testing.T) {
	// SimpleAssignmentExpression
	program, err := NewParser().ProduceAST(`<?php $variable = "abc";`)
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected := ast.NewSimpleAssignmentExpression(
		ast.NewSimpleVariableExpression(ast.NewVariableNameExpression("$variable")),
		ast.NewStringLiteralExpression("abc", ast.DoubleQuotedString),
	)
	actual := ast.ExprToSimpleAssignExpr(ast.StmtToExprStatement(program.GetStatements()[0]).GetExpression())
	if expected.String() != actual.String() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected, actual)
	}

	// CompoundAssignmentExpression
	program, err = NewParser().ProduceAST(`<?php $a = 42; $a += 2;`)
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected1 := ast.NewCompoundAssignmentExpression(
		ast.NewSimpleVariableExpression(ast.NewVariableNameExpression("$a")),
		"+",
		ast.NewIntegerLiteralExpression(2),
	)
	actual1 := ast.ExprToCompoundAssignExpr(ast.StmtToExprStatement(program.GetStatements()[1]).GetExpression())
	if expected1.String() != actual1.String() || expected1.GetOperator() != "+" {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected1, actual1)
	}

	// ConditionalExpression
	program, err = NewParser().ProduceAST(`<?php 1 ? "a" : "b";`)
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected2 := ast.NewConditionalExpression(
		ast.NewIntegerLiteralExpression(1),
		ast.NewStringLiteralExpression("a", ast.DoubleQuotedString),
		ast.NewStringLiteralExpression("b", ast.DoubleQuotedString),
	)
	actual2 := ast.ExprToCondExpr(ast.StmtToExprStatement(program.GetStatements()[0]).GetExpression())
	if expected2.String() != actual2.String() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected2, actual2)
	}

	program, err = NewParser().ProduceAST(`<?php 1 ?: "b";`)
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected2 = ast.NewConditionalExpression(
		ast.NewIntegerLiteralExpression(1),
		nil,
		ast.NewStringLiteralExpression("b", ast.DoubleQuotedString),
	)
	actual2 = ast.ExprToCondExpr(ast.StmtToExprStatement(program.GetStatements()[0]).GetExpression())
	if expected2.String() != actual2.String() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected2, actual2)
	}

	program, err = NewParser().ProduceAST(`<?php 1 ? "a" : 2 ? "b": "c";`)
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected2 = ast.NewConditionalExpression(
		ast.NewIntegerLiteralExpression(1),
		ast.NewStringLiteralExpression("a", ast.DoubleQuotedString),
		ast.NewConditionalExpression(
			ast.NewIntegerLiteralExpression(2),
			ast.NewStringLiteralExpression("b", ast.DoubleQuotedString),
			ast.NewStringLiteralExpression("c", ast.DoubleQuotedString),
		),
	)
	actual2 = ast.ExprToCondExpr(ast.StmtToExprStatement(program.GetStatements()[0]).GetExpression())
	if expected2.String() != actual2.String() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected2, actual2)
	}
}

func TestCoalesceExpression(t *testing.T) {
	program, err := NewParser().ProduceAST(`<?php "a" ?? "b";`)
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected := ast.NewCoalesceExpression(
		ast.NewStringLiteralExpression("a", ast.DoubleQuotedString),
		ast.NewStringLiteralExpression("b", ast.DoubleQuotedString),
	)
	actual := ast.ExprToCoalesceExpr(ast.StmtToExprStatement(program.GetStatements()[0]).GetExpression())
	if expected.String() != actual.String() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected, actual)
	}

	program, err = NewParser().ProduceAST(`<?php "a" ?? "b" ?? "c";`)
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected = ast.NewCoalesceExpression(
		ast.NewStringLiteralExpression("a", ast.DoubleQuotedString),
		ast.NewCoalesceExpression(
			ast.NewStringLiteralExpression("b", ast.DoubleQuotedString),
			ast.NewStringLiteralExpression("c", ast.DoubleQuotedString),
		),
	)
	actual = ast.ExprToCoalesceExpr(ast.StmtToExprStatement(program.GetStatements()[0]).GetExpression())
	if expected.String() != actual.String() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected, actual)
	}
}

func TestConstDeclaration(t *testing.T) {
	program, err := NewParser().ProduceAST(`<?php const PI = 3.141, ZERO = 0;`)
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected := ast.NewConstDeclarationStatement("PI", ast.NewFloatingLiteralExpression(3.141))
	actual := ast.StmtToConstDeclStatement(program.GetStatements()[0])
	if expected.String() != actual.String() || expected.GetName() != actual.GetName() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected, actual)
	}
	expected = ast.NewConstDeclarationStatement("ZERO", ast.NewIntegerLiteralExpression(0))
	actual = ast.StmtToConstDeclStatement(program.GetStatements()[1])
	if expected.String() != actual.String() || expected.GetName() != actual.GetName() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected, actual)
	}
}

func TestEqualityExpression(t *testing.T) {
	program, err := NewParser().ProduceAST(`<?php "234" !== true;`)
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected := ast.NewEqualityExpression(
		ast.NewStringLiteralExpression("234", ast.DoubleQuotedString), "!==", ast.NewBooleanLiteralExpression(true),
	)
	actual := ast.ExprToEqualExpr(ast.StmtToExprStatement(program.GetStatements()[0]).GetExpression())
	if expected.String() != actual.String() || expected.GetOperator() != actual.GetOperator() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected, actual)
	}
}

func TestShiftExpression(t *testing.T) {
	program, err := NewParser().ProduceAST(`<?php "234" << 23;`)
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected := ast.NewShiftExpression(
		ast.NewStringLiteralExpression("234", ast.DoubleQuotedString), "<<", ast.NewIntegerLiteralExpression(23),
	)
	actual := ast.ExprToEqualExpr(ast.StmtToExprStatement(program.GetStatements()[0]).GetExpression())
	if expected.String() != actual.String() || expected.GetOperator() != actual.GetOperator() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected, actual)
	}
}

func TestAdditiveExpression(t *testing.T) {
	program, err := NewParser().ProduceAST(`<?php "234" + 23;`)
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected := ast.NewAdditiveExpression(
		ast.NewStringLiteralExpression("234", ast.DoubleQuotedString), "+", ast.NewIntegerLiteralExpression(23),
	)
	actual := ast.ExprToEqualExpr(ast.StmtToExprStatement(program.GetStatements()[0]).GetExpression())
	if expected.String() != actual.String() || expected.GetOperator() != actual.GetOperator() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected, actual)
	}
}

func TestMultiplicativeExpression(t *testing.T) {
	program, err := NewParser().ProduceAST(`<?php "234" * 12;`)
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected := ast.NewMultiplicativeExpression(
		ast.NewStringLiteralExpression("234", ast.DoubleQuotedString), "*", ast.NewIntegerLiteralExpression(12),
	)
	actual := ast.ExprToEqualExpr(ast.StmtToExprStatement(program.GetStatements()[0]).GetExpression())
	if expected.String() != actual.String() || expected.GetOperator() != actual.GetOperator() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected, actual)
	}
}

func TestExponentiationExpression(t *testing.T) {
	program, err := NewParser().ProduceAST(`<?php "234" ** 12;`)
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected := ast.NewExponentiationExpression(
		ast.NewStringLiteralExpression("234", ast.DoubleQuotedString), ast.NewIntegerLiteralExpression(12),
	)
	actual := ast.ExprToEqualExpr(ast.StmtToExprStatement(program.GetStatements()[0]).GetExpression())
	if expected.String() != actual.String() || expected.GetOperator() != actual.GetOperator() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected, actual)
	}
}

func TestBitwiseIncOrExpression(t *testing.T) {
	program, err := NewParser().ProduceAST(`<?php "234" | 12;`)
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected := ast.NewBitwiseIncOrExpression(
		ast.NewStringLiteralExpression("234", ast.DoubleQuotedString), ast.NewIntegerLiteralExpression(12),
	)
	actual := ast.ExprToEqualExpr(ast.StmtToExprStatement(program.GetStatements()[0]).GetExpression())
	if expected.String() != actual.String() || expected.GetOperator() != actual.GetOperator() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected, actual)
	}
}

func TestLogicalIncOrExpression(t *testing.T) {
	program, err := NewParser().ProduceAST(`<?php "234" || 12;`)
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected := ast.NewLogicalIncOrExpression(
		ast.NewStringLiteralExpression("234", ast.DoubleQuotedString), ast.NewIntegerLiteralExpression(12),
	)
	actual := ast.ExprToEqualExpr(ast.StmtToExprStatement(program.GetStatements()[0]).GetExpression())
	if expected.String() != actual.String() || expected.GetOperator() != actual.GetOperator() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected, actual)
	}
}

func TestLogicalAndExpression(t *testing.T) {
	program, err := NewParser().ProduceAST(`<?php "234" && 12;`)
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected := ast.NewLogicalAndExpression(
		ast.NewStringLiteralExpression("234", ast.DoubleQuotedString), ast.NewIntegerLiteralExpression(12),
	)
	actual := ast.ExprToEqualExpr(ast.StmtToExprStatement(program.GetStatements()[0]).GetExpression())
	if expected.String() != actual.String() || expected.GetOperator() != actual.GetOperator() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected, actual)
	}
}

func TestBitwiseExcOrExpression(t *testing.T) {
	program, err := NewParser().ProduceAST(`<?php "234" ^ 12;`)
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected := ast.NewBitwiseExcOrExpression(
		ast.NewStringLiteralExpression("234", ast.DoubleQuotedString), ast.NewIntegerLiteralExpression(12),
	)
	actual := ast.ExprToEqualExpr(ast.StmtToExprStatement(program.GetStatements()[0]).GetExpression())
	if expected.String() != actual.String() || expected.GetOperator() != actual.GetOperator() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected, actual)
	}
}

func TestBitwiseAndExpression(t *testing.T) {
	program, err := NewParser().ProduceAST(`<?php "234" & 12;`)
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected := ast.NewBitwiseAndExpression(
		ast.NewStringLiteralExpression("234", ast.DoubleQuotedString), ast.NewIntegerLiteralExpression(12),
	)
	actual := ast.ExprToEqualExpr(ast.StmtToExprStatement(program.GetStatements()[0]).GetExpression())
	if expected.String() != actual.String() || expected.GetOperator() != actual.GetOperator() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected, actual)
	}
}

func TestLogicalNotExpression(t *testing.T) {
	program, err := NewParser().ProduceAST(`<?php !true;`)
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected := ast.NewLogicalNotExpression(ast.NewBooleanLiteralExpression(true))
	actual := ast.ExprToUnaryOpExpr(ast.StmtToExprStatement(program.GetStatements()[0]).GetExpression())
	if expected.String() != actual.String() || expected.GetOperator() != actual.GetOperator() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected, actual)
	}
}

func TestParenthesizedExpression(t *testing.T) {
	program, err := NewParser().ProduceAST(`<?php (1+2);`)
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected := ast.NewAdditiveExpression(
		ast.NewIntegerLiteralExpression(1),
		"+",
		ast.NewIntegerLiteralExpression(2),
	)
	actual := ast.ExprToEqualExpr(ast.StmtToExprStatement(program.GetStatements()[0]).GetExpression())
	if expected.String() != actual.String() || expected.GetOperator() != actual.GetOperator() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected, actual)
	}
}

func TestEmptyIntrinsic(t *testing.T) {
	program, err := NewParser().ProduceAST(`<?php empty(false);`)
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected := ast.NewEmptyIntrinsic(ast.NewBooleanLiteralExpression(false))
	actual := ast.ExprToFuncCallExpr(ast.StmtToExprStatement(program.GetStatements()[0]).GetExpression())
	if expected.String() != actual.String() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected, actual)
	}
}

func TestIssetIntrinsic(t *testing.T) {
	program, err := NewParser().ProduceAST(`<?php isset($a);`)
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected := ast.NewIssetIntrinsic([]ast.IExpression{ast.NewSimpleVariableExpression(ast.NewVariableNameExpression("$a"))})
	actual := ast.ExprToFuncCallExpr(ast.StmtToExprStatement(program.GetStatements()[0]).GetExpression())
	if expected.String() != actual.String() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected, actual)
	}
}

func TestUnsetIntrinsic(t *testing.T) {
	program, err := NewParser().ProduceAST(`<?php unset($a);`)
	if err != nil {
		t.Errorf("Unexpected error: \"%s\"", err)
		return
	}
	expected := ast.NewUnsetIntrinsic([]ast.IExpression{ast.NewSimpleVariableExpression(ast.NewVariableNameExpression("$a"))})
	actual := ast.ExprToFuncCallExpr(ast.StmtToExprStatement(program.GetStatements()[0]).GetExpression())
	if expected.String() != actual.String() {
		t.Errorf("Expected: \"%s\", Got \"%s\"", expected, actual)
	}
}

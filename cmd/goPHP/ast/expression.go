package ast

import (
	"GoPHP/cmd/goPHP/position"
	"fmt"
)

// ------------------- MARK: Expression -------------------

type IExpression interface {
	IStatement
}

type Expression struct {
	id   int64
	kind NodeType
	pos  *position.Position
}

func NewExpr(id int64, kind NodeType, pos *position.Position) *Expression {
	return &Expression{id: id, kind: kind, pos: pos}
}

func (stmt *Expression) GetId() int64 {
	return stmt.id
}

func (expr *Expression) GetKind() NodeType {
	return expr.kind
}

func (expr *Expression) GetPosition() *position.Position {
	return expr.pos
}

func (expr *Expression) String() string {
	return fmt.Sprintf("{%s}", expr.GetKind())
}

func NewEmptyExpr() *Expression {
	return NewExpr(0, EmptyNode, nil)
}

// ------------------- MARK: TextExpression -------------------

type ITextExpression interface {
	IExpression
	GetValue() string
}

type TextExpression struct {
	expr  IExpression
	value string
}

func NewTextExpr(id int64, value string) *TextExpression {
	return &TextExpression{expr: NewExpr(id, TextNode, nil), value: value}
}

func (expr *TextExpression) GetId() int64 {
	return expr.expr.GetId()
}

func (expr *TextExpression) GetKind() NodeType {
	return expr.expr.GetKind()
}

func (expr *TextExpression) GetPosition() *position.Position {
	return expr.expr.GetPosition()
}

func (expr *TextExpression) GetValue() string {
	return expr.value
}

func (expr *TextExpression) String() string {
	return fmt.Sprintf("{%s - value: \"%s\" }", expr.GetKind(), expr.value)
}

func ExprToTextExpr(expr IExpression) ITextExpression {
	var i interface{} = expr
	return i.(ITextExpression)
}

// ------------------- MARK: VariableNameExpression -------------------

type IVariableNameExpression interface {
	IExpression
	GetVariableName() string
}

type VariableNameExpression struct {
	expr         IExpression
	variableName string
}

func NewVariableNameExpr(id int64, pos *position.Position, variableName string) *VariableNameExpression {
	return &VariableNameExpression{expr: NewExpr(id, VariableNameExpr, pos), variableName: variableName}
}

func (expr *VariableNameExpression) GetId() int64 {
	return expr.expr.GetId()
}

func (expr *VariableNameExpression) GetKind() NodeType {
	return expr.expr.GetKind()
}

func (expr *VariableNameExpression) GetPosition() *position.Position {
	return expr.expr.GetPosition()
}

func (expr *VariableNameExpression) GetVariableName() string {
	return expr.variableName
}

func (expr *VariableNameExpression) String() string {
	return fmt.Sprintf("{%s - variableName: \"%s\" }", expr.GetKind(), expr.variableName)
}

func ExprToVarNameExpr(expr IExpression) IVariableNameExpression {
	var i interface{} = expr
	return i.(IVariableNameExpression)
}

// ------------------- MARK: SimpleVariableExpression -------------------

type ISimpleVariableExpression interface {
	IExpression
	GetVariableName() IExpression
}

type SimpleVariableExpression struct {
	expr         IExpression
	variableName IExpression
}

func NewSimpleVariableExpr(id int64, variableName IExpression) *SimpleVariableExpression {
	return &SimpleVariableExpression{expr: NewExpr(id, SimpleVariableExpr, variableName.GetPosition()), variableName: variableName}
}

func (expr *SimpleVariableExpression) GetId() int64 {
	return expr.expr.GetId()
}

func (expr *SimpleVariableExpression) GetKind() NodeType {
	return expr.expr.GetKind()
}

func (expr *SimpleVariableExpression) GetPosition() *position.Position {
	return expr.expr.GetPosition()
}

func (expr *SimpleVariableExpression) GetVariableName() IExpression {
	return expr.variableName
}

func (expr *SimpleVariableExpression) String() string {
	return fmt.Sprintf("{%s - variableName: \"%s\" }", expr.GetKind(), expr.variableName)
}

func ExprToSimpleVarExpr(expr IExpression) ISimpleVariableExpression {
	var i interface{} = expr
	return i.(ISimpleVariableExpression)
}

// ------------------- MARK: SubscriptExpression -------------------

type ISubscriptExpression interface {
	IExpression
	GetVariable() IExpression
	GetIndex() IExpression
}

type SubscriptExpression struct {
	expr     IExpression
	variable IExpression
	index    IExpression
}

func NewSubscriptExpr(id int64, variable IExpression, index IExpression) *SubscriptExpression {
	return &SubscriptExpression{expr: NewExpr(id, SubscriptExpr, variable.GetPosition()), variable: variable, index: index}
}

func (expr *SubscriptExpression) GetId() int64 {
	return expr.expr.GetId()
}

func (expr *SubscriptExpression) GetKind() NodeType {
	return expr.expr.GetKind()
}

func (expr *SubscriptExpression) GetPosition() *position.Position {
	return expr.expr.GetPosition()
}

func (expr *SubscriptExpression) GetVariable() IExpression {
	return expr.variable
}

func (expr *SubscriptExpression) GetIndex() IExpression {
	return expr.index
}

func (expr *SubscriptExpression) String() string {
	return fmt.Sprintf("{%s - variable: %s, index: \"%s\" }", expr.GetKind(), expr.variable, expr.index)
}

func ExprToSubscriptExpr(expr IExpression) ISubscriptExpression {
	var i interface{} = expr
	return i.(ISubscriptExpression)
}

// ------------------- MARK: FunctionCallExpression -------------------

type IFunctionCallExpression interface {
	IExpression
	GetFunctionName() string
	GetArguments() []IExpression
}

type FunctionCallExpression struct {
	expr         IExpression
	functionName string
	arguments    []IExpression
}

func NewFunctionCallExpr(id int64, pos *position.Position, functionName string, arguments []IExpression) *FunctionCallExpression {
	return &FunctionCallExpression{expr: NewExpr(id, FunctionCallExpr, pos), functionName: functionName, arguments: arguments}
}

func (expr *FunctionCallExpression) GetId() int64 {
	return expr.expr.GetId()
}

func (expr *FunctionCallExpression) GetKind() NodeType {
	return expr.expr.GetKind()
}

func (expr *FunctionCallExpression) GetPosition() *position.Position {
	return expr.expr.GetPosition()
}

func (expr *FunctionCallExpression) GetFunctionName() string {
	return expr.functionName
}

func (expr *FunctionCallExpression) GetArguments() []IExpression {
	return expr.arguments
}

func (expr *FunctionCallExpression) String() string {
	return fmt.Sprintf("{%s - functionName: \"%s\" arguments: %s}", expr.GetKind(), expr.functionName, expr.arguments)
}

func ExprToFuncCallExpr(expr IExpression) IFunctionCallExpression {
	var i interface{} = expr
	return i.(IFunctionCallExpression)
}

// ------------------- MARK: EmptyIntrinsic -------------------

func NewExitIntrinsic(id int64, pos *position.Position, expression IExpression) *FunctionCallExpression {
	return &FunctionCallExpression{expr: NewExpr(id, ExitIntrinsicExpr, pos),
		functionName: "exit", arguments: []IExpression{expression},
	}
}

// ------------------- MARK: EmptyIntrinsic -------------------

func NewEmptyIntrinsic(id int64, pos *position.Position, expression IExpression) *FunctionCallExpression {
	return &FunctionCallExpression{expr: NewExpr(id, EmptyIntrinsicExpr, pos),
		functionName: "empty", arguments: []IExpression{expression},
	}
}

// ------------------- MARK: IssetIntrinsic -------------------

func NewIssetIntrinsic(id int64, pos *position.Position, arguments []IExpression) *FunctionCallExpression {
	return &FunctionCallExpression{expr: NewExpr(id, IssetIntrinsicExpr, pos),
		functionName: "isset", arguments: arguments,
	}
}

// ------------------- MARK: UnsetIntrinsic -------------------

func NewUnsetIntrinsic(id int64, pos *position.Position, arguments []IExpression) *FunctionCallExpression {
	return &FunctionCallExpression{expr: NewExpr(id, UnsetIntrinsicExpr, pos),
		functionName: "unset", arguments: arguments,
	}
}

// ------------------- MARK: ConstantAccessExpression -------------------

type IConstantAccessExpression interface {
	IExpression
	GetConstantName() string
}

type ConstantAccessExpression struct {
	expr         IExpression
	constantName string
}

func NewConstantAccessExpr(id int64, pos *position.Position, constantName string) *ConstantAccessExpression {
	return &ConstantAccessExpression{expr: NewExpr(id, ConstantAccessExpr, pos), constantName: constantName}
}

func (expr *ConstantAccessExpression) GetId() int64 {
	return expr.expr.GetId()
}

func (expr *ConstantAccessExpression) GetKind() NodeType {
	return expr.expr.GetKind()
}

func (expr *ConstantAccessExpression) GetPosition() *position.Position {
	return expr.expr.GetPosition()
}

func (expr *ConstantAccessExpression) GetConstantName() string {
	return expr.constantName
}

func (expr *ConstantAccessExpression) String() string {
	return fmt.Sprintf("{%s - constantName: %s}", expr.GetKind(), expr.constantName)
}

func ExprToConstAccessExpr(expr IExpression) IConstantAccessExpression {
	var i interface{} = expr
	return i.(IConstantAccessExpression)
}

// ------------------- MARK: ArrayLiteralExpression -------------------

type IArrayLiteralExpression interface {
	IExpression
	GetKeys() []IExpression
	AddElement(key IExpression, value IExpression)
	GetElements() map[IExpression]IExpression
}

type ArrayLiteralExpression struct {
	expr     IExpression
	keys     []IExpression
	elements map[IExpression]IExpression
}

func NewArrayLiteralExpr(id int64, pos *position.Position) *ArrayLiteralExpression {
	return &ArrayLiteralExpression{
		expr:     NewExpr(id, ArrayLiteralExpr, pos),
		keys:     []IExpression{},
		elements: map[IExpression]IExpression{},
	}
}

func (expr *ArrayLiteralExpression) GetId() int64 {
	return expr.expr.GetId()
}

func (expr *ArrayLiteralExpression) GetKind() NodeType {
	return expr.expr.GetKind()
}

func (expr *ArrayLiteralExpression) GetPosition() *position.Position {
	return expr.expr.GetPosition()
}

func (expr *ArrayLiteralExpression) AddElement(key IExpression, value IExpression) {
	expr.keys = append(expr.keys, key)
	expr.elements[key] = value
}

func (expr *ArrayLiteralExpression) GetKeys() []IExpression {
	return expr.keys
}

func (expr *ArrayLiteralExpression) GetElements() map[IExpression]IExpression {
	return expr.elements
}

func (expr *ArrayLiteralExpression) String() string {
	return fmt.Sprintf("{%s - elements: %s }", expr.GetKind(), expr.elements)
}

func ExprToArrayLitExpr(expr IExpression) IArrayLiteralExpression {
	var i interface{} = expr
	return i.(IArrayLiteralExpression)
}

// ------------------- MARK: BooleanLiteralExpression -------------------

func NewBooleanLiteralExpr(id int64, pos *position.Position, value bool) *ConstantAccessExpression {
	if value {
		return NewConstantAccessExpr(id, pos, "TRUE")
	}
	return NewConstantAccessExpr(id, pos, "FALSE")
}

// ------------------- MARK: IntegerLiteralExpression -------------------

type IIntegerLiteralExpression interface {
	IExpression
	GetValue() int64
}

type IntegerLiteralExpression struct {
	expr  IExpression
	value int64
}

func NewIntegerLiteralExpr(id int64, pos *position.Position, value int64) *IntegerLiteralExpression {
	return &IntegerLiteralExpression{expr: NewExpr(id, IntegerLiteralExpr, pos), value: value}
}

func (expr *IntegerLiteralExpression) GetId() int64 {
	return expr.expr.GetId()
}

func (expr *IntegerLiteralExpression) GetKind() NodeType {
	return expr.expr.GetKind()
}

func (expr *IntegerLiteralExpression) GetPosition() *position.Position {
	return expr.expr.GetPosition()
}

func (expr *IntegerLiteralExpression) GetValue() int64 {
	return expr.value
}

func (expr *IntegerLiteralExpression) String() string {
	return fmt.Sprintf("{%s - value: %d }", expr.GetKind(), expr.value)
}

func ExprToIntLitExpr(expr IExpression) IIntegerLiteralExpression {
	var i interface{} = expr
	return i.(IIntegerLiteralExpression)
}

// ------------------- MARK: FloatingLiteralExpression -------------------

type IFloatingLiteralExpression interface {
	IExpression
	GetValue() float64
}

type FloatingLiteralExpression struct {
	expr  IExpression
	value float64
}

func NewFloatingLiteralExpr(id int64, pos *position.Position, value float64) *FloatingLiteralExpression {
	return &FloatingLiteralExpression{expr: NewExpr(id, FloatingLiteralExpr, pos), value: value}
}

func (expr *FloatingLiteralExpression) GetId() int64 {
	return expr.expr.GetId()
}

func (expr *FloatingLiteralExpression) GetKind() NodeType {
	return expr.expr.GetKind()
}

func (expr *FloatingLiteralExpression) GetPosition() *position.Position {
	return expr.expr.GetPosition()
}

func (expr *FloatingLiteralExpression) GetValue() float64 {
	return expr.value
}

func (expr *FloatingLiteralExpression) String() string {
	return fmt.Sprintf("{%s - value: %f }", expr.GetKind(), expr.value)
}

func ExprToFloatLitExpr(expr IExpression) IFloatingLiteralExpression {
	var i interface{} = expr
	return i.(IFloatingLiteralExpression)
}

// ------------------- MARK: StringLiteralExpression -------------------

type StringType string

const (
	SingleQuotedString StringType = "SingleQuotedString"
	DoubleQuotedString StringType = "DoubleQuotedString"
)

type IStringLiteralExpression interface {
	IExpression
	GetStringType() StringType
	GetValue() string
}

type StringLiteralExpression struct {
	expr       IExpression
	stringType StringType
	value      string
}

func NewStringLiteralExpr(id int64, pos *position.Position, value string, stringType StringType) *StringLiteralExpression {
	return &StringLiteralExpression{expr: NewExpr(id, StringLiteralExpr, pos), value: value, stringType: stringType}
}

func (expr *StringLiteralExpression) GetId() int64 {
	return expr.expr.GetId()
}

func (expr *StringLiteralExpression) GetKind() NodeType {
	return expr.expr.GetKind()
}

func (expr *StringLiteralExpression) GetPosition() *position.Position {
	return expr.expr.GetPosition()
}

func (expr *StringLiteralExpression) GetStringType() StringType {
	return expr.stringType
}

func (expr *StringLiteralExpression) GetValue() string {
	return expr.value
}

func (expr *StringLiteralExpression) String() string {
	return fmt.Sprintf("{%s - type: \"%s\" value: \"%s\" }", expr.GetKind(), expr.stringType, expr.value)
}

func ExprToStrLitExpr(expr IExpression) IStringLiteralExpression {
	var i interface{} = expr
	return i.(IStringLiteralExpression)
}

// ------------------- MARK: NullLiteralExpression -------------------

func NewNullLiteralExpr(id int64, pos *position.Position) *ConstantAccessExpression {
	return NewConstantAccessExpr(id, pos, "NULL")
}

// ------------------- MARK: SimpleAssignmentExpression -------------------

type ISimpleAssignmentExpression interface {
	IExpression
	GetVariable() IExpression
	GetValue() IExpression
}

type SimpleAssignmentExpression struct {
	expr     IExpression
	variable IExpression
	value    IExpression
}

func NewSimpleAssignmentExpr(id int64, variable IExpression, value IExpression) *SimpleAssignmentExpression {
	return &SimpleAssignmentExpression{expr: NewExpr(id, SimpleAssignmentExpr, variable.GetPosition()), variable: variable, value: value}
}

func (expr *SimpleAssignmentExpression) GetId() int64 {
	return expr.expr.GetId()
}

func (expr *SimpleAssignmentExpression) GetKind() NodeType {
	return expr.expr.GetKind()
}

func (expr *SimpleAssignmentExpression) GetPosition() *position.Position {
	return expr.expr.GetPosition()
}

func (expr *SimpleAssignmentExpression) GetVariable() IExpression {
	return expr.variable
}

func (expr *SimpleAssignmentExpression) GetValue() IExpression {
	return expr.value
}

func (expr *SimpleAssignmentExpression) String() string {
	return fmt.Sprintf("{%s - variable: %s, value: %s }", expr.GetKind(), expr.variable, expr.value)
}

func ExprToSimpleAssignExpr(expr IExpression) ISimpleAssignmentExpression {
	var i interface{} = expr
	return i.(ISimpleAssignmentExpression)
}

// ------------------- MARK: CompoundAssignmentExpression -------------------

type ICompoundAssignmentExpression interface {
	IExpression
	GetVariable() IExpression
	GetOperator() string
	GetValue() IExpression
}

type CompoundAssignmentExpression struct {
	expr     IExpression
	variable IExpression
	operator string
	value    IExpression
}

func NewCompoundAssignmentExpr(id int64, variable IExpression, operator string, value IExpression) *CompoundAssignmentExpression {
	return &CompoundAssignmentExpression{
		expr: NewExpr(id, CompoundAssignmentExpr, variable.GetPosition()), variable: variable, operator: operator, value: value,
	}
}

func (expr *CompoundAssignmentExpression) GetId() int64 {
	return expr.expr.GetId()
}

func (expr *CompoundAssignmentExpression) GetKind() NodeType {
	return expr.expr.GetKind()
}

func (expr *CompoundAssignmentExpression) GetPosition() *position.Position {
	return expr.expr.GetPosition()
}

func (expr *CompoundAssignmentExpression) GetVariable() IExpression {
	return expr.variable
}

func (expr *CompoundAssignmentExpression) GetOperator() string {
	return expr.operator
}

func (expr *CompoundAssignmentExpression) GetValue() IExpression {
	return expr.value
}

func (expr *CompoundAssignmentExpression) String() string {
	return fmt.Sprintf(
		"{%s - variable: %s, operator: \"%s\", value: %s }",
		expr.GetKind(), expr.variable, expr.operator, expr.value,
	)
}

func ExprToCompoundAssignExpr(expr IExpression) ICompoundAssignmentExpression {
	var i interface{} = expr
	return i.(ICompoundAssignmentExpression)
}

// ------------------- MARK: ConditionalExpression -------------------

type IConditionalExpression interface {
	IExpression
	GetCondition() IExpression
	GetIfExpr() IExpression
	GetElseExpr() IExpression
}

type ConditionalExpression struct {
	expr     IExpression
	cond     IExpression
	ifExpr   IExpression
	elseExpr IExpression
}

func NewConditionalExpr(id int64, cond IExpression, ifExpr IExpression, elseExpr IExpression) *ConditionalExpression {
	return &ConditionalExpression{expr: NewExpr(id, ConditionalExpr, cond.GetPosition()), cond: cond, ifExpr: ifExpr, elseExpr: elseExpr}
}

func (expr *ConditionalExpression) GetId() int64 {
	return expr.expr.GetId()
}

func (expr *ConditionalExpression) GetKind() NodeType {
	return expr.expr.GetKind()
}

func (expr *ConditionalExpression) GetPosition() *position.Position {
	return expr.expr.GetPosition()
}

func (expr *ConditionalExpression) GetCondition() IExpression {
	return expr.cond
}

func (expr *ConditionalExpression) GetIfExpr() IExpression {
	return expr.ifExpr
}

func (expr *ConditionalExpression) GetElseExpr() IExpression {
	return expr.elseExpr
}

func (expr *ConditionalExpression) String() string {
	return fmt.Sprintf("{%s - condition: %s, ifExpr: %s, elseExpr: %s }", expr.GetKind(), expr.cond, expr.ifExpr, expr.elseExpr)
}

func ExprToCondExpr(expr IExpression) IConditionalExpression {
	var i interface{} = expr
	return i.(IConditionalExpression)
}

// ------------------- MARK: CoalesceExpression -------------------

type ICoalesceExpression interface {
	IExpression
	GetCondition() IExpression
	GetElseExpr() IExpression
}

type CoalesceExpression struct {
	expr     IExpression
	cond     IExpression
	elseExpr IExpression
}

func NewCoalesceExpr(id int64, cond IExpression, elseExpr IExpression) *CoalesceExpression {
	return &CoalesceExpression{expr: NewExpr(id, CoalesceExpr, cond.GetPosition()), cond: cond, elseExpr: elseExpr}
}

func (expr *CoalesceExpression) GetId() int64 {
	return expr.expr.GetId()
}

func (expr *CoalesceExpression) GetKind() NodeType {
	return expr.expr.GetKind()
}

func (expr *CoalesceExpression) GetPosition() *position.Position {
	return expr.expr.GetPosition()
}

func (expr *CoalesceExpression) GetCondition() IExpression {
	return expr.cond
}

func (expr *CoalesceExpression) GetElseExpr() IExpression {
	return expr.elseExpr
}

func (expr *CoalesceExpression) String() string {
	return fmt.Sprintf("{%s - condition: %s, elseExpr: %s }", expr.GetKind(), expr.cond, expr.elseExpr)
}

func ExprToCoalesceExpr(expr IExpression) ICoalesceExpression {
	var i interface{} = expr
	return i.(ICoalesceExpression)
}

// ------------------- MARK: BinaryOpExpression -------------------

type IBinaryOpExpression interface {
	IExpression
	GetLHS() IExpression
	GetOperator() string
	GetRHS() IExpression
}

type BinaryOpExpression struct {
	expr     IExpression
	lhs      IExpression
	operator string
	rhs      IExpression
}

func NewRelationalExpr(id int64, lhs IExpression, operator string, rhs IExpression) *BinaryOpExpression {
	return &BinaryOpExpression{expr: NewExpr(id, RelationalExpr, lhs.GetPosition()), lhs: lhs, operator: operator, rhs: rhs}
}

func NewEqualityExpr(id int64, lhs IExpression, operator string, rhs IExpression) *BinaryOpExpression {
	return &BinaryOpExpression{expr: NewExpr(id, EqualityExpr, lhs.GetPosition()), lhs: lhs, operator: operator, rhs: rhs}
}

func NewBinaryOpExpr(id int64, lhs IExpression, operator string, rhs IExpression) *BinaryOpExpression {
	return &BinaryOpExpression{expr: NewExpr(id, BinaryOpExpr, lhs.GetPosition()), lhs: lhs, operator: operator, rhs: rhs}
}

func (expr *BinaryOpExpression) GetId() int64 {
	return expr.expr.GetId()
}

func (expr *BinaryOpExpression) GetKind() NodeType {
	return expr.expr.GetKind()
}

func (expr *BinaryOpExpression) GetPosition() *position.Position {
	return expr.expr.GetPosition()
}

func (expr *BinaryOpExpression) GetLHS() IExpression {
	return expr.lhs
}

func (expr *BinaryOpExpression) GetOperator() string {
	return expr.operator
}

func (expr *BinaryOpExpression) GetRHS() IExpression {
	return expr.rhs
}

func (expr *BinaryOpExpression) String() string {
	return fmt.Sprintf("{%s - lhs: %s, operator: \"%s\" rhs: %s }", expr.GetKind(), expr.lhs, expr.operator, expr.rhs)
}

func ExprToBinOpExpr(expr IExpression) IBinaryOpExpression {
	var i interface{} = expr
	return i.(IBinaryOpExpression)
}

// ------------------- MARK: UnaryOpExpression -------------------

type IUnaryOpExpression interface {
	IExpression
	GetOperator() string
	GetExpression() IExpression
}

type UnaryOpExpression struct {
	expr       IExpression
	operator   string
	expression IExpression
}

func NewPrefixIncExpr(id int64, pos *position.Position, expression IExpression, operator string) *UnaryOpExpression {
	return &UnaryOpExpression{expr: NewExpr(id, PrefixIncExpr, pos), operator: operator, expression: expression}
}

func NewPostfixIncExpr(id int64, pos *position.Position, expression IExpression, operator string) *UnaryOpExpression {
	return &UnaryOpExpression{expr: NewExpr(id, PostfixIncExpr, pos), operator: operator, expression: expression}
}

func NewLogicalNotExpr(id int64, pos *position.Position, expression IExpression) *UnaryOpExpression {
	return &UnaryOpExpression{expr: NewExpr(id, LogicalNotExpr, pos), operator: "!", expression: expression}
}

func NewCastExpr(id int64, pos *position.Position, castType string, expression IExpression) *UnaryOpExpression {
	return &UnaryOpExpression{expr: NewExpr(id, CastExpr, pos), operator: castType, expression: expression}
}

func NewUnaryOpExpr(id int64, pos *position.Position, operator string, expression IExpression) *UnaryOpExpression {
	return &UnaryOpExpression{expr: NewExpr(id, UnaryOpExpr, pos), operator: operator, expression: expression}
}

func (expr *UnaryOpExpression) GetId() int64 {
	return expr.expr.GetId()
}

func (expr *UnaryOpExpression) GetKind() NodeType {
	return expr.expr.GetKind()
}

func (expr *UnaryOpExpression) GetPosition() *position.Position {
	return expr.expr.GetPosition()
}

func (expr *UnaryOpExpression) GetOperator() string {
	return expr.operator
}

func (expr *UnaryOpExpression) GetExpression() IExpression {
	return expr.expression
}

func (expr *UnaryOpExpression) String() string {
	return fmt.Sprintf("{%s - operator: \"%s\" expression: %s }", expr.GetKind(), expr.operator, expr.expression)
}

func ExprToUnaryOpExpr(expr IExpression) IUnaryOpExpression {
	var i interface{} = expr
	return i.(IUnaryOpExpression)
}

// ------------------- MARK: ExprExpression -------------------

type IExprExpression interface {
	IExpression
	GetExpression() IExpression
}

type ExprExpression struct {
	expr       IExpression
	expression IExpression
}

func NewIncludeExpr(id int64, pos *position.Position, expression IExpression) *ExprExpression {
	return &ExprExpression{expr: NewExpr(id, IncludeExpr, pos), expression: expression}
}

func NewIncludeOnceExpr(id int64, pos *position.Position, expression IExpression) *ExprExpression {
	return &ExprExpression{expr: NewExpr(id, IncludeOnceExpr, pos), expression: expression}
}

func NewRequireExpr(id int64, pos *position.Position, expression IExpression) *ExprExpression {
	return &ExprExpression{expr: NewExpr(id, RequireExpr, pos), expression: expression}
}

func NewRequireOnceExpr(id int64, pos *position.Position, expression IExpression) *ExprExpression {
	return &ExprExpression{expr: NewExpr(id, RequireOnceExpr, pos), expression: expression}
}

func (expr *ExprExpression) GetId() int64 {
	return expr.expr.GetId()
}

func (expr *ExprExpression) GetKind() NodeType {
	return expr.expr.GetKind()
}

func (expr *ExprExpression) GetPosition() *position.Position {
	return expr.expr.GetPosition()
}

func (expr *ExprExpression) GetExpression() IExpression {
	return expr.expression
}

func (expr *ExprExpression) String() string {
	return fmt.Sprintf("{%s - expression: %s }", expr.GetKind(), expr.expression)
}

func ExprToExprExpr(expr IExpression) IExprExpression {
	var i interface{} = expr
	return i.(IExprExpression)
}

package ast

type NodeType string

const (
	EmptyNode   NodeType = "Empty"
	ProgramNode NodeType = "Program"
	TextNode    NodeType = "Text"
	// Expressions
	ArrayLiteralExpr       NodeType = "ArrayLiteralExpression"
	BooleanLiteralExpr     NodeType = "BooleanLiteralExpression"
	IntegerLiteralExpr     NodeType = "IntegerLiteralExpression"
	FloatingLiteralExpr    NodeType = "FloatingLiteralExpression"
	StringLiteralExpr      NodeType = "StringLiteralExpression"
	NullLiteralExpr        NodeType = "NullLiteralExpression"
	VariableNameExpr       NodeType = "VariableNameExpression"
	SimpleVariableExpr     NodeType = "SimpleVariableExpression"
	SubscriptExpr          NodeType = "SubscriptExpression"
	FunctionCallExpr       NodeType = "FunctionCallExpression"
	EmptyIntrinsicExpr     NodeType = "EmptyIntrinsicExpression"
	IssetIntrinsicExpr     NodeType = "IssetIntrinsicExpression"
	UnsetIntrinsicExpr     NodeType = "UnsetIntrinsicExpression"
	SimpleAssignmentExpr   NodeType = "SimpleAssignmentExpression"
	ConstantAccessExpr     NodeType = "ConstantAccessExpression"
	CompoundAssignmentExpr NodeType = "CompoundAssignmentExpression"
	ConditionalExpr        NodeType = "ConditionalExpression"
	CoalesceExpr           NodeType = "CoalesceExpression"
	EqualityExpr           NodeType = "EqualityExpression"
	AdditiveExpr           NodeType = "AdditiveExpression"
	MultiplicativeExpr     NodeType = "MultiplicativeExpression"
	ExponentiationExpr     NodeType = "ExponentiationExpression"
	UnaryOpExpr            NodeType = "UnaryOpExpression"
	LogicalNotExpr         NodeType = "LogicalNotExpression"
	// Statements
	EchoStmt             NodeType = "EchoStatement"
	ConstDeclarationStmt NodeType = "ConstDeclarationStatement"
	ExpressionStmt       NodeType = "ExpressionStatement"
)

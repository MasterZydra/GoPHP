package ast

type NodeType string

const (
	EmptyNode   NodeType = "Empty"
	ProgramNode NodeType = "Program"
	TextNode    NodeType = "Text"
	// Expressions
	ArrayLiteralExpr       NodeType = "ArrayLiteralExpression"
	ArrayNextKeyExpr       NodeType = "ArrayNextKeyExpression"
	BinaryOpExpr           NodeType = "BinaryOpExpression"
	CastExpr               NodeType = "CastExpression"
	CoalesceExpr           NodeType = "CoalesceExpression"
	CompoundAssignmentExpr NodeType = "CompoundAssignmentExpression"
	ConditionalExpr        NodeType = "ConditionalExpression"
	ConstantAccessExpr     NodeType = "ConstantAccessExpression"
	EmptyIntrinsicExpr     NodeType = "EmptyIntrinsicExpression"
	EqualityExpr           NodeType = "EqualityExpression"
	ErrorControlExpr       NodeType = "ErrorControlExpression"
	EvalIntrinsicExpr      NodeType = "EvalIntrinsicExpression"
	ExitIntrinsicExpr      NodeType = "ExitIntrinsicExpression"
	FloatingLiteralExpr    NodeType = "FloatingLiteralExpression"
	FunctionCallExpr       NodeType = "FunctionCallExpression"
	IncludeExpr            NodeType = "IncludeExpression"
	IncludeOnceExpr        NodeType = "IncludeOnceExpression"
	IntegerLiteralExpr     NodeType = "IntegerLiteralExpression"
	IssetIntrinsicExpr     NodeType = "IssetIntrinsicExpression"
	LogicalNotExpr         NodeType = "LogicalNotExpression"
	MemberAccessExpr       NodeType = "MemberAccessExpression"
	ObjectCreationExpr     NodeType = "ObjectCreationExpression"
	ParenthesizedExpr      NodeType = "ParenthesizedExpression"
	PostfixIncExpr         NodeType = "PostfixIncExpression"
	PrefixIncExpr          NodeType = "PrefixIncExpression"
	PrintExpr              NodeType = "PrintExpression"
	RelationalExpr         NodeType = "RelationalExpression"
	RequireExpr            NodeType = "RequireExpression"
	RequireOnceExpr        NodeType = "RequireOnceExpression"
	ShiftExpr              NodeType = "ShiftExpression"
	SimpleAssignmentExpr   NodeType = "SimpleAssignmentExpression"
	SimpleVariableExpr     NodeType = "SimpleVariableExpression"
	StringLiteralExpr      NodeType = "StringLiteralExpression"
	SubscriptExpr          NodeType = "SubscriptExpression"
	UnaryOpExpr            NodeType = "UnaryOpExpression"
	UnsetIntrinsicExpr     NodeType = "UnsetIntrinsicExpression"
	VariableNameExpr       NodeType = "VariableNameExpression"
	// Statements
	BreakStmt              NodeType = "BreakStatement"
	CompoundStmt           NodeType = "CompoundStatement"
	ConstDeclarationStmt   NodeType = "ConstDeclarationStatement"
	ContinueStmt           NodeType = "ContinueStatement"
	DeclareStmt            NodeType = "DeclareStatement"
	DoStmt                 NodeType = "DoStatement"
	EchoStmt               NodeType = "EchoStatement"
	ExpressionStmt         NodeType = "ExpressionStatement"
	ForStmt                NodeType = "ForStatement"
	FunctionDefinitionStmt NodeType = "FunctionDefinitionStatement"
	GlobalDeclarationStmt  NodeType = "GlobalDeclarationStatement"
	IfStmt                 NodeType = "IfStatement"
	ReturnStmt             NodeType = "ReturnStatement"
	ThrowStmt              NodeType = "ThrowStatement"
	TraitUseStmt           NodeType = "TraitUseStatement"
	WhileStmt              NodeType = "WhileStatement"
	// Class
	ClassConstDeclarationStmt NodeType = "ClassConstDeclarationStatement"
	ClassDeclarationStmt      NodeType = "ClassDeclarationStatement"
	MethodDefinitionStmt      NodeType = "MethodDefinitionStatement"
	PropertyDeclarationStmt   NodeType = "ClassPropertyDeclarationStatement"
)

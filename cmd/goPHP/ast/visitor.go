package ast

type Visitor interface {
	// Statements
	ProcessBreakStmt(stmt *BreakStatement, context any) (any, error)
	ProcessCompoundStmt(stmt *CompoundStatement, context any) (any, error)
	ProcessConstDeclarationStmt(stmt *ConstDeclarationStatement, context any) (any, error)
	ProcessContinueStmt(stmt *ContinueStatement, context any) (any, error)
	ProcessDoStmt(stmt *DoStatement, context any) (any, error)
	ProcessEchoStmt(stmt *EchoStatement, context any) (any, error)
	ProcessExpressionStmt(stmt *ExpressionStatement, context any) (any, error)
	ProcessForStmt(stmt *ForStatement, context any) (any, error)
	ProcessFunctionDefinitionStmt(stmt *FunctionDefinitionStatement, context any) (any, error)
	ProcessIfStmt(stmt *IfStatement, context any) (any, error)
	ProcessReturnStmt(stmt *ReturnStatement, context any) (any, error)
	ProcessStmt(stmt *Statement, context any) (any, error)
	ProcessWhileStmt(stmt *WhileStatement, context any) (any, error)

	// Expressions
	ProcessArrayLiteralExpr(stmt *ArrayLiteralExpression, context any) (any, error)
	ProcessBinaryOpExpr(stmt *BinaryOpExpression, context any) (any, error)
	ProcessCastExpr(stmt *CastExpression, context any) (any, error)
	ProcessCoalesceExpr(stmt *CoalesceExpression, context any) (any, error)
	ProcessCompoundAssignmentExpr(stmt *CompoundAssignmentExpression, context any) (any, error)
	ProcessConditionalExpr(stmt *ConditionalExpression, context any) (any, error)
	ProcessConstantAccessExpr(stmt *ConstantAccessExpression, context any) (any, error)
	ProcessEmptyIntrinsicExpr(stmt *EmptyIntrinsicExpression, context any) (any, error)
	ProcessEqualityExpr(stmt *EqualityExpression, context any) (any, error)
	ProcessEvalIntrinsicExpr(stmt *EvalIntrinsicExpression, context any) (any, error)
	ProcessExitIntrinsicExpr(stmt *ExitIntrinsicExpression, context any) (any, error)
	ProcessExpr(stmt *Expression, context any) (any, error)
	ProcessFloatingLiteralExpr(stmt *FloatingLiteralExpression, context any) (any, error)
	ProcessFunctionCallExpr(stmt *FunctionCallExpression, context any) (any, error)
	ProcessIncludeExpr(stmt *IncludeExpression, context any) (any, error)
	ProcessIncludeOnceExpr(stmt *IncludeOnceExpression, context any) (any, error)
	ProcessIntegerLiteralExpr(stmt *IntegerLiteralExpression, context any) (any, error)
	ProcessIssetIntrinsicExpr(stmt *IssetIntrinsicExpression, context any) (any, error)
	ProcessLogicalExpr(stmt *LogicalExpression, context any) (any, error)
	ProcessLogicalNotExpr(stmt *LogicalNotExpression, context any) (any, error)
	ProcessParenthesizedExpr(stmt *ParenthesizedExpression, context any) (any, error)
	ProcessPostfixIncExpr(stmt *PostfixIncExpression, context any) (any, error)
	ProcessPrefixIncExpr(stmt *PrefixIncExpression, context any) (any, error)
	ProcessPrintExpr(stmt *PrintExpression, context any) (any, error)
	ProcessRelationalExpr(stmt *RelationalExpression, context any) (any, error)
	ProcessRequireExpr(stmt *RequireExpression, context any) (any, error)
	ProcessRequireOnceExpr(stmt *RequireOnceExpression, context any) (any, error)
	ProcessSimpleAssignmentExpr(stmt *SimpleAssignmentExpression, context any) (any, error)
	ProcessSimpleVariableExpr(stmt *SimpleVariableExpression, context any) (any, error)
	ProcessStringLiteralExpr(stmt *StringLiteralExpression, context any) (any, error)
	ProcessSubscriptExpr(stmt *SubscriptExpression, context any) (any, error)
	ProcessTextExpr(stmt *TextExpression, context any) (any, error)
	ProcessUnaryExpr(stmt *UnaryOpExpression, context any) (any, error)
	ProcessUnsetIntrinsicExpr(stmt *UnsetIntrinsicExpression, context any) (any, error)
	ProcessVariableNameExpr(stmt *VariableNameExpression, context any) (any, error)
}

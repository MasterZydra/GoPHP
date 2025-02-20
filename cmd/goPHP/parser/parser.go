package parser

import (
	"GoPHP/cmd/goPHP/ast"
	"GoPHP/cmd/goPHP/common"
	"GoPHP/cmd/goPHP/ini"
	"GoPHP/cmd/goPHP/lexer"
	"GoPHP/cmd/goPHP/phpError"
	"GoPHP/cmd/goPHP/position"
	"GoPHP/cmd/goPHP/stats"
	"slices"
	"strings"
)

type Parser struct {
	ini     *ini.Ini
	program *ast.Program
	lexer   *lexer.Lexer
	tokens  []*lexer.Token
	currPos int
	id      int64
}

func NewParser(ini *ini.Ini) *Parser {
	return &Parser{ini: ini}
}

func (parser *Parser) init() {
	parser.program = ast.NewProgram()
	parser.lexer = lexer.NewLexer(parser.ini)
	parser.currPos = 0
}

func (parser *Parser) nextId() int64 {
	parser.id++
	return parser.id
}

func (parser *Parser) ProduceAST(sourceCode string, filename string) (*ast.Program, phpError.Error) {
	parser.init()

	var lexerErr error
	parser.tokens, lexerErr = parser.lexer.Tokenize(sourceCode, filename)
	if lexerErr != nil {
		return parser.program, phpError.NewParseError(lexerErr.Error())
	}

	stat := stats.Start()
	defer stats.StopAndPrint(stat, "Parser")

	PrintParserCallstack("Parser callstack", nil)
	PrintParserCallstack("----------------", nil)

	for !parser.isEof() {
		if parser.isTokenType(lexer.StartTagToken, true) || parser.isTokenType(lexer.EndTagToken, true) ||
			parser.isToken(lexer.OpOrPuncToken, ";", true) {
			continue
		}
		stmt, err := parser.parseStmt()
		if err != nil {
			return parser.program, err
		}
		parser.program.Append(stmt)
	}

	return parser.program, nil
}

func (parser *Parser) parseStmt() (ast.IStatement, phpError.Error) {
	// Spec: https://phplang.org/spec/11-statements.html#general

	// statement:
	//    compound-statement
	//    named-label-statement
	//    expression-statement
	//    selection-statement
	//    iteration-statement
	//    jump-statement
	//    try-statement
	//    declare-statement
	//    echo-statement
	//    unset-statement
	//    const-declaration
	//    function-definition
	//    class-declaration
	//    interface-declaration
	//    trait-declaration
	//    namespace-definition
	//    namespace-use-declaration
	//    global-declaration
	//    function-static-declaration

	// Resolve text expressions
	if parser.isTextExpression(true) {
		PrintParserCallstack("text-expression", parser)
		statements := []ast.IStatement{}
		textExpr := ast.NewExpressionStmt(parser.nextId(), ast.NewTextExpr(parser.nextId(), parser.eat().Value))
		parser.isTokenType(lexer.StartTagToken, true)

		statements = append(statements, textExpr)

		stmt, err := parser.parseStmt()
		for err == nil || parser.isTextExpression(false) {
			statements = append(statements, stmt)
			stmt, err = parser.parseStmt()
		}

		return ast.NewCompoundStmt(parser.nextId(), statements), nil
	}

	if parser.isTokenType(lexer.TextToken, false) {
		return ast.NewExpressionStmt(parser.nextId(), ast.NewTextExpr(parser.nextId(), parser.eat().Value)), nil
	}

	// ------------------- MARK: compound-statement -------------------

	// Spec: https://phplang.org/spec/11-statements.html#grammar-compound-statement

	// compound-statement:
	//    {   statement-list(opt)   }

	// statement-list:
	//    statement
	//    statement-list   statement

	if parser.isToken(lexer.OpOrPuncToken, "{", true) {
		PrintParserCallstack("compound-statement", parser)
		statements := []ast.IStatement{}
		for !parser.isEof() && !parser.isToken(lexer.OpOrPuncToken, "}", false) {
			stmt, err := parser.parseStmt()
			if err != nil {
				return ast.NewEmptyStmt(), err
			}
			statements = append(statements, stmt)
		}

		if !parser.isToken(lexer.OpOrPuncToken, "}", true) {
			return ast.NewEmptyStmt(), phpError.NewParseError("Expected \"}\", Got: %s", parser.at())
		}
		return ast.NewCompoundStmt(parser.nextId(), statements), nil
	}

	// TODO named-label-statement

	// selection-statement
	if parser.isToken(lexer.KeywordToken, "if", false) || parser.isToken(lexer.KeywordToken, "switch", false) {
		return parser.parseSelectionStmt()
	}

	// iteration-statement
	if parser.isToken(lexer.KeywordToken, "while", false) || parser.isToken(lexer.KeywordToken, "do", false) ||
		parser.isToken(lexer.KeywordToken, "for", false) || parser.isToken(lexer.KeywordToken, "foreach", false) {
		return parser.parseIterationStmt()
	}

	// jump-statement
	if parser.isToken(lexer.KeywordToken, "goto", false) || parser.isToken(lexer.KeywordToken, "continue", false) ||
		parser.isToken(lexer.KeywordToken, "break", false) || parser.isToken(lexer.KeywordToken, "return", false) ||
		parser.isToken(lexer.KeywordToken, "throw", false) {
		return parser.parseJumpStmt()
	}

	// TODO try-statement
	// TODO declare-statement

	// ------------------- MARK: echo-statement -------------------

	// Spec https://phplang.org/spec/11-statements.html#the-echo-statement

	// echo-statement:
	//    echo   expression-list   ;

	// expression-list:
	//    expression
	//    expression-list   ,   expression

	if parser.isToken(lexer.KeywordToken, "echo", false) {
		PrintParserCallstack("echo-statement", parser)
		pos := parser.eat().Position
		expressions := make([]ast.IExpression, 0)
		for {
			expr, err := parser.parseExpr()
			if err != nil {
				return ast.NewEmptyStmt(), err
			}

			expressions = append(expressions, expr)

			if parser.isToken(lexer.OpOrPuncToken, ",", true) {
				continue
			}
			if parser.isToken(lexer.OpOrPuncToken, ";", true) {
				break
			}
			return ast.NewEmptyStmt(), phpError.NewParseError("Invalid echo statement detected")
		}

		if len(expressions) == 0 {
			return ast.NewEmptyStmt(), phpError.NewParseError("Invalid echo statement detected")
		}

		return ast.NewEchoStmt(parser.nextId(), pos, expressions), nil
	}

	// ------------------- MARK: unset-statement -------------------

	// Spec: https://phplang.org/spec/11-statements.html#grammar-unset-statement

	// unset-statement:
	//    unset   (   variable-list   ,opt   )   ;

	// variable-list:
	//    variable
	//    variable-list   ,   variable

	if parser.isToken(lexer.KeywordToken, "unset", false) {
		PrintParserCallstack("unset-statement", parser)
		pos := parser.eat().Position
		if !parser.isToken(lexer.OpOrPuncToken, "(", true) {
			return ast.NewEmptyStmt(), phpError.NewParseError("Expected \"(\". Got: \"%s\"", parser.at())
		}
		args := []ast.IExpression{}
		for {
			if len(args) > 0 && parser.isToken(lexer.OpOrPuncToken, ")", true) {
				break
			}

			arg, err := parser.parseExpr()
			if err != nil {
				return ast.NewEmptyStmt(), err
			}
			if !ast.IsVariableExpr(arg) {
				return ast.NewEmptyStmt(), phpError.NewParseError("Fatal error: Cannot use unset() on the result of an expression")
			}
			args = append(args, arg)

			if parser.isToken(lexer.OpOrPuncToken, ",", true) ||
				parser.isToken(lexer.OpOrPuncToken, ")", false) {
				continue
			}
			return ast.NewEmptyStmt(), phpError.NewParseError("Expected \",\" or \")\". Got: %s", parser.at())
		}
		if !parser.isToken(lexer.OpOrPuncToken, ";", true) {
			return ast.NewEmptyStmt(), phpError.NewParseError("Expected \";\". Got: \"%s\"", parser.at())
		}
		return ast.NewExpressionStmt(parser.nextId(), ast.NewUnsetIntrinsic(parser.nextId(), pos, args)), nil
	}

	// ------------------- MARK: const-declaration -------------------

	// Spec: https://phplang.org/spec/14-classes.html#grammar-const-declaration

	// const-declaration:
	//    const   const-elements   ;

	// const-elements:
	//    const-element
	//    const-elements   ,   const-element

	// const-element:
	//    name   =   constant-expression

	if parser.isToken(lexer.KeywordToken, "const", false) {
		PrintParserCallstack("const-statement", parser)
		pos := parser.eat().Position
		if err := parser.expectTokenType(lexer.NameToken, false); err != nil {
			return ast.NewEmptyStmt(), err
		}
		for {
			name := parser.eat().Value
			if err := parser.expect(lexer.OpOrPuncToken, "=", true); err != nil {
				return ast.NewEmptyStmt(), err
			}
			// TODO parse constant-expression
			value, err := parser.parseExpr()
			if err != nil {
				return ast.NewEmptyStmt(), err
			}

			stmt := ast.NewConstDeclarationStmt(parser.nextId(), pos, name, value)
			if parser.isToken(lexer.OpOrPuncToken, ",", true) {
				parser.program.Append(stmt)
				continue
			}
			if parser.isToken(lexer.OpOrPuncToken, ";", true) {
				return stmt, nil
			}
			return ast.NewEmptyStmt(), phpError.NewParseError("Const declaration - unexpected token %s", parser.at())
		}
	}

	// function-definition
	if parser.isToken(lexer.KeywordToken, "function", false) {
		return parser.parseFunctionDefinition()
	}

	// TODO class-declaration
	// TODO interface-declaration
	// TODO trait-declaration
	// TODO namespace-definition
	// TODO namespace-use-declaration
	// TODO global-declaration
	// TODO function-static-declaration

	// ------------------- MARK: expression-statement -------------------

	// Spec: https://phplang.org/spec/11-statements.html#grammar-expression-statement

	// expression-statement:
	//    expression(opt)   ;

	// If present, expression is evaluated for its side effects, if any, and any resulting value is discarded.
	// If expression is omitted, the statement is a null statement, which has no effect on execution.
	parser.isToken(lexer.OpOrPuncToken, ";", true)

	PrintParserCallstack("expression-statement", parser)
	if expr, err := parser.parseExpr(); err != nil {
		return ast.NewEmptyExpr(), err
	} else {
		if parser.isToken(lexer.OpOrPuncToken, ";", true) {
			return ast.NewExpressionStmt(parser.nextId(), expr), nil
		}
		return ast.NewEmptyExpr(),
			phpError.NewParseError(`Statement must end with a semicolon. Got: "%s" at %s`, parser.at().Value, parser.at().Position.ToPosString())
	}
}

func (parser *Parser) parseSelectionStmt() (ast.IStatement, phpError.Error) {
	// ------------------- MARK: selection-statement -------------------

	// Spec: https://phplang.org/spec/11-statements.html#grammar-selection-statement

	// selection-statement:
	//    if-statement
	//    switch-statement

	if parser.isToken(lexer.KeywordToken, "if", false) {
		// Spec: https://phplang.org/spec/11-statements.html#grammar-if-statement

		// if-statement:
		//    if   (   expression   )   statement   elseif-clauses-1(opt)   else-clause-1(opt)
		//    if   (   expression   )   :   statement-list   elseif-clauses-2(opt)   else-clause-2(opt)   endif   ;

		// elseif-clauses-1:
		//    elseif-clause-1
		//    elseif-clauses-1   elseif-clause-1

		// elseif-clause-1:
		//    elseif   (   expression   )   statement

		// else-clause-1:
		//    else   statement

		// elseif-clauses-2:
		//    elseif-clause-2
		//    elseif-clauses-2   elseif-clause-2

		// elseif-clause-2:
		//    elseif   (   expression   )   :   statement-list

		// else-clause-2:
		//    else   :   statement-list

		PrintParserCallstack("if-statement", parser)

		// if
		ifPos := parser.eat().Position
		if !parser.isToken(lexer.OpOrPuncToken, "(", true) {
			return ast.NewEmptyStmt(), phpError.NewParseError("Expected \"(\". Got %s", parser.at())
		}

		condition, err := parser.parseExpr()
		if err != nil {
			return ast.NewEmptyStmt(), err
		}

		if !parser.isToken(lexer.OpOrPuncToken, ")", true) {
			return ast.NewEmptyStmt(), phpError.NewParseError("Expected \")\". Got %s", parser.at())
		}

		isAltSytax := parser.isToken(lexer.OpOrPuncToken, ":", true)

		var ifBlock ast.IStatement
		if !isAltSytax {
			ifBlock, err = parser.parseStmt()
			if err != nil {
				return ast.NewEmptyStmt(), err
			}
		} else {
			statements := []ast.IStatement{}
			for !parser.isToken(lexer.KeywordToken, "elseif", false) && !parser.isToken(lexer.KeywordToken, "else", false) &&
				!parser.isToken(lexer.KeywordToken, "endif", false) {
				statement, err := parser.parseStmt()
				if err != nil {
					return ast.NewEmptyStmt(), err
				}
				statements = append(statements, statement)
			}
			ifBlock = ast.NewCompoundStmt(parser.nextId(), statements)
		}

		// elseif
		elseIf := []*ast.IfStatement{}
		for parser.isToken(lexer.KeywordToken, "elseif", false) {
			elseIfPos := parser.eat().Position
			if !parser.isToken(lexer.OpOrPuncToken, "(", true) {
				return ast.NewEmptyStmt(), phpError.NewParseError("Expected \"(\". Got %s", parser.at())
			}

			elseIfCondition, err := parser.parseExpr()
			if err != nil {
				return ast.NewEmptyStmt(), err
			}

			if !parser.isToken(lexer.OpOrPuncToken, ")", true) {
				return ast.NewEmptyStmt(), phpError.NewParseError("Expected \")\". Got %s", parser.at())
			}

			if isAltSytax && !parser.isToken(lexer.OpOrPuncToken, ":", true) {
				return ast.NewEmptyStmt(), phpError.NewParseError("Expected \":\". Got %s", parser.at())
			}

			var elseIfBlock ast.IStatement
			if !isAltSytax {
				elseIfBlock, err = parser.parseStmt()
				if err != nil {
					return ast.NewEmptyStmt(), err
				}
			} else {
				statements := []ast.IStatement{}
				for !parser.isToken(lexer.KeywordToken, "elseif", false) && !parser.isToken(lexer.KeywordToken, "else", false) &&
					!parser.isToken(lexer.KeywordToken, "endif", false) {
					statement, err := parser.parseStmt()
					if err != nil {
						return ast.NewEmptyStmt(), err
					}
					statements = append(statements, statement)
				}
				elseIfBlock = ast.NewCompoundStmt(parser.nextId(), statements)
			}

			elseIf = append(elseIf, ast.NewIfStmt(parser.nextId(), elseIfPos, elseIfCondition, elseIfBlock, nil, nil))
		}

		// else
		var elseBlock ast.IStatement = nil
		if parser.isToken(lexer.KeywordToken, "else", true) {
			if isAltSytax && !parser.isToken(lexer.OpOrPuncToken, ":", true) {
				return ast.NewEmptyStmt(), phpError.NewParseError("Expected \":\". Got %s", parser.at())
			}

			if !isAltSytax {
				elseBlock, err = parser.parseStmt()
				if err != nil {
					return ast.NewEmptyStmt(), err
				}
			} else {
				statements := []ast.IStatement{}
				for !parser.isToken(lexer.KeywordToken, "endif", false) {
					statement, err := parser.parseStmt()
					if err != nil {
						return ast.NewEmptyStmt(), err
					}
					statements = append(statements, statement)
				}
				elseBlock = ast.NewCompoundStmt(parser.nextId(), statements)
			}
		}

		if isAltSytax && !parser.isToken(lexer.KeywordToken, "endif", true) {
			return ast.NewEmptyStmt(), phpError.NewParseError("Expected \"endif\". Got %s", parser.at())
		}
		if isAltSytax && !parser.isToken(lexer.OpOrPuncToken, ";", true) {
			return ast.NewEmptyStmt(), phpError.NewParseError("Expected \";\". Got %s", parser.at())
		}

		return ast.NewIfStmt(parser.nextId(), ifPos, condition, ifBlock, elseIf, elseBlock), nil
	}

	// TODO switch-statement
	// if parser.isToken(lexer.KeywordToken, "switch", false) {
	// }

	return ast.NewEmptyStmt(), phpError.NewParseError("Unsupported selection statement %s", parser.at())
}

func (parser *Parser) parseIterationStmt() (ast.IStatement, phpError.Error) {
	// ------------------- MARK: iteration-statement -------------------

	// Spec: https://phplang.org/spec/11-statements.html#grammar-iteration-statement

	// iteration-statement:
	//    while-statement
	//    do-statement
	//    for-statement
	//    foreach-statement

	if parser.isToken(lexer.KeywordToken, "while", false) {
		// Spec: https://phplang.org/spec/11-statements.html#grammar-iteration-statement

		// while-statement:
		//    while   (   expression   )   statement
		//    while   (   expression   )   :   statement-list   endwhile   ;

		PrintParserCallstack("while-statement", parser)

		// condition
		whilePos := parser.eat().Position
		if !parser.isToken(lexer.OpOrPuncToken, "(", true) {
			return ast.NewEmptyStmt(), phpError.NewParseError("Expected \"(\". Got %s", parser.at())
		}

		condition, err := parser.parseExpr()
		if err != nil {
			return ast.NewEmptyStmt(), err
		}

		if !parser.isToken(lexer.OpOrPuncToken, ")", true) {
			return ast.NewEmptyStmt(), phpError.NewParseError("Expected \")\". Got %s", parser.at())
		}

		isAltSytax := parser.isToken(lexer.OpOrPuncToken, ":", true)

		var block ast.IStatement
		if !isAltSytax {
			block, err = parser.parseStmt()
			if err != nil {
				return ast.NewEmptyStmt(), err
			}
		} else {
			statements := []ast.IStatement{}
			for !parser.isToken(lexer.KeywordToken, "endwhile", false) {
				statement, err := parser.parseStmt()
				if err != nil {
					return ast.NewEmptyStmt(), err
				}
				statements = append(statements, statement)
			}
			block = ast.NewCompoundStmt(parser.nextId(), statements)
		}

		if isAltSytax && !parser.isToken(lexer.KeywordToken, "endwhile", true) {
			return ast.NewEmptyStmt(), phpError.NewParseError("Expected \"endwhile\". Got %s", parser.at())
		}
		if isAltSytax && !parser.isToken(lexer.OpOrPuncToken, ";", true) {
			return ast.NewEmptyStmt(), phpError.NewParseError("Expected \";\". Got %s", parser.at())
		}

		return ast.NewWhileStmt(parser.nextId(), whilePos, condition, block), nil
	}

	if parser.isToken(lexer.KeywordToken, "do", false) {
		// Spec: https://phplang.org/spec/11-statements.html#grammar-do-statement

		// do-statement:
		//    do   statement   while   (   expression   )   ;

		PrintParserCallstack("do-statement", parser)

		doPos := parser.eat().Position

		// statement
		block, err := parser.parseStmt()
		if err != nil {
			return ast.NewEmptyStmt(), err
		}

		// condition
		if !parser.isToken(lexer.KeywordToken, "while", true) {
			return ast.NewEmptyStmt(), phpError.NewParseError("Expected \"while\". Got %s", parser.at())
		}

		if !parser.isToken(lexer.OpOrPuncToken, "(", true) {
			return ast.NewEmptyStmt(), phpError.NewParseError("Expected \"(\". Got %s", parser.at())
		}

		condition, err := parser.parseExpr()
		if err != nil {
			return ast.NewEmptyStmt(), err
		}

		if !parser.isToken(lexer.OpOrPuncToken, ")", true) {
			return ast.NewEmptyStmt(), phpError.NewParseError("Expected \")\". Got %s", parser.at())
		}

		if !parser.isToken(lexer.OpOrPuncToken, ";", true) {
			return ast.NewEmptyStmt(), phpError.NewParseError("Expected \";\". Got %s", parser.at())
		}

		return ast.NewDoStmt(parser.nextId(), doPos, condition, block), nil
	}

	if parser.isToken(lexer.KeywordToken, "for", false) {
		// Spec: https://phplang.org/spec/11-statements.html#grammar-for-statement

		// for-statement:
		//    for   (   for-initializer(opt)   ;   for-control(opt)   ;   for-end-of-loop(opt)   )   statement
		//    for   (   for-initializer(opt)   ;   for-control(opt)   ;   for-end-of-loop(opt)   )   :   statement-list   endfor   ;

		// for-initializer:
		//    for-expression-group

		// for-control:
		//    for-expression-group

		// for-end-of-loop:
		//    for-expression-group

		// for-expression-group:
		//    expression
		//    for-expression-group   ,   expression

		PrintParserCallstack("for-statement", parser)

		pos := parser.eat().Position

		if !parser.isToken(lexer.OpOrPuncToken, "(", true) {
			return ast.NewEmptyStmt(), phpError.NewParseError("Expected \"(\". Got %s", parser.at())
		}

		var err phpError.Error = nil
		parseExprGroup := func() (*ast.CompoundStatement, phpError.Error) {
			stmts := []ast.IStatement{}
			for {
				expr, err := parser.parseExpr()
				if err != nil {
					return nil, err
				}
				stmts = append(stmts, expr)

				if !parser.isToken(lexer.OpOrPuncToken, ",", true) {
					break
				}
			}
			return ast.NewCompoundStmt(parser.nextId(), stmts), nil
		}

		// for-initializer
		var forInitializer *ast.CompoundStatement = nil
		if !parser.isToken(lexer.OpOrPuncToken, ";", false) {
			forInitializer, err = parseExprGroup()
			if err != nil {
				return ast.NewEmptyExpr(), err
			}
		}

		if !parser.isToken(lexer.OpOrPuncToken, ";", true) {
			return ast.NewEmptyStmt(), phpError.NewParseError("Expected \";\". Got %s", parser.at())
		}

		// for-control
		var forControl *ast.CompoundStatement = nil
		if !parser.isToken(lexer.OpOrPuncToken, ";", false) {
			forControl, err = parseExprGroup()
			if err != nil {
				return ast.NewEmptyExpr(), err
			}
		}

		if !parser.isToken(lexer.OpOrPuncToken, ";", true) {
			return ast.NewEmptyStmt(), phpError.NewParseError("Expected \";\". Got %s", parser.at())
		}

		// for-end-of-loop
		var forEndOfLoop *ast.CompoundStatement = nil
		if !parser.isToken(lexer.OpOrPuncToken, ")", false) {
			forEndOfLoop, err = parseExprGroup()
			if err != nil {
				return ast.NewEmptyExpr(), err
			}
		}

		if !parser.isToken(lexer.OpOrPuncToken, ")", true) {
			return ast.NewEmptyStmt(), phpError.NewParseError("Expected \")\". Got %s", parser.at())
		}

		isAltSytax := parser.isToken(lexer.OpOrPuncToken, ":", true)

		var block ast.IStatement
		if !isAltSytax {
			block, err = parser.parseStmt()
			if err != nil {
				return ast.NewEmptyStmt(), err
			}
		} else {
			statements := []ast.IStatement{}
			for !parser.isToken(lexer.KeywordToken, "endfor", false) {
				statement, err := parser.parseStmt()
				if err != nil {
					return ast.NewEmptyStmt(), err
				}
				statements = append(statements, statement)
			}
			block = ast.NewCompoundStmt(parser.nextId(), statements)
		}

		if isAltSytax && !parser.isToken(lexer.KeywordToken, "endfor", true) {
			return ast.NewEmptyStmt(), phpError.NewParseError("Expected \"endfor\". Got %s", parser.at())
		}
		if isAltSytax && !parser.isToken(lexer.OpOrPuncToken, ";", true) {
			return ast.NewEmptyStmt(), phpError.NewParseError("Expected \";\". Got %s", parser.at())
		}

		return ast.NewForStmt(parser.nextId(), pos, forInitializer, forControl, forEndOfLoop, block), nil
	}

	// TODO foreach-statement

	return ast.NewEmptyStmt(), phpError.NewParseError("Unsupported iteration statement %s", parser.at())
}

func (parser *Parser) parseJumpStmt() (ast.IStatement, phpError.Error) {
	// ------------------- MARK: jump-statement -------------------

	// Spec: https://phplang.org/spec/11-statements.html#grammar-jump-statement

	// jump-statement:
	//    goto-statement
	//    continue-statement
	//    break-statement
	//    return-statement
	//    throw-statement

	// TODO goto-statement

	if parser.isToken(lexer.KeywordToken, "continue", false) {
		// Spec: https://phplang.org/spec/11-statements.html#grammar-continue-statement

		// continue-statement:
		//    continue   breakout-level(opt)   ;

		// breakout-level:
		//   integer-literal
		//   (   breakout-level   )

		PrintParserCallstack("continue-statement", parser)

		pos := parser.eat().Position

		var expr ast.IExpression = nil
		var err phpError.Error

		if !parser.isToken(lexer.OpOrPuncToken, ";", false) {
			isParenthesized := parser.isToken(lexer.OpOrPuncToken, "(", true)

			expr, err = parser.parseExpr()
			if err != nil {
				return ast.NewEmptyStmt(), err
			}

			if isParenthesized && !parser.isToken(lexer.OpOrPuncToken, ")", true) {
				return ast.NewEmptyExpr(), phpError.NewError("Expected closing parentheses. Got %s", parser.at())
			}
		}

		if !parser.isToken(lexer.OpOrPuncToken, ";", true) {
			return ast.NewEmptyStmt(), phpError.NewParseError("Expected \";\". Got %s", parser.at())
		}

		if expr == nil {
			expr = ast.NewIntegerLiteralExpr(parser.nextId(), nil, 1)
		}

		return ast.NewContinueStmt(parser.nextId(), pos, expr), nil
	}

	if parser.isToken(lexer.KeywordToken, "break", false) {
		// Spec: https://phplang.org/spec/11-statements.html#grammar-break-statement

		// break-statement:
		//    break   breakout-level(opt)   ;

		// breakout-level:
		//   integer-literal
		//   (   breakout-level   )

		PrintParserCallstack("break-statement", parser)

		pos := parser.eat().Position

		var expr ast.IExpression = nil
		var err phpError.Error

		if !parser.isToken(lexer.OpOrPuncToken, ";", false) {
			isParenthesized := parser.isToken(lexer.OpOrPuncToken, "(", true)

			expr, err = parser.parseExpr()
			if err != nil {
				return ast.NewEmptyStmt(), err
			}

			if isParenthesized && !parser.isToken(lexer.OpOrPuncToken, ")", true) {
				return ast.NewEmptyExpr(), phpError.NewError("Expected closing parentheses. Got %s", parser.at())
			}
		}

		if !parser.isToken(lexer.OpOrPuncToken, ";", true) {
			return ast.NewEmptyStmt(), phpError.NewParseError("Expected \";\". Got %s", parser.at())
		}

		if expr == nil {
			expr = ast.NewIntegerLiteralExpr(parser.nextId(), nil, 1)
		}

		return ast.NewBreakStmt(parser.nextId(), pos, expr), nil
	}

	// ------------------- MARK: return-statement -------------------

	// Spec: https://phplang.org/spec/11-statements.html#grammar-return-statement

	// return-statement:
	//    return   expressionopt   ;

	if parser.isToken(lexer.KeywordToken, "return", false) {
		PrintParserCallstack("return-statement", parser)
		pos := parser.eat().Position
		var expr ast.IExpression = nil
		if !parser.isToken(lexer.OpOrPuncToken, ";", false) {
			var err phpError.Error
			expr, err = parser.parseExpr()
			if err != nil {
				return ast.NewEmptyStmt(), err
			}
		}
		if !parser.isToken(lexer.OpOrPuncToken, ";", true) {
			return ast.NewEmptyStmt(), phpError.NewParseError("Expected: \";\". Got: \"%s\"", parser.at())
		}

		return ast.NewReturnStmt(parser.nextId(), pos, expr), nil
	}

	// TODO throw-statement

	return ast.NewEmptyStmt(), phpError.NewParseError("Unsupported jump statement %s", parser.at())
}

func (parser *Parser) parseFunctionDefinition() (ast.IStatement, phpError.Error) {
	// ------------------- MARK: function-definition -------------------
	// Spec: https://phplang.org/spec/13-functions.html#grammar-function-definition

	// function-definition:
	//    function-definition-header   compound-statement

	// function-definition-header:
	//    function   &(opt)   name   (   parameter-declaration-list(opt)   )   return-type(opt)

	// parameter-declaration-list:
	//    simple-parameter-declaration-list
	//    variadic-declaration-list

	// simple-parameter-declaration-list:
	//    parameter-declaration
	//    parameter-declaration-list   ,   parameter-declaration

	// variadic-declaration-list:
	//    simple-parameter-declaration-list   ,   variadic-parameter
	//    variadic-parameter

	// parameter-declaration:
	//    type-declaration(opt)   &(opt)   variable-name   default-argument-specifier(opt)

	// variadic-parameter:
	//    type-declaration(opt)   &(opt)   ...   variable-name

	// type-declaration:
	//    ?(opt)   base-type-declaration

	// return-type:
	//    :   type-declaration
	//    :   void

	// base-type-declaration:
	//    array
	//    callable
	//    iterable
	//    scalar-type
	//    qualified-name

	// scalar-type:
	//    bool
	//    float
	//    int
	//    string

	// default-argument-specifier:
	//    =   constant-expression

	PrintParserCallstack("function-definition", parser)
	if !parser.isToken(lexer.KeywordToken, "function", false) {
		return ast.NewEmptyStmt(), phpError.NewParseError("Expected \"function\". Got %s", parser.at())
	}

	pos := parser.eat().Position

	// TODO function-definition - &(opt)

	if parser.at().TokenType != lexer.NameToken {
		return ast.NewEmptyStmt(), phpError.NewParseError("Function name expected. Got %s", parser.at().TokenType)
	}

	functionName := parser.eat().Value

	if !parser.isToken(lexer.OpOrPuncToken, "(", true) {
		return ast.NewEmptyStmt(), phpError.NewParseError("Expected \"(\". Got %s", parser.at())
	}

	parameters := []ast.FunctionParameter{}
	if !parser.isToken(lexer.OpOrPuncToken, ")", false) {
		for {
			// Allow trailing comma
			if parser.isToken(lexer.OpOrPuncToken, ")", false) {
				break
			}

			// TODO function-definition - parameter-declaration - type-declaration - ?(opt)

			paramTypes := []string{}
			for parser.at().TokenType == lexer.KeywordToken && common.IsParamTypeKeyword(parser.at().Value) {
				paramTypes = append(paramTypes, strings.ToLower(parser.eat().Value))
				if parser.isToken(lexer.OpOrPuncToken, "|", true) {
					continue
				}
				break
			}

			if len(paramTypes) == 0 {
				paramTypes = append(paramTypes, "mixed")
			}

			// TODO function-definition - parameter-declaration - &(opt)

			if parser.at().TokenType != lexer.VariableNameToken {
				return ast.NewEmptyExpr(), phpError.NewParseError("Expected variable. Got %s", parser.at().TokenType)
			}
			parameters = append(parameters, ast.FunctionParameter{Name: parser.eat().Value, Type: paramTypes})

			if parser.isToken(lexer.OpOrPuncToken, ",", true) {
				continue
			}
			if parser.isToken(lexer.OpOrPuncToken, ")", false) {
				break
			}
			return ast.NewEmptyExpr(), phpError.NewParseError("Expected \",\" or \")\". Got %s", parser.at())
		}

		// TODO function-definition - parameter-declaration - default-argument-specifier(opt)

		// TODO function-definition - variadic-parameter
	}

	if !parser.isToken(lexer.OpOrPuncToken, ")", true) {
		return ast.NewEmptyStmt(), phpError.NewParseError("Expected \")\". Got %s", parser.at())
	}

	returnTypes := []string{}
	if parser.isToken(lexer.OpOrPuncToken, ":", true) {
		for parser.at().TokenType == lexer.KeywordToken && common.IsReturnTypeKeyword(parser.at().Value) {
			returnTypes = append(returnTypes, strings.ToLower(parser.eat().Value))
			if parser.isToken(lexer.OpOrPuncToken, "|", true) {
				continue
			}
			break
		}
	}
	if len(returnTypes) == 0 {
		returnTypes = append(returnTypes, "mixed")
	}

	body, err := parser.parseStmt()
	if err != nil {
		return ast.NewEmptyStmt(), err
	}
	if body.GetKind() != ast.CompoundStmt {
		return ast.NewEmptyStmt(), phpError.NewParseError("Expected compound statement. Got %s", body.GetKind())
	}

	return ast.NewFunctionDefinitionStmt(parser.nextId(), pos, functionName, parameters, body.(*ast.CompoundStatement), returnTypes), nil
}

func (parser *Parser) parseExpr() (ast.IExpression, phpError.Error) {
	// Spec: https://phplang.org/spec/10-expressions.html#grammar-expression

	// expression:
	//    logical-inc-OR-expression-2
	//    include-expression
	//    include-once-expression
	//    require-expression
	//    require-once-expression

	// ------------------- MARK: include-expression -------------------

	// Spec: https://phplang.org/spec/10-expressions.html#grammar-include-expression

	// include-expression:
	//    include   expression

	if parser.isToken(lexer.KeywordToken, "include", false) {
		PrintParserCallstack("include-expression", parser)
		pos := parser.eat().Position

		expr, err := parser.parseExpr()
		if err != nil {
			return ast.NewEmptyExpr(), err
		}

		return ast.NewIncludeExpr(parser.nextId(), pos, expr), nil
	}

	// ------------------- MARK: include-once-expression -------------------

	// Spec: https://phplang.org/spec/10-expressions.html#grammar-include-once-expression

	// include-once-expression:
	//    include_once   expression

	if parser.isToken(lexer.KeywordToken, "include_once", false) {
		PrintParserCallstack("include-once-expression", parser)
		pos := parser.eat().Position

		expr, err := parser.parseExpr()
		if err != nil {
			return ast.NewEmptyExpr(), err
		}

		return ast.NewIncludeOnceExpr(parser.nextId(), pos, expr), nil
	}

	// ------------------- MARK: require-expression -------------------

	// Spec: https://phplang.org/spec/10-expressions.html#grammar-require-expression

	// require-expression:
	//    require   expression

	if parser.isToken(lexer.KeywordToken, "require", false) {
		PrintParserCallstack("require-expression", parser)
		pos := parser.eat().Position

		expr, err := parser.parseExpr()
		if err != nil {
			return ast.NewEmptyExpr(), err
		}

		return ast.NewRequireExpr(parser.nextId(), pos, expr), nil
	}

	// ------------------- MARK: require-once-expression -------------------

	// Spec: https://phplang.org/spec/10-expressions.html#grammar-require-once-expression

	// require-once-expression:
	//    require_once   expression

	if parser.isToken(lexer.KeywordToken, "require_once", false) {
		PrintParserCallstack("require-once-expression", parser)
		pos := parser.eat().Position

		expr, err := parser.parseExpr()
		if err != nil {
			return ast.NewEmptyExpr(), err
		}

		return ast.NewRequireOnceExpr(parser.nextId(), pos, expr), nil
	}

	return parser.parseLogicalIncOrExpr2()
}

func (parser *Parser) parseLogicalIncOrExpr2() (ast.IExpression, phpError.Error) {
	// Spec: https://phplang.org/spec/10-expressions.html#grammar-logical-inc-OR-expression-2

	// logical-inc-OR-expression-2:
	//    logical-exc-OR-expression
	//    logical-inc-OR-expression-2   or   logical-exc-OR-expression

	lhs, err := parser.parseLogicalExcOrExpr()
	if err != nil {
		return ast.NewEmptyExpr(), err
	}

	for parser.isToken(lexer.KeywordToken, "or", true) {
		PrintParserCallstack("logical-inc-OR-expression-2", parser)
		rhs, err := parser.parseLogicalIncOrExpr2()
		if err != nil {
			return ast.NewEmptyExpr(), err
		}
		lhs = ast.NewLogicalExpr(parser.nextId(), lhs, "||", rhs)
	}
	return lhs, nil
}

func (parser *Parser) parseLogicalExcOrExpr() (ast.IExpression, phpError.Error) {
	// Spec: https://phplang.org/spec/10-expressions.html#grammar-logical-exc-OR-expression

	// logical-exc-OR-expression:
	//    logical-AND-expression-2
	//    logical-exc-OR-expression   xor   logical-AND-expression-2

	lhs, err := parser.parseLogicalAndExpr2()
	if err != nil {
		return ast.NewEmptyExpr(), err
	}

	for parser.isToken(lexer.KeywordToken, "xor", true) {
		PrintParserCallstack("logical-exc-OR-expression", parser)
		rhs, err := parser.parseLogicalExcOrExpr()
		if err != nil {
			return ast.NewEmptyExpr(), err
		}
		lhs = ast.NewLogicalExpr(parser.nextId(), lhs, "xor", rhs)
	}
	return lhs, nil
}

func (parser *Parser) parseLogicalAndExpr2() (ast.IExpression, phpError.Error) {
	// Spec: https://phplang.org/spec/10-expressions.html?#grammar-logical-AND-expression-2

	// logical-AND-expression-2:
	//    print-expression
	//    logical-AND-expression-2   and   yield-expression

	// print-expression

	// Spec: https://phplang.org/spec/10-expressions.html#grammar-print-expression

	// print-expression:
	//    yield-expression
	//    print   print-expression
	// Spec-Fix: So that by following assignment-expression the primary-expression is reachable
	//    assignment-expression

	var lhs ast.IExpression
	var err phpError.Error
	if parser.isToken(lexer.KeywordToken, "print", false) {
		PrintParserCallstack("print-expression", parser)
		pos := parser.eat().Position

		lhs, err = parser.parseLogicalAndExpr2()
		if err != nil {
			return ast.NewEmptyExpr(), err
		}
		lhs = ast.NewPrintExpr(parser.nextId(), pos, lhs)
	} else {
		lhs, err = parser.parseAssignmentExpr()
		if err != nil {
			return ast.NewEmptyExpr(), err
		}
	}

	for parser.isToken(lexer.KeywordToken, "and", true) {
		PrintParserCallstack("logical-AND-expression-2", parser)
		rhs, err := parser.parseLogicalAndExpr2()
		if err != nil {
			return ast.NewEmptyExpr(), err
		}
		lhs = ast.NewLogicalExpr(parser.nextId(), lhs, "&&", rhs)
	}
	return lhs, nil

	// TODO yield-expression
}

func (parser *Parser) parseAssignmentExpr() (ast.IExpression, phpError.Error) {
	// Spec: https://phplang.org/spec/10-expressions.html#grammar-assignment-expression

	// assignment-expression:
	//    conditional-expression
	//    simple-assignment-expression
	//    compound-assignment-expression

	// conditional-expression

	// Spec: https://phplang.org/spec/10-expressions.html#grammar-conditional-expression

	// conditional-expression:
	//    coalesce-expression
	//    conditional-expression   ?   expression(opt)   :   coalesce-expression

	// coalesce-expression
	expr, err := parser.parseCoalesceExpr()
	if err != nil {
		return ast.NewEmptyExpr(), err
	}

	// conditional-expression   ?   expression(opt)   :   coalesce-expression
	for parser.isToken(lexer.OpOrPuncToken, "?", true) {
		PrintParserCallstack("conditional-expression", parser)
		var ifExpr ast.IExpression = nil
		if !parser.isToken(lexer.OpOrPuncToken, ":", false) {
			ifExpr, err = parser.parseExpr()
			if err != nil {
				return ast.NewEmptyExpr(), err
			}
		}
		if err := parser.expect(lexer.OpOrPuncToken, ":", true); err != nil {
			return ast.NewEmptyExpr(), err
		}
		elseExpr, err := parser.parseExpr()
		if err != nil {
			return ast.NewEmptyExpr(), err
		}
		expr = ast.NewConditionalExpr(parser.nextId(), expr, ifExpr, elseExpr)
	}

	// Spec: https://phplang.org/spec/10-expressions.html#grammar-simple-assignment-expression

	// simple-assignment-expression:
	//    variable   =   assignment-expression
	//    list-intrinsic   =   assignment-expression

	// TODO simple-assignment-expression - list-intrinsic
	if ast.IsVariableExpr(expr) && parser.isToken(lexer.OpOrPuncToken, "=", true) {
		PrintParserCallstack("simple-assignment-expression", parser)
		value, err := parser.parseAssignmentExpr()
		if err != nil {
			return ast.NewEmptyExpr(), err
		}
		return ast.NewSimpleAssignmentExpr(parser.nextId(), expr, value), nil
	}

	// Spec: https://phplang.org/spec/10-expressions.html#grammar-compound-assignment-expression

	// compound-assignment-expression:
	//    variable   compound-assignment-operator   assignment-expression

	// compound-assignment-operator: one of
	//    **=   *=   /=   %=   +=   -=   .=   <<=   >>=   &=   ^=   |=

	if ast.IsVariableExpr(expr) &&
		parser.isTokenType(lexer.OpOrPuncToken, false) && common.IsCompoundAssignmentOp(parser.at().Value) {
		PrintParserCallstack("compound-assignment-operator", parser)
		operatorStr := strings.ReplaceAll(parser.eat().Value, "=", "")
		value, err := parser.parseAssignmentExpr()
		if err != nil {
			return ast.NewEmptyExpr(), nil
		}
		return ast.NewCompoundAssignmentExpr(parser.nextId(), expr, operatorStr, value), nil
	}

	return expr, nil
}

func (parser *Parser) parseCoalesceExpr() (ast.IExpression, phpError.Error) {
	// Spec: https://phplang.org/spec/10-expressions.html#grammar-coalesce-expression

	// coalesce-expression:
	//    logical-inc-OR-expression-1
	//    logical-inc-OR-expression-1   ??   coalesce-expression

	// logical-inc-OR-expression-1
	expr, err := parser.parseLogicalIncOrExpr1()
	if err != nil {
		return ast.NewEmptyExpr(), err
	}

	// logical-inc-OR-expression-1   ??   coalesce-expression
	if parser.isToken(lexer.OpOrPuncToken, "??", true) {
		PrintParserCallstack("coalesce-expression", parser)
		elseExpr, err := parser.parseCoalesceExpr()
		if err != nil {
			return ast.NewEmptyExpr(), err
		}

		return ast.NewCoalesceExpr(parser.nextId(), expr, elseExpr), nil
	}

	return expr, nil
}

func (parser *Parser) parseLogicalIncOrExpr1() (ast.IExpression, phpError.Error) {
	// Spec: https://phplang.org/spec/10-expressions.html#grammar-logical-inc-OR-expression-1

	// logical-inc-OR-expression-1:
	//    logical-AND-expression-1
	//    logical-inc-OR-expression-1   ||   logical-AND-expression-1

	lhs, err := parser.parseLogicalAndExpr1()
	if err != nil {
		return ast.NewEmptyExpr(), err
	}

	for parser.isToken(lexer.OpOrPuncToken, "||", true) {
		PrintParserCallstack("logical-inc-OR-expression-1", parser)
		rhs, err := parser.parseLogicalIncOrExpr1()
		if err != nil {
			return ast.NewEmptyExpr(), err
		}
		lhs = ast.NewLogicalExpr(parser.nextId(), lhs, "||", rhs)
	}
	return lhs, nil
}

func (parser *Parser) parseLogicalAndExpr1() (ast.IExpression, phpError.Error) {
	// Spec: https://phplang.org/spec/10-expressions.html#grammar-logical-AND-expression-1

	// logical-AND-expression-1:
	//    bitwise-inc-OR-expression
	//    logical-AND-expression-1   &&   bitwise-inc-OR-expression

	lhs, err := parser.parseBitwiseIncOrExpr()
	if err != nil {
		return ast.NewEmptyExpr(), err
	}

	for parser.isToken(lexer.OpOrPuncToken, "&&", true) {
		PrintParserCallstack("logical-AND-expression-1", parser)
		rhs, err := parser.parseLogicalAndExpr1()
		if err != nil {
			return ast.NewEmptyExpr(), err
		}
		lhs = ast.NewLogicalExpr(parser.nextId(), lhs, "&&", rhs)
	}
	return lhs, nil
}

func (parser *Parser) parseBitwiseIncOrExpr() (ast.IExpression, phpError.Error) {
	// Spec: https://phplang.org/spec/10-expressions.html#grammar-bitwise-inc-OR-expression

	// bitwise-inc-OR-expression:
	//    bitwise-exc-OR-expression
	//    bitwise-inc-OR-expression   |   bitwise-exc-OR-expression

	lhs, err := parser.parseBitwiseExcOrExpr()
	if err != nil {
		return ast.NewEmptyExpr(), err
	}

	for parser.isToken(lexer.OpOrPuncToken, "|", true) {
		PrintParserCallstack("bitwise-inc-OR-expression", parser)
		rhs, err := parser.parseBitwiseIncOrExpr()
		if err != nil {
			return ast.NewEmptyExpr(), err
		}
		lhs = ast.NewBinaryOpExpr(parser.nextId(), lhs, "|", rhs)
	}
	return lhs, nil
}

func (parser *Parser) parseBitwiseExcOrExpr() (ast.IExpression, phpError.Error) {
	// Spec: https://phplang.org/spec/10-expressions.html#grammar-bitwise-exc-OR-expression

	// bitwise-exc-OR-expression:
	//    bitwise-AND-expression
	//    bitwise-exc-OR-expression   ^   bitwise-AND-expression

	lhs, err := parser.parseBitwiseAndExpr()
	if err != nil {
		return ast.NewEmptyExpr(), err
	}

	for parser.isToken(lexer.OpOrPuncToken, "^", true) {
		PrintParserCallstack("bitwise-exc-OR-expression", parser)
		rhs, err := parser.parseBitwiseExcOrExpr()
		if err != nil {
			return ast.NewEmptyExpr(), err
		}
		lhs = ast.NewBinaryOpExpr(parser.nextId(), lhs, "^", rhs)
	}
	return lhs, nil
}

func (parser *Parser) parseBitwiseAndExpr() (ast.IExpression, phpError.Error) {
	// Spec: https://phplang.org/spec/10-expressions.html#grammar-bitwise-AND-expression

	// bitwise-AND-expression:
	//    equality-expression
	//    bitwise-AND-expression   &   equality-expression

	lhs, err := parser.parseEqualityExpr()
	if err != nil {
		return ast.NewEmptyExpr(), err
	}

	for parser.isToken(lexer.OpOrPuncToken, "&", true) {
		PrintParserCallstack("bitwise-AND-expression", parser)
		rhs, err := parser.parseBitwiseAndExpr()
		if err != nil {
			return ast.NewEmptyExpr(), err
		}
		lhs = ast.NewBinaryOpExpr(parser.nextId(), lhs, "&", rhs)
	}
	return lhs, nil
}

func (parser *Parser) parseEqualityExpr() (ast.IExpression, phpError.Error) {
	// Spec: https://phplang.org/spec/10-expressions.html#grammar-equality-expression

	// equality-expression:
	//    relational-expression
	//    equality-expression   ==   relational-expression
	//    equality-expression   !=   relational-expression
	//    equality-expression   <>   relational-expression
	//    equality-expression   ===   relational-expression
	//    equality-expression   !==   relational-expression

	lhs, err := parser.parserRelationalExpr()
	if err != nil {
		return ast.NewEmptyExpr(), err
	}

	for parser.isTokenType(lexer.OpOrPuncToken, false) && common.IsEqualityOp(parser.at().Value) {
		PrintParserCallstack("equality-expression", parser)
		operator := parser.eat().Value
		rhs, err := parser.parserRelationalExpr()
		if err != nil {
			return ast.NewEmptyExpr(), err
		}
		lhs = ast.NewEqualityExpr(parser.nextId(), lhs, operator, rhs)
	}
	return lhs, nil
}

func (parser *Parser) parserRelationalExpr() (ast.IExpression, phpError.Error) {
	// Spec: https://phplang.org/spec/10-expressions.html#grammar-relational-expression

	// relational-expression:
	//    shift-expression
	//    relational-expression   <   shift-expression
	//    relational-expression   >   shift-expression
	//    relational-expression   <=   shift-expression
	//    relational-expression   >=   shift-expression
	//    relational-expression   <=>   shift-expression

	lhs, err := parser.parseShiftExpr()
	if err != nil {
		return ast.NewEmptyExpr(), err
	}

	for parser.isTokenType(lexer.OpOrPuncToken, false) && common.IsRelationalExpressionOp(parser.at().Value) {
		PrintParserCallstack("relational-expression", parser)
		operator := parser.eat().Value
		rhs, err := parser.parseShiftExpr()
		if err != nil {
			return ast.NewEmptyExpr(), err
		}
		lhs = ast.NewRelationalExpr(parser.nextId(), lhs, operator, rhs)
	}
	return lhs, nil
}

func (parser *Parser) parseShiftExpr() (ast.IExpression, phpError.Error) {
	// Spec: https://phplang.org/spec/10-expressions.html#grammar-shift-expression

	// shift-expression:
	//    additive-expression
	//    shift-expression   <<   additive-expression
	//    shift-expression   >>   additive-expression

	lhs, err := parser.parseAdditiveExpr()
	if err != nil {
		return ast.NewEmptyExpr(), err
	}

	for parser.isTokenType(lexer.OpOrPuncToken, false) && slices.Contains([]string{"<<", ">>"}, parser.at().Value) {
		PrintParserCallstack("shift-expression", parser)
		operator := parser.eat().Value
		rhs, err := parser.parseShiftExpr()
		if err != nil {
			return ast.NewEmptyExpr(), err
		}
		lhs = ast.NewBinaryOpExpr(parser.nextId(), lhs, operator, rhs)
	}
	return lhs, nil
}

func (parser *Parser) parseAdditiveExpr() (ast.IExpression, phpError.Error) {
	// Spec: https://phplang.org/spec/10-expressions.html#grammar-additive-expression

	// additive-expression:
	//    multiplicative-expression
	//    additive-expression   +   multiplicative-expression
	//    additive-expression   -   multiplicative-expression
	//    additive-expression   .   multiplicative-expression

	lhs, err := parser.parseMultiplicativeExpr()
	if err != nil {
		return ast.NewEmptyExpr(), err
	}

	for parser.isTokenType(lexer.OpOrPuncToken, false) && common.IsAdditiveOp(parser.at().Value) {
		PrintParserCallstack("additive-expression", parser)
		operator := parser.eat().Value
		rhs, err := parser.parseMultiplicativeExpr()
		if err != nil {
			return ast.NewEmptyExpr(), err
		}
		lhs = ast.NewBinaryOpExpr(parser.nextId(), lhs, operator, rhs)
	}

	return lhs, nil
}

func (parser *Parser) parseMultiplicativeExpr() (ast.IExpression, phpError.Error) {
	// Spec: https://phplang.org/spec/10-expressions.html#grammar-multiplicative-expression

	// multiplicative-expression:
	//    logical-NOT-expression
	//    multiplicative-expression   *   logical-NOT-expression
	//    multiplicative-expression   /   logical-NOT-expression
	//    multiplicative-expression   %   logical-NOT-expression

	lhs, err := parser.parseLogicalNotExpr()
	if err != nil {
		return ast.NewEmptyExpr(), err
	}

	for parser.isTokenType(lexer.OpOrPuncToken, false) && common.IsMultiplicativeOp(parser.at().Value) {
		PrintParserCallstack("multiplicative-expression", parser)
		operator := parser.eat().Value
		rhs, err := parser.parseLogicalNotExpr()
		if err != nil {
			return ast.NewEmptyExpr(), err
		}
		lhs = ast.NewBinaryOpExpr(parser.nextId(), lhs, operator, rhs)
	}
	return lhs, nil
}

func (parser *Parser) parseLogicalNotExpr() (ast.IExpression, phpError.Error) {
	// Spec: https://phplang.org/spec/10-expressions.html#grammar-logical-NOT-expression

	// logical-NOT-expression:
	//    instanceof-expression
	//    !   instanceof-expression

	isNotExpression := parser.isToken(lexer.OpOrPuncToken, "!", false)
	var pos *position.Position = nil
	if isNotExpression {
		PrintParserCallstack("logical-NOT-expression", parser)
		pos = parser.eat().Position
	}

	expr, err := parser.parseInstanceofExpr()
	if err != nil {
		return ast.NewEmptyExpr(), err
	}

	if isNotExpression {
		return ast.NewLogicalNotExpr(parser.nextId(), pos, expr), nil
	}
	return expr, nil
}

func (parser *Parser) parseInstanceofExpr() (ast.IExpression, phpError.Error) {
	// Spec: https://phplang.org/spec/10-expressions.html#grammar-instanceof-expression

	// instanceof-expression:
	//    unary-expression
	//    instanceof-subject   instanceof   class-type-designator

	// instanceof-subject:
	//    instanceof-expression

	// TODO instanceof-expression
	return parser.parseUnaryExpr()
}

func (parser *Parser) parseUnaryExpr() (ast.IExpression, phpError.Error) {
	// Spec: https://phplang.org/spec/10-expressions.html#grammar-unary-expression

	// unary-expression:
	//    exponentiation-expression
	//    unary-op-expression
	//    error-control-expression
	//    cast-expression

	// These operators associate right-to-left.

	// unary-op-expression

	// Spec: https://phplang.org/spec/10-expressions.html#grammar-unary-op-expression

	// unary-op-expression:
	//    unary-operator   unary-expression

	// unary-operator: one of
	//    +   -   ~

	if parser.isTokenType(lexer.OpOrPuncToken, false) && common.IsUnaryOp(parser.at().Value) {
		PrintParserCallstack("unary-op-expression", parser)
		pos := parser.at().Position
		operator := parser.eat().Value
		expr, err := parser.parseUnaryExpr()
		if err != nil {
			return ast.NewEmptyExpr(), err
		}

		return ast.NewUnaryOpExpr(parser.nextId(), pos, operator, expr), nil
	}

	// TODO error-control-expression

	// cast-expression

	// Spec: https://phplang.org/spec/10-expressions.html#grammar-cast-expression

	// cast-expression:
	//    (   cast-type   )   unary-expression

	if parser.isToken(lexer.OpOrPuncToken, "(", false) &&
		parser.next(0).TokenType == lexer.KeywordToken && common.IsCastTypeKeyword(parser.next(0).Value) {
		PrintParserCallstack("cast-expression", parser)
		pos := parser.eat().Position
		castType := parser.eat().Value
		if !parser.isToken(lexer.OpOrPuncToken, ")", true) {
			return ast.NewEmptyExpr(), phpError.NewParseError("Expected \")\". Got %s", parser.at())
		}
		expr, err := parser.parseUnaryExpr()
		if err != nil {
			return ast.NewEmptyExpr(), err
		}
		return ast.NewCastExpr(parser.nextId(), pos, castType, expr), nil
	}

	return parser.parseExponentiationExpr()
}

func (parser *Parser) parseExponentiationExpr() (ast.IExpression, phpError.Error) {
	// Spec: https://phplang.org/spec/10-expressions.html#grammar-exponentiation-expression

	// exponentiation-expression:
	//    clone-expression
	//    clone-expression   **   exponentiation-expression

	lhs, err := parser.parseCloneExpr()
	if err != nil {
		return ast.NewEmptyExpr(), err
	}
	if parser.isToken(lexer.OpOrPuncToken, "**", true) {
		PrintParserCallstack("exponentiation-expression", parser)
		rhs, err := parser.parseExponentiationExpr()
		if err != nil {
			return ast.NewEmptyExpr(), err
		}
		return ast.NewBinaryOpExpr(parser.nextId(), lhs, "**", rhs), nil
	}
	return lhs, nil
}

func (parser *Parser) parseCloneExpr() (ast.IExpression, phpError.Error) {
	// Spec: https://phplang.org/spec/10-expressions.html#grammar-clone-expression

	// clone-expression:
	//	primary-expression
	//	clone   primary-expression

	// TODO clone-expression
	return parser.parsePrimaryExpr()
}

func (parser *Parser) parsePrimaryExpr() (ast.IExpression, phpError.Error) {
	// Spec: https://phplang.org/spec/10-expressions.html#grammar-primary-expression

	// primary-expression:
	//    variable
	//    class-constant-access-expression
	//    constant-access-expression
	//    literal
	//    array-creation-expression
	//    intrinsic
	//    anonymous-function-creation-expression
	//    object-creation-expression
	//    postfix-increment-expression
	//    postfix-decrement-expression
	//    prefix-increment-expression
	//    prefix-decrement-expression
	//    byref-assignment-expression
	//    shell-command-expression
	//    (   expression   )

	// ------------------- MARK: variable -------------------

	// Spec: https://phplang.org/spec/10-expressions.html#grammar-variable

	// variable:
	//    callable-variable
	//    scoped-property-access-expression
	//    member-access-expression

	var variable ast.IExpression

	// ------------------- MARK: callable-variable -------------------

	// Spec: https://phplang.org/spec/10-expressions.html#grammar-callable-variable

	// callable-variable:
	//    simple-variable
	//    subscript-expression
	//    member-call-expression
	//    scoped-call-expression
	//    function-call-expression

	// ------------------- MARK: simple-variable -------------------

	// Spec: https://phplang.org/spec/10-expressions.html#simple-variable

	// simple-variable:
	//    variable-name

	if parser.isTokenType(lexer.VariableNameToken, false) {
		variable = ast.NewSimpleVariableExpr(parser.nextId(), ast.NewVariableNameExpr(parser.nextId(), parser.at().Position, parser.eat().Value))
	}

	// Spec: https://phplang.org/spec/10-expressions.html#simple-variable

	// simple-variable:
	//    $   {   expression   }

	if parser.isToken(lexer.OpOrPuncToken, "$", false) &&
		parser.next(0).TokenType == lexer.OpOrPuncToken && parser.next(0).Value == "{" {
		PrintParserCallstack("simple-variable", parser)
		parser.eatN(2)
		// Get expression
		expr, err := parser.parseExpr()
		if err != nil {
			return ast.NewEmptyExpr(), err
		}

		if parser.at().Value == "}" {
			parser.eat()
			variable = ast.NewSimpleVariableExpr(parser.nextId(), expr)
		} else {
			return ast.NewEmptyExpr(), phpError.NewParseError("End of simple variable expression not detected at %s", parser.at().Position.ToPosString())
		}
	}

	// Spec: https://phplang.org/spec/10-expressions.html#simple-variable

	// simple-variable:
	//    $   simple-variable

	if parser.isToken(lexer.OpOrPuncToken, "$", true) {
		PrintParserCallstack("simple-variable", parser)
		if expr, err := parser.parsePrimaryExpr(); err != nil {
			return ast.NewEmptyExpr(), err
		} else {
			variable = ast.NewSimpleVariableExpr(parser.nextId(), expr)
		}
	}

	if ast.IsVariableExpr(variable) &&
		!parser.isToken(lexer.OpOrPuncToken, "[", false) && !parser.isToken(lexer.OpOrPuncToken, "{", false) &&
		!parser.isToken(lexer.OpOrPuncToken, "++", false) && !parser.isToken(lexer.OpOrPuncToken, "--", false) {
		return variable, nil
	}

	// ------------------- MARK: subscript-expression -------------------

	// Spec: https://phplang.org/spec/10-expressions.html#grammar-subscript-expression

	// subscript-expression:
	//    dereferencable-expression   [   expression(opt)   ]
	//    dereferencable-expression   {   expression   }   <b>[Deprecated form]</b>

	// dereferencable-expression:
	//    variable
	//    (   expression   )
	//    array-creation-expression
	//    string-literal

	// TODO subscript-expression - dereferencable-expression (expression, array-creation, string)
	// TODO allow nesting
	if ast.IsVariableExpr(variable) && parser.isToken(lexer.OpOrPuncToken, "[", false) {
		PrintParserCallstack("subscript-expression", parser)
		for ast.IsVariableExpr(variable) && parser.isToken(lexer.OpOrPuncToken, "[", true) {
			var err phpError.Error
			var index ast.IExpression
			if !parser.isToken(lexer.OpOrPuncToken, "]", false) {
				index, err = parser.parseExpr()
				if err != nil {
					return ast.NewEmptyExpr(), err
				}
			}
			if !parser.isToken(lexer.OpOrPuncToken, "]", true) {
				return ast.NewEmptyExpr(), phpError.NewParseError("Expected \"]\". Got: %s", parser.at())
			}
			variable = ast.NewSubscriptExpr(parser.nextId(), variable, index)
		}
		return variable, nil
	}

	// TODO member-call-expression
	// TODO scoped-call-expression

	// ------------------- MARK: function-call-expression -------------------

	// Spec: https://phplang.org/spec/10-expressions.html#grammar-function-call-expression

	// function-call-expression:
	//    qualified-name   (   argument-expression-list(opt)   )
	//    qualified-name   (   argument-expression-list   ,   )
	//    callable-expression   (   argument-expression-list(opt)   )
	//    callable-expression   (   argument-expression-list   ,   )

	// argument-expression-list:
	//    argument-expression
	//    argument-expression-list   ,   argument-expression

	// argument-expression:
	//    variadic-unpacking
	//    expression

	// variadic-unpacking:
	//    ...   expression

	if parser.isTokenType(lexer.NameToken, false) &&
		parser.next(0).TokenType == lexer.OpOrPuncToken && parser.next(0).Value == "(" {
		PrintParserCallstack("function-call-expression", parser)
		pos := parser.at().Position
		functionName := parser.eat().Value
		args := []ast.IExpression{}
		// Eat opening parentheses
		parser.eat()
		for {
			if parser.isToken(lexer.OpOrPuncToken, ")", true) {
				break
			}

			arg, err := parser.parseExpr()
			if err != nil {
				return ast.NewEmptyExpr(), err
			}
			args = append(args, arg)

			if parser.isToken(lexer.OpOrPuncToken, ",", true) || parser.isToken(lexer.OpOrPuncToken, ")", false) {
				continue
			}
			return ast.NewEmptyExpr(), phpError.NewParseError("Expected \",\" or \")\". Got: %s", parser.at())
		}
		return ast.NewFunctionCallExpr(parser.nextId(), pos, functionName, args), nil
	}

	// TODO scoped-property-access-expression
	// TODO member-access-expression

	// TODO class-constant-access-expression

	// ------------------- MARK: constant-access-expression -------------------

	// Spec: https://phplang.org/spec/10-expressions.html#grammar-constant-access-expression

	// constant-access-expression:
	//    qualified-name

	// A constant-access-expression evaluates to the value of the constant with name qualified-name.

	// Spec: https://phplang.org/spec/09-lexical-structure.html#grammar-qualified-name

	// qualified-name::
	//    namespace-name-as-a-prefix(opt)   name

	if parser.isTokenType(lexer.NameToken, false) ||
		(parser.isTokenType(lexer.KeywordToken, false) && common.IsCorePredefinedConstants(parser.at().Value)) {
		// TODO constant-access-expression - namespace-name-as-a-prefix
		// TODO constant-access-expression - check if name is a defined constant here or in interpreter
		PrintParserCallstack("constant-access-expression", parser)
		return ast.NewConstantAccessExpr(parser.nextId(), parser.at().Position, parser.eat().Value), nil
	}

	// literal
	if parser.isTokenType(lexer.IntegerLiteralToken, false) || parser.isTokenType(lexer.FloatingLiteralToken, false) ||
		parser.isTokenType(lexer.StringLiteralToken, false) {
		return parser.parseLiteral()
	}

	// array-creation-expression
	if (parser.isToken(lexer.KeywordToken, "array", false) &&
		parser.next(0).TokenType == lexer.OpOrPuncToken && parser.next(0).Value == "(") ||
		parser.isToken(lexer.OpOrPuncToken, "[", false) {
		return parser.parseArrayCreationExpr()
	}

	// intrinsic
	if parser.isToken(lexer.KeywordToken, "empty", false) || parser.isToken(lexer.KeywordToken, "eval", false) ||
		parser.isToken(lexer.KeywordToken, "exit", false) || parser.isToken(lexer.KeywordToken, "die", false) ||
		parser.isToken(lexer.KeywordToken, "isset", false) {
		return parser.parseIntrinsic()
	}

	// TODO anonymous-function-creation-expression
	// TODO object-creation-expression

	// ------------------- MARK: postfix-increment-expression -------------------

	// Spec: https://phplang.org/spec/10-expressions.html#grammar-postfix-increment-expression

	// postfix-increment-expression:
	//    variable   ++

	// Spec: https://phplang.org/spec/10-expressions.html#grammar-postfix-decrement-expression

	// postfix-decrement-expression:
	//    variable   --

	if ast.IsVariableExpr(variable) && (parser.isToken(lexer.OpOrPuncToken, "++", false) ||
		parser.isToken(lexer.OpOrPuncToken, "--", false)) {
		PrintParserCallstack("postfix-decrement-expression", parser)
		return ast.NewPostfixIncExpr(parser.nextId(), parser.at().Position, variable, parser.eat().Value), nil
	}

	// ------------------- MARK: prefix-increment-expression -------------------

	// Spec: https://phplang.org/spec/10-expressions.html#grammar-prefix-increment-expression

	// prefix-increment-expression:
	//    ++   variable

	// Spec: https://phplang.org/spec/10-expressions.html#grammar-prefix-decrement-expression
	// prefix-decrement-expression:
	//    --   variable

	if parser.isToken(lexer.OpOrPuncToken, "++", false) ||
		parser.isToken(lexer.OpOrPuncToken, "--", false) {
		PrintParserCallstack("prefix-decrement-expression", parser)
		pos := parser.at().Position
		operator := parser.eat().Value
		variable, err := parser.parsePrimaryExpr()
		if err != nil {
			return ast.NewEmptyExpr(), err
		}
		if !ast.IsVariableExpr(variable) {
			return ast.NewEmptyExpr(), phpError.NewParseError("Syntax error, unexpected %s", variable)
		}
		return ast.NewPrefixIncExpr(parser.nextId(), pos, variable, operator), nil
	}

	// TODO byref-assignment-expression
	// TODO shell-command-expression

	// ------------------- MARK: (   expression   ) -------------------

	if parser.isToken(lexer.OpOrPuncToken, "(", false) {
		PrintParserCallstack("parenthesized-expression", parser)
		pos := parser.eat().Position
		expr, err := parser.parseExpr()
		if err != nil {
			return ast.NewEmptyExpr(), err
		}
		if parser.isToken(lexer.OpOrPuncToken, ")", true) {
			return ast.NewParenthesizedExpr(parser.nextId(), pos, expr), nil
		} else {
			return ast.NewEmptyExpr(), phpError.NewParseError("Expected \")\". Got: %s", parser.at())
		}
	}

	// if parser.isToken(lexer.OpOrPuncToken, ";", false) &&
	// 	slices.Contains([]lexer.TokenType{lexer.StartTagToken, lexer.EndTagToken}, parser.next(0).TokenType) {
	// 	parser.eatN(2)
	// }

	// if parser.isTokenType(lexer.TextToken, false) {
	// 	return ast.NewExpressionStmt(parser.nextId(), ast.NewTextExpr(parser.nextId(), parser.eat().Value)), nil
	// }

	return ast.NewEmptyExpr(), phpError.NewParseError("Unsupported expression type: %s", parser.at())
}

func (parser *Parser) parseLiteral() (ast.IExpression, phpError.Error) {
	// ------------------- MARK: literal -------------------

	// Spec: https://phplang.org/spec/10-expressions.html#grammar-literal

	// literal:
	//    integer-literal
	//    floating-literal
	//    string-literal

	// A literal evaluates to its value, as specified in the lexical specification for literals.

	// integer-literal
	if parser.isTokenType(lexer.IntegerLiteralToken, false) {
		PrintParserCallstack("integer-literal", parser)
		intValue, err := common.IntegerLiteralToInt64(parser.at().Value)
		if err != nil {
			return ast.NewEmptyExpr(), phpError.NewParseError("Unsupported integer literal \"%s\"", parser.at().Value)
		}

		return ast.NewIntegerLiteralExpr(parser.nextId(), parser.eat().Position, intValue), nil
	}

	// floating-literal
	if parser.isTokenType(lexer.FloatingLiteralToken, false) {
		PrintParserCallstack("floating-literal", parser)
		if common.IsFloatingLiteral(parser.at().Value) {
			return ast.NewFloatingLiteralExpr(parser.nextId(), parser.at().Position, common.FloatingLiteralToFloat64(parser.eat().Value)), nil
		}

		return ast.NewEmptyExpr(), phpError.NewParseError("Unsupported floating literal \"%s\"", parser.at().Value)
	}

	// string-literal
	if parser.isTokenType(lexer.StringLiteralToken, false) {
		PrintParserCallstack("string-literal", parser)
		// single-quoted-string-literal
		if common.IsSingleQuotedStringLiteral(parser.at().Value) {
			return ast.NewStringLiteralExpr(
					parser.nextId(), parser.at().Position, common.SingleQuotedStringLiteralToString(parser.eat().Value), ast.SingleQuotedString),
				nil
		}

		// double-quoted-string-literal
		if common.IsDoubleQuotedStringLiteral(parser.at().Value) {
			return ast.NewStringLiteralExpr(
					parser.nextId(), parser.at().Position, common.DoubleQuotedStringLiteralToString(parser.eat().Value), ast.DoubleQuotedString),
				nil
		}

		// TODO heredoc-string-literal
		// TODO nowdoc-string-literal
	}

	return ast.NewEmptyExpr(), phpError.NewParseError("parseLiteral: Unsupported literal: %s", parser.at())
}

func (parser *Parser) parseArrayCreationExpr() (ast.IExpression, phpError.Error) {
	// ------------------- MARK: array-creation-expression -------------------

	// Spec: https://phplang.org/spec/10-expressions.html#grammar-array-creation-expression

	// array-creation-expression:
	//    array   (   array-initializer(opt)   )
	//    [   array-initializer(opt)   ]

	// array-initializer:
	//    array-initializer-list   ,(opt)

	// array-initializer-list:
	//    array-element-initializer
	//    array-element-initializer   ,   array-initializer-list

	// array-element-initializer:
	//    &(opt)   element-value
	//    element-key   =>   &(opt)   element-value

	// element-key:
	//    expression

	// element-value:
	//    expression

	PrintParserCallstack("array-creation-expression", parser)

	if !((parser.isToken(lexer.KeywordToken, "array", false) &&
		parser.next(0).TokenType == lexer.OpOrPuncToken && parser.next(0).Value == "(") ||
		parser.isToken(lexer.OpOrPuncToken, "[", false)) {
		return ast.NewEmptyExpr(), phpError.NewParseError("Unsupported array creation: %s", parser.at())
	}

	isShortSyntax := true
	var pos *position.Position
	if parser.isToken(lexer.KeywordToken, "array", false) &&
		parser.next(0).TokenType == lexer.OpOrPuncToken && parser.next(0).Value == "(" {
		pos = parser.eat().Position
		parser.eat()
		isShortSyntax = false
	} else {
		pos = parser.eat().Position
	}

	var index int64 = 0
	arrayExpr := ast.NewArrayLiteralExpr(parser.nextId(), pos)
	for {
		if (!isShortSyntax && parser.isToken(lexer.OpOrPuncToken, ")", true)) ||
			(isShortSyntax && parser.isToken(lexer.OpOrPuncToken, "]", true)) {
			break
		}

		// TODO array-creation-expression - "key => value"
		element, err := parser.parseExpr()
		if err != nil {
			return ast.NewEmptyExpr(), err
		}
		arrayExpr.AddElement(ast.NewIntegerLiteralExpr(parser.nextId(), element.GetPosition(), index), element)
		index++

		if parser.isToken(lexer.OpOrPuncToken, ",", true) ||
			(!isShortSyntax && parser.isToken(lexer.OpOrPuncToken, ")", false)) ||
			(isShortSyntax && parser.isToken(lexer.OpOrPuncToken, "]", false)) {
			continue
		}
		if isShortSyntax {
			return ast.NewEmptyExpr(), phpError.NewParseError("Expected \",\" or \"]\". Got: %s", parser.at())
		} else {
			return ast.NewEmptyExpr(), phpError.NewParseError("Expected \",\" or \")\". Got: %s", parser.at())
		}
	}
	return arrayExpr, nil
}

func (parser *Parser) parseIntrinsic() (ast.IExpression, phpError.Error) {
	// ------------------- MARK: intrinsic -------------------

	// Spec: https://phplang.org/spec/10-expressions.html#grammar-intrinsic

	// intrinsic:
	//    empty-intrinsic
	//    eval-intrinsic
	//    exit-intrinsic
	//    isset-intrinsic

	// ------------------- MARK: empty-intrinsic -------------------

	// Spec: https://phplang.org/spec/10-expressions.html#grammar-empty-intrinsic

	// empty-intrinsic:
	//    empty   (   expression   )

	if parser.isToken(lexer.KeywordToken, "empty", false) {
		PrintParserCallstack("empty-intrinsic", parser)
		pos := parser.eat().Position
		if !parser.isToken(lexer.OpOrPuncToken, "(", true) {
			return ast.NewEmptyExpr(), phpError.NewParseError("Expected \"(\". Got: \"%s\"", parser.at())
		}
		expr, err := parser.parseExpr()
		if err != nil {
			return ast.NewEmptyExpr(), err
		}
		if !parser.isToken(lexer.OpOrPuncToken, ")", true) {
			return ast.NewEmptyExpr(), phpError.NewParseError("Expected \")\". Got: \"%s\"", parser.at())
		}
		return ast.NewEmptyIntrinsic(parser.nextId(), pos, expr), nil
	}

	// TODO eval-intrinsic

	// ------------------- MARK: exit-intrinsic -------------------

	// Spec: https://phplang.org/spec/10-expressions.html#grammar-exit-intrinsic

	// exit-intrinsic:
	//    exit
	//    exit   (   expression(opt)   )
	//    die
	//    die   (   expression(opt)   )

	if parser.isToken(lexer.KeywordToken, "exit", false) || parser.isToken(lexer.KeywordToken, "die", false) {
		PrintParserCallstack("exit-intrinsic", parser)
		pos := parser.eat().Position
		var expr ast.IExpression = nil
		if parser.isToken(lexer.OpOrPuncToken, "(", true) {
			if !parser.isToken(lexer.OpOrPuncToken, ")", false) {
				var err phpError.Error
				expr, err = parser.parseExpr()
				if err != nil {
					return ast.NewEmptyExpr(), err
				}
			}
			if !parser.isToken(lexer.OpOrPuncToken, ")", true) {
				return ast.NewEmptyExpr(), phpError.NewParseError("Expected \")\". Got %s", parser.at())
			}
		}
		return ast.NewExitIntrinsic(parser.nextId(), pos, expr), nil
	}

	// ------------------- MARK: isset-intrinsic -------------------

	// Spec: https://phplang.org/spec/10-expressions.html#grammar-isset-intrinsic

	// isset-intrinsic:
	//    isset   (   variable-list   ,(opt)   )

	// variable-list:
	//    variable
	//    variable-list   ,   variable

	if parser.isToken(lexer.KeywordToken, "isset", false) {
		PrintParserCallstack("isset-intrinsic", parser)
		pos := parser.eat().Position
		if !parser.isToken(lexer.OpOrPuncToken, "(", true) {
			return ast.NewEmptyExpr(), phpError.NewParseError("Expected \"(\". Got: \"%s\"", parser.at())
		}
		args := []ast.IExpression{}
		for {
			if len(args) > 0 && parser.isToken(lexer.OpOrPuncToken, ")", true) {
				break
			}

			arg, err := parser.parseExpr()
			if err != nil {
				return ast.NewEmptyExpr(), err
			}
			if !ast.IsVariableExpr(arg) {
				return ast.NewEmptyExpr(), phpError.NewParseError("Fatal error: Cannot use isset() on the result of an expression")
			}
			args = append(args, arg)

			if parser.isToken(lexer.OpOrPuncToken, ",", true) ||
				parser.isToken(lexer.OpOrPuncToken, ")", false) {
				continue
			}
			return ast.NewEmptyExpr(), phpError.NewParseError("Expected \",\" or \")\". Got: %s", parser.at())
		}
		return ast.NewIssetIntrinsic(parser.nextId(), pos, args), nil
	}

	return ast.NewEmptyExpr(), phpError.NewParseError("Unsupported intrinsic: %s", parser.at())
}

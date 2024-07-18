package interpreter

import (
	"GoPHP/cmd/goPHP/ast"
	"GoPHP/cmd/goPHP/ini"
	"GoPHP/cmd/goPHP/parser"
	"GoPHP/cmd/goPHP/phpError"
)

type Interpreter struct {
	filename      string
	includedFiles []string
	ini           *ini.Ini
	request       *Request
	parser        *parser.Parser
	env           *Environment
	cache         map[int64]IRuntimeValue
	result        string
	exitCode      int64
}

func NewInterpreter(ini *ini.Ini, request *Request, filename string) *Interpreter {
	return &Interpreter{
		filename: filename, includedFiles: []string{}, ini: ini, request: request, parser: parser.NewParser(ini),
		env: NewEnvironment(nil, request), cache: map[int64]IRuntimeValue{},
		exitCode: 0,
	}
}

func (interpreter *Interpreter) GetExitCode() int {
	return int(interpreter.exitCode)
}

func (interpreter *Interpreter) Process(sourceCode string) (string, phpError.Error) {
	return interpreter.process(sourceCode, interpreter.env)
}

func (interpreter *Interpreter) process(sourceCode string, env *Environment) (string, phpError.Error) {
	interpreter.result = ""
	program, parserErr := interpreter.parser.ProduceAST(sourceCode, interpreter.filename)
	if parserErr != nil {
		return interpreter.result, parserErr
	}

	_, err := interpreter.processProgram(program, env)

	return interpreter.result, err
}

func (interpreter *Interpreter) processProgram(program *ast.Program, env *Environment) (IRuntimeValue, phpError.Error) {
	err := interpreter.scanForFunctionDefinition(program.GetStatements(), env)
	if err != nil {
		return NewVoidRuntimeValue(), err
	}

	var runtimeValue IRuntimeValue = NewVoidRuntimeValue()
	for _, stmt := range program.GetStatements() {
		if runtimeValue, err = interpreter.processStmt(stmt, env); err != nil {
			// Handle exit event - Stop code execution
			if err.GetErrorType() == phpError.EventError && err.GetMessage() == phpError.ExitEvent {
				break
			}
			return runtimeValue, err
		}
	}
	return runtimeValue, nil
}

func (interpreter *Interpreter) processStmt(stmt ast.IStatement, env any) (IRuntimeValue, phpError.Error) {
	runtimeValue, err := stmt.Process(interpreter, env)
	var phpErr phpError.Error = nil
	if err != nil {
		phpErr = err.(phpError.Error)
	}
	return runtimeValue.(IRuntimeValue), phpErr
}

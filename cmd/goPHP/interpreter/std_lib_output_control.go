package interpreter

import (
	"GoPHP/cmd/goPHP/phpError"
	"GoPHP/cmd/goPHP/runtime"
	"GoPHP/cmd/goPHP/runtime/values"
)

func registerNativeOutputControlFunctions(environment *Environment) {
	environment.nativeFunctions["ob_clean"] = nativeFn_ob_clean
	environment.nativeFunctions["ob_flush"] = nativeFn_ob_flush
	environment.nativeFunctions["ob_end_clean"] = nativeFn_ob_end_clean
	environment.nativeFunctions["ob_end_flush"] = nativeFn_ob_end_flush
	environment.nativeFunctions["ob_get_clean"] = nativeFn_ob_get_clean
	environment.nativeFunctions["ob_get_flush"] = nativeFn_ob_get_flush
	environment.nativeFunctions["ob_get_contents"] = nativeFn_ob_get_contents
	environment.nativeFunctions["ob_get_level"] = nativeFn_ob_get_level
	environment.nativeFunctions["ob_start"] = nativeFn_ob_start
}

// ------------------- MARK: ob_clean -------------------

func nativeFn_ob_clean(args []values.RuntimeValue, context runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.ob-clean.php

	_, err := NewFuncParamValidator("ob_clean").validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	// TODO Call output handler
	// Spec: https://www.php.net/manual/en/function.ob-clean.php
	// This function calls the output handler (with the PHP_OUTPUT_HANDLER_CLEAN flag),
	// discards it's return value and cleans (erases) the contents of the active output buffer.

	// TODO Throw notice if no buffer: e.g. Notice: ob_clean(): Failed to delete buffer. No buffer to delete in /home/user/scripts/code.php on line 5

	if context.Interpreter.GetOutputBufferStack().Len() == 0 {
		return values.NewBool(false), nil
	}

	context.Interpreter.GetOutputBufferStack().GetLast().Content = ""
	return values.NewBool(true), nil
}

// ------------------- MARK: ob_flush -------------------

func nativeFn_ob_flush(args []values.RuntimeValue, context runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.ob-flush.php

	_, err := NewFuncParamValidator("ob_flush").validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	// TODO Call output handler
	// Spec: https://www.php.net/manual/en/function.ob-flush.php
	// This function calls the output handler (with the PHP_OUTPUT_HANDLER_FLUSH flag),
	// discards it's return value and flushs (erases) the contents of the active output buffer.

	// TODO Throw notice if no buffer: e.g. Notice: ob_flush(): Failed to delete buffer. No buffer to delete in /home/user/scripts/code.php on line 5

	if context.Interpreter.GetOutputBufferStack().Len() == 0 {
		return values.NewBool(false), nil
	}

	if context.Interpreter.GetOutputBufferStack().Len() == 1 {
		context.Interpreter.WriteResult(context.Interpreter.GetOutputBufferStack().GetLast().Content)
	} else {
		context.Interpreter.GetOutputBufferStack().Get(context.Interpreter.GetOutputBufferStack().Len() - 2).Content += context.Interpreter.GetOutputBufferStack().GetLast().Content
	}

	nativeFn_ob_clean(args, context)
	return values.NewBool(true), nil
}

// ------------------- MARK: ob_end_clean -------------------

func nativeFn_ob_end_clean(args []values.RuntimeValue, context runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.ob-end-clean.php

	_, err := NewFuncParamValidator("ob_end_clean").validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	// TODO Call output handler
	// Spec: https://www.php.net/manual/en/function.ob-end-clean.php
	// This function calls the output handler (with the PHP_OUTPUT_HANDLER_CLEAN and PHP_OUTPUT_HANDLER_FINAL flags),
	// discards it's return value, discards the contents of the active output buffer and turns off the active output buffer.

	// TODO Throw notice if no buffer: e.g. Notice: ob_end_clean(): Failed to delete buffer. No buffer to delete in /home/user/scripts/code.php on line 5

	if context.Interpreter.GetOutputBufferStack().Len() == 0 {
		return values.NewBool(false), nil
	}

	context.Interpreter.GetOutputBufferStack().Pop()
	return values.NewBool(true), nil
}

// ------------------- MARK: ob_end_flush -------------------

func nativeFn_ob_end_flush(args []values.RuntimeValue, context runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.ob-end-flush.php

	_, err := NewFuncParamValidator("ob_end_flush").validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	// TODO Call output handler
	// Spec: https://www.php.net/manual/en/function.ob-end-flush.php
	// This function calls the output handler (with the PHP_OUTPUT_HANDLER_FINAL flag),
	// flushes (sends) it's return value, discards the contents of the active output buffer and turns off the active output buffer.

	// TODO Throw notice if no buffer: e.g. Notice: ob_end_flush(): Failed to delete buffer. No buffer to delete in /home/user/scripts/code.php on line 5

	if context.Interpreter.GetOutputBufferStack().Len() == 0 {
		return values.NewBool(false), nil
	}

	nativeFn_ob_flush(args, context)
	context.Interpreter.GetOutputBufferStack().Pop()
	return values.NewBool(true), nil
}

// ------------------- MARK: ob_get_clean -------------------

func nativeFn_ob_get_clean(args []values.RuntimeValue, context runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.ob-get-clean.php

	_, err := NewFuncParamValidator("ob_get_clean").validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	// TODO Call output handler
	// Spec: https://www.php.net/manual/en/function.ob-get-clean.php
	// This function calls the output handler (with the PHP_OUTPUT_HANDLER_CLEAN and PHP_OUTPUT_HANDLER_FINAL flags),
	// discards it's return value, returns the contents of the active output buffer and turns off the active output buffer.

	// TODO Throw notice if no buffer: e.g. Notice: ob_get_clean(): Failed to delete buffer. No buffer to delete in /home/user/scripts/code.php on line 5

	if context.Interpreter.GetOutputBufferStack().Len() == 0 {
		return values.NewBool(false), nil
	}

	content := context.Interpreter.GetOutputBufferStack().GetLast().Content
	context.Interpreter.GetOutputBufferStack().Pop()
	return values.NewStr(content), nil
}

// ------------------- MARK: ob_get_flush -------------------

func nativeFn_ob_get_flush(args []values.RuntimeValue, context runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.ob-get-flush.php

	_, err := NewFuncParamValidator("ob_get_flush").validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	// TODO Call output handler
	// Spec: https://www.php.net/manual/en/function.ob-get-flush.php
	// This function calls the output handler (with the PHP_OUTPUT_HANDLER_FINAL flag),
	// flushes (sends) it's return value, returns the contents of the active output buffer and turns off the active output buffer.

	// TODO Throw notice if no buffer: e.g. Notice: ob_get_flush(): Failed to delete buffer. No buffer to delete in /home/user/scripts/code.php on line 5

	if context.Interpreter.GetOutputBufferStack().Len() == 0 {
		return values.NewBool(false), nil
	}

	content := context.Interpreter.GetOutputBufferStack().GetLast().Content
	nativeFn_ob_end_flush(args, context)
	return values.NewStr(content), nil
}

// ------------------- MARK: ob_get_contents -------------------

func nativeFn_ob_get_contents(args []values.RuntimeValue, context runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.ob-get-contents.php

	_, err := NewFuncParamValidator("nativeFn_ob_get_contents").validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	if context.Interpreter.GetOutputBufferStack().Len() == 0 {
		return values.NewBool(false), nil
	}

	return values.NewStr(context.Interpreter.GetOutputBufferStack().GetLast().Content), nil
}

// ------------------- MARK: ob_get_level -------------------

func nativeFn_ob_get_level(args []values.RuntimeValue, context runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.ob-get-level.php

	_, err := NewFuncParamValidator("ob_get_level").validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	return values.NewInt(int64(context.Interpreter.GetOutputBufferStack().Len())), nil
}

// ------------------- MARK: ob_start -------------------

func nativeFn_ob_start(args []values.RuntimeValue, context runtime.Context) (values.RuntimeValue, phpError.Error) {
	// Spec: https://www.php.net/manual/en/function.ob-start

	_, err := NewFuncParamValidator("ob_start").validate(args)
	if err != nil {
		return values.NewVoid(), err
	}

	// TODO ob_start parameters
	//  ob_start(?callable $callback = null, int $chunk_size = 0, int $flags = PHP_OUTPUT_HANDLER_STDFLAGS): bool

	context.Interpreter.GetOutputBufferStack().Push()

	return values.NewBool(true), nil
}

// TODO flush
// TODO ob_​get_​length
// TODO ob_​get_​status
// TODO ob_​implicit_​flush
// TODO ob_​list_​handlers
// TODO output_​add_​rewrite_​var
// TODO output_​reset_​rewrite_​vars

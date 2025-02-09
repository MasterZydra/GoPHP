# Supported syntax and features

## Ini directives
- error_reporting
- register_argc_argv
- short_open_tag
- variables_order

## Expressions and statements
- echo statement: `echo "abc", 123, true;`
- print statement: `print "abc";`
- short open tag: `<? 1 + 2;`
- short echo statement: `<?= "123";`
- declare and access variable: `$var = "abc"; echo $abc;`
- declare and access constant: `const PI = 3.141; echo PI;`
- simple assignment: `$var = 42;`
- compound assignment: `$var += 42;`
- cast expression: `(int)$a;(string)$a;`
- conditional expression: `$var ? $a : "b";`
- coalesce expression: `$var ?? "b";`
- equality expression: `$var === 42;`
- relational expression: `$var >= 42;`
- additive expression: `$var + 42; $var - 42; "a" . "b";`
- multiplicative expression: `$var * 42; $var / 42; $var % 42;`
- logical and expression: `$var && 8;`
- logical and expression 2: `$var and 8;`
- logical exc or expression: `$var xor 8;`
- logical inc or expression: `$var || 8;`
- logical inc or expression 2: `$var or 8;`
- bitwise exc or expression: `$var ^ 8;`
- bitwise inc or expression: `$var | 8;`
- bitwise and expression: `$var & 8;`
- shift expression: `$var << 8;`
- exponentiation expression: `$var ** 42;`
- unary expression: `-1; +1; ~1;`
- prefix (in/de)crease expression: `++$var; --$var;`
- postfix (in/de)crease expression: `$var++; $var--;`
- logical not expression: `!$var;`
- parenthesized expression: `(1 + 2) * 3;`
- subscript expression: `$a[1];`
- variable substitution: `echo "{$a}";`
- if statement: `if (true) { ... } elseif (false) { ... } else { ... }`
- for statement: `for (...; ...; ...) { ... }`
- while statement: `while (true) { ... }`
- do statement: `do { ... } while (true);`
- function definition: `function func1($param1) { ... }`
- break statement: `break 1;`
- continue statement: `continue (2);`
- return statement: `return 42;`
- require(_once), include(_once): `require 'lib.php';`

## Data types
- array
- bool
- float (including numeric literal separator)
- int  (including numeric literal separator)
- null
- string

## Predefined variables
- $_ENV
- $_GET
- $_POST
- $_SERVER

## Predefined constants
- DIRECTORY_SEPARATOR
- E_ALL
- E_COMPILE_ERROR
- E_COMPILE_WARNING
- E_CORE_ERROR
- E_CORE_WARNING
- E_DEPRECATED
- E_ERROR
- E_NOTICE
- E_PARSE
- E_RECOVERABLE_ERROR
- E_STRICT
- E_USER_DEPRECATED
- E_USER_ERROR
- E_USER_NOTICE
- E_USER_WARNING
- E_WARNING
- FALSE
- M_1_PI
- M_2_PI
- M_2_SQRTPI
- M_E
- M_EULER
- M_LN10
- M_LN2
- M_LNPI
- M_LOG10E
- M_LOG2E
- M_PI
- M_PI_2
- M_PI_4
- M_SQRT1_2
- M_SQRT2
- M_SQRT3
- M_SQRTPI
- NULL
- PHP_EOL
- PHP_EXTRA_VERSION
- PHP_INT_MAX
- PHP_INT_MIN
- PHP_INT_SIZE
- PHP_MAJOR_VERSION
- PHP_MINOR_VERSION
- PHP_OS
- PHP_OS_FAMILY
- PHP_RELEASE_VERSION
- PHP_ROUND_HALF_DOWN
- PHP_ROUND_HALF_EVEN
- PHP_ROUND_HALF_ODD
- PHP_ROUND_HALF_UP
- PHP_VERSION
- PHP_VERSION_ID
- TRUE

## Magic constants
- \_\_DIR\_\_
- \_\_FILE\_\_
- \_\_FUNCTION\_\_
- \_\_LINE\_\_

## Standard library
- array_key_exists
- die
- error_reporting
- exit
- getenv
- ini_get
- key_exists

### Date / Time
- checkdate
- date
- getdate
- localtime
- microtime
- mktime
- time

### Math
- abs
- acos
- acosh
- asin
- asinh
- pi

### Misc
- constant
- defined

### Strings
- bin2hex
- chr
- lcfirst
- quotemeta
- str_contains
- str_ends_with
- str_repeat
- str_starts_with
- strlen
- strtolower
- strtoupper
- ucfirst

### Variable handling functions
- boolval
- doubleval
- empty
- floatval
- get_debug_type
- gettype
- intval
- is_array
- is_bool
- is_double
- is_float
- is_int
- is_integer
- is_long
- is_null
- is_scalar
- is_string
- isset
- print_r
- strval
- unset
- var_dump
- var_export

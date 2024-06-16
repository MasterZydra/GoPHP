package common

import (
	"regexp"
	"strings"
)

func IsNondigit(char string) bool {
	// Spec: https://phplang.org/spec/19-grammar.html#grammar-nondigit

	// nondigit:: one of
	//    _
	//    a   b   c   d   e   f   g   h   i   j   k   l   m
	//    n   o   p   q   r   s   t   u   v   w   x   y   z
	//    A   B   C   D   E   F   G   H   I   J   K   L   M
	//    N   O   P   Q   R   S   T   U   V   W   X   Y   Z

	match, _ := regexp.MatchString("^[a-zA-Z_]$", char)
	return match
}

func IsNameNondigit(char string) bool {
	// Spec: https://phplang.org/spec/19-grammar.html#grammar-name-nondigit

	// name-nondigit::
	//    nondigit
	//    one of the characters 0x80–0xff

	b := []byte(char)[0]
	return IsNondigit(char) || (b >= 128 && b <= 255)
}

func IsSingleQuotedStringLiteral(str string) bool {
	// Spec: https://phplang.org/spec/09-lexical-structure.html#grammar-single-quoted-string-literal

	// single-quoted-string-literal::
	//    b-prefix(opt)   '   sq-char-sequence(opt)   '

	// sq-char-sequence::
	//    sq-char
	//    sq-char-sequence   sq-char

	// sq-char::
	//    sq-escape-sequence
	//    \(opt)   any member of the source character set except single-quote (') or backslash (\)

	// sq-escape-sequence:: one of
	//    \'   \\

	//  b-prefix:: one of
	//    b   B

	match, _ := regexp.MatchString(`^[bB]?'(\\?[^'\\]|\\'|\\\\)*'`, str)
	return match
}

func ReplaceSingleQuoteControlChars(str string) string {
	findCtrlChars := regexp.MustCompile(`(\\?[^'\\]|\\'|\\\\)`)
	return findCtrlChars.ReplaceAllStringFunc(str,
		func(sub string) string {
			switch sub {
			case `\'`:
				return `'`
			case `\\`:
				return `\`
			default:
				return sub
			}
		},
	)
}

func extractStringContent(str string) string {
	if strings.ToLower(str[0:1]) == "b" {
		return str[2 : len(str)-1]
	}
	return str[1 : len(str)-1]
}

func SingleQuotedStringLiteralToString(str string) string {
	return ReplaceSingleQuoteControlChars(extractStringContent(str))
}

func IsDoubleQuotedStringLiteral(str string) bool {
	// Spec: https://phplang.org/spec/09-lexical-structure.html#grammar-double-quoted-string-literal

	// double-quoted-string-literal::
	//    b-prefix(opt)   "   dq-char-sequence(opt)   "

	//  b-prefix:: one of
	//    b   B

	// dq-char-sequence::
	//    dq-char
	//    dq-char-sequence   dq-char

	// dq-char::
	//    dq-escape-sequence
	//    any member of the source character set except double-quote (") or backslash (\)
	//    \   any member of the source character set except "\$efnrtvxX or   octal-digit

	// dq-escape-sequence::
	//    dq-simple-escape-sequence
	//    dq-octal-escape-sequence
	//    dq-hexadecimal-escape-sequence
	//    dq-unicode-escape-sequence

	// dq-simple-escape-sequence:: one of
	//    \"   \\   \$   \e   \f   \n   \r   \t   \v

	// dq-octal-escape-sequence::
	//    \   octal-digit
	//    \   octal-digit   octal-digit
	//    \   octal-digit   octal-digit   octal-digit

	// octal-digit:: one of
	//    0   1   2   3   4   5   6   7

	// dq-hexadecimal-escape-sequence::
	//    \x   hexadecimal-digit   hexadecimal-digit(opt)
	//    \X   hexadecimal-digit   hexadecimal-digit(opt)

	// dq-unicode-escape-sequence::
	//    \u{   codepoint-digits   }

	// codepoint-digits::
	//    hexadecimal-digit
	//    hexadecimal-digit   codepoint-digits

	// hexadecimal-digit:: one of
	//    0   1   2   3   4   5   6   7   8   9
	//    a   b   c   d   e   f
	//    A   B   C   D   E   F

	match, _ := regexp.MatchString(
		`^[bB]?"(`+
			`(\\"|\\\\|\\\$|\\e|\\f|\\n|\\r|\\t|\\v)|`+
			`(\\[0-7]{1,3})|`+
			`(\\[xX][0-9a-fA-F]{1,2})|`+
			`(\\u\{[0-9a-fA-F]+\})|`+
			`([^"\\])|`+
			`(\\[^"\\$efnrtvxX]|\\[0-7])`+
			`)*"$`,
		str)
	return match
}

func ReplaceDoubleQuoteControlChars(str string) string {
	findCtrlChars := regexp.MustCompile(
		`(\\"|\\\\|\\\$|\\e|\\f|\\n|\\r|\\t|\\v)|` +
			`(\\[0-7]{1,3})|` +
			`(\\[xX][0-9a-fA-F]{1,2})|` +
			`(\\u\{[0-9a-fA-F]+\})|` +
			`([^"\\])|` +
			`(\\[^"\\$efnrtvxX]|\\[0-7])`,
	)

	return findCtrlChars.ReplaceAllStringFunc(str,
		func(sub string) string {
			switch sub {
			case `\"`:
				return `"`
			case `\\`:
				return `\`
			case `\r`:
				return "\r"
			case `\n`:
				return "\n"
			case `\t`:
				return "\t"
			default:
				return sub
			}
		},
	)
}

func DoubleQuotedStringLiteralToString(str string) string {
	return ReplaceDoubleQuoteControlChars(extractStringContent(str))
}

func TrimTrailingLineBreak(str string) string {
	if len(str) > 0 && str[len(str)-1] == '\n' {
		str = str[:len(str)-1]
	}
	return str
}

func TrimTrailingLineBreaks(str string) string {
	for len(str) > 0 && str[len(str)-1] == '\n' {
		str = str[:len(str)-1]
	}
	return str
}

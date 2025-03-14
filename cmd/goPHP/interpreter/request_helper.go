package interpreter

import (
	"GoPHP/cmd/goPHP/common"
	"GoPHP/cmd/goPHP/config"
	"GoPHP/cmd/goPHP/ini"
	"GoPHP/cmd/goPHP/request"
	"GoPHP/cmd/goPHP/runtime/values"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

func parseCookies(cookies string) *values.Array {
	result := values.NewArray()

	for cookies != "" {
		var cookie string
		cookie, cookies, _ = strings.Cut(cookies, ";")
		if cookie == "" {
			continue
		}

		var name string
		var value string
		if !strings.Contains(cookie, "=") {
			// Cookie without value is an empty string
			name = cookie
			value = ""
		} else {
			// Get parameter with key-value-pair
			name, value, _ = strings.Cut(cookie, "=")
		}
		name = strings.Trim(name, " ")
		key := values.NewStr(strings.NewReplacer(
			" ", "_",
			"[", "_",
			".", "_",
		).Replace(name))
		if result.Contains(key) {
			continue
		}
		// Escape plus sign so that it will not be replaced with space
		value = strings.ReplaceAll(value, "+", "%2b")
		value, err := url.QueryUnescape(fixPercentEscaping(value))
		if err != nil {
			if config.IsDevMode {
				println("parseCookies: ", err)
			}
			continue
		}
		result.SetElement(key, values.NewStr(value))
	}

	return result
}

func parseQuery(query string, ini *ini.Ini) (*values.Array, error) {
	result := values.NewArray()

	for query != "" {
		var key string
		key, query, _ = strings.Cut(query, ini.GetStr("arg_separator.input"))
		if key == "" {
			continue
		}

		// Get parameters without key e.g. ab+cd+ef
		// TODO this is only correct if it is in "phpt mode". "Normal" GET will parse it differently
		// ab+cd+ef => array(1) { ["ab_cd_ef"]=> string(0) "" }
		if !strings.Contains(key, "=") && strings.Contains(key, "+") {
			parts := strings.Split(key, "+")
			for i := 0; i < len(parts); i++ {
				if err := result.SetElement(nil, values.NewStr(parts[i])); err != nil {
					return result, err
				}
			}
			continue
		}

		// Get parameter with key-value-pair
		key, value, _ := strings.Cut(key, "=")

		key, err := url.QueryUnescape(fixPercentEscaping(key))
		if err != nil {
			return result, err
		}

		value, err = url.QueryUnescape(value)
		if err != nil {
			return result, err
		}
		if strings.Contains(key, "[") && strings.Contains(key, "]") {
			result, err = parseQueryKey(key, value, result)
			if err != nil {
				return result, err
			}
		} else {
			key = strings.NewReplacer(
				" ", "_",
				"+", "_",
				"[", "_",
				".", "_",
			).Replace(key)

			var keyValue values.RuntimeValue
			if common.IsIntegerLiteral(key, false) {
				intValue, _ := common.IntegerLiteralToInt64(key, false)
				keyValue = values.NewInt(intValue)
			} else {
				keyValue = values.NewStr(key)
			}
			result.SetElement(keyValue, values.NewStr(value))
		}
	}

	return result, nil
}

func parseQueryKey(key string, value string, result *values.Array) (*values.Array, error) {
	// The parsing of a complex key with arrays is solved by using the interpreter itself:
	// The key and value is transformed into valid PHP code and executed.
	// Example:
	//   Input: 123[][12][de]=abc
	//   Key:   123[][12][de]
	//   Value: abc
	//   PHP:   $array[123][][12]["de"] = "abc";

	firstKey, key, _ := strings.Cut(key, "[")
	key = "[" + key

	phpArrayKeys := []string{firstKey}

	for key != "" {
		key = strings.TrimPrefix(key, "[")
		var nextKey string
		nextKey, key, _ = strings.Cut(key, "]")
		phpArrayKeys = append(phpArrayKeys, nextKey)
		for key != "" && !strings.HasPrefix(key, "[") {
			key = key[1:]
		}
	}

	php := "<?php $array"
	for _, phpArrayKey := range phpArrayKeys {
		if phpArrayKey == "" {
			php += "[]"
		} else if common.IsIntegerLiteral(phpArrayKey, false) {
			phpArrayKeyInt, _ := common.IntegerLiteralToInt64(phpArrayKey, false)
			php += fmt.Sprintf("[%d]", phpArrayKeyInt)
		} else {
			php += "['" + phpArrayKey + "']"
		}
	}
	php += " = '" + value + "';"

	interpreter := NewInterpreter(ini.NewDefaultIni(), &request.Request{}, "")
	interpreter.env.declareVariable("$array", result)
	_, err := interpreter.Process(php)

	return interpreter.env.variables["$array"].(*values.Array), err
}

// This fix is required because "url.QueryUnescape()" cannot handle an unescaped percent
func fixPercentEscaping(key string) string {
	re, _ := regexp.Compile("%([^0-9A-Fa-f]|$)")
	// Replace only the '%' character with '%25' without affecting the following character
	return re.ReplaceAllStringFunc(key, func(match string) string {
		return "%25" + match[1:]
	})
}

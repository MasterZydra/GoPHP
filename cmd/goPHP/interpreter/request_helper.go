package interpreter

import (
	"GoPHP/cmd/goPHP/common"
	"GoPHP/cmd/goPHP/config"
	"GoPHP/cmd/goPHP/ini"
	"GoPHP/cmd/goPHP/phpError"
	"GoPHP/cmd/goPHP/request"
	"GoPHP/cmd/goPHP/runtime"
	"GoPHP/cmd/goPHP/runtime/values"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strings"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
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

func parsePost(query string, interpreter runtime.Interpreter) (*values.Array, *values.Array, error) {
	postArray := values.NewArray()
	filesArray := values.NewArray()

	if strings.HasPrefix(query, "Content-Type: multipart/form-data;") {
		// TODO Improve code
		var boundary string
		lines := strings.Split(query, "\n")
		lineNum := 0
		for {
			if lineNum >= len(lines) {
				break
			}

			if lineNum == 0 {
				boundary = strings.Replace(lines[lineNum], "Content-Type: multipart/form-data;", "", 1)
				boundary = strings.Replace(strings.TrimSpace(boundary), "boundary=", "", 1)
				if strings.HasPrefix(boundary, "\"") {
					boundary = boundary[1:]
					if strings.Contains(boundary, "\"") {
						boundary = boundary[:strings.Index(boundary, "\"")]
					}
				}
				boundary = "--" + boundary
				if strings.Contains(boundary, ";") {
					// Content-Type: multipart/form-data; boundary=abc; charset=...
					boundary = boundary[:strings.Index(boundary, ";")]
				} else if strings.Contains(boundary, ",") {
					// Content-Type: multipart/form-data; boundary=abc, charset=...
					boundary = boundary[:strings.Index(boundary, ",")]
				}
				lineNum++
				continue
			}

			if lines[lineNum] == boundary+"--" {
				return postArray, filesArray, nil
			}

			if lines[lineNum] == boundary {
				lineNum++
				if strings.HasPrefix(lines[lineNum], "Content-Disposition: form-data;") {
					isFile := strings.Contains(lines[lineNum], "filename=")
					fullname := strings.Replace(lines[lineNum], "Content-Disposition: form-data;", "", 1)

					name := strings.Replace(strings.TrimSpace(fullname), "name=", "", 1)
					if strings.Contains(name, ";") {
						name = name[:strings.Index(name, ";")]
					}
					if strings.HasPrefix(name, "'") {
						name = name[1:strings.LastIndex(name, "'")]
						name = strings.ReplaceAll(name, `\'`, "'")
					}
					if strings.HasPrefix(name, "\"") {
						name = name[1:strings.LastIndex(name, "\"")]
						name = strings.ReplaceAll(name, `\"`, `"`)
					}
					name = strings.ReplaceAll(name, `\\`, `\`)
					name = recode(name, interpreter.GetIni())

					filename := ""
					contentType := ""
					if isFile {
						filename = fullname[strings.Index(fullname, "filename="):]
						filename = strings.TrimPrefix(filename, "filename=")
						if strings.HasPrefix(filename, "\"") {
							filename = filename[1:strings.LastIndex(filename, "\"")]
						}
						filename = recode(filename, interpreter.GetIni())
						lineNum++
						if strings.HasPrefix(lines[lineNum], "Content-Type:") {
							contentType = strings.TrimPrefix(lines[lineNum], "Content-Type:")
							contentType = strings.TrimSpace(contentType)
						} else {
							lineNum--
						}
					}

					lineNum += 2
					content := ""
					for lineNum < len(lines) && lines[lineNum] != boundary && lines[lineNum] != boundary+"--" {
						content += lines[lineNum]
						if isFile {
							content += "\n"
						}
						lineNum++
					}
					if isFile && strings.HasSuffix(content, "\n\n") {
						content = content[:len(content)-1]
					}
					content = recode(content, interpreter.GetIni())
					if len(content) > getPostMaxSize(interpreter.GetIni()) {
						interpreter.PrintError(phpError.NewWarning(
							"PHP Request Startup: POST Content-Length of %d bytes exceeds the limit of %d bytes in Unknown on line 0",
							len(content),
							getPostMaxSize(interpreter.GetIni()),
						))
						continue
					}

					if !isFile {
						postArray.SetElement(values.NewStr(name), values.NewStr(content))
					} else {
						data := values.NewArray()
						data.SetElement(values.NewStr("name"), values.NewStr(filename))
						data.SetElement(values.NewStr("full_path"), values.NewStr(filename))
						data.SetElement(values.NewStr("type"), values.NewStr(contentType))
						// TODO store to file
						data.SetElement(values.NewStr("tmp_name"), values.NewStr("tmp.file"))
						data.SetElement(values.NewStr("error"), values.NewInt(0))
						data.SetElement(values.NewStr("size"), values.NewInt(int64(len(content))))

						filesArray.SetElement(values.NewStr(name), data)
					}
					continue
				}
				lineNum++
				continue
			}
			return postArray, filesArray, fmt.Errorf("parsePost - Unexpected line %d: %s", lineNum, lines[lineNum])
		}
		return postArray, filesArray, nil
	}

	postArray, err := parseQuery(strings.ReplaceAll(query, "\n", ""), interpreter)
	return postArray, filesArray, err
}

func parseQuery(query string, interpreter runtime.Interpreter) (*values.Array, error) {
	result := values.NewArray()

	for query != "" {
		var key string
		key, query, _ = strings.Cut(query, interpreter.GetIni().GetStr("arg_separator.input"))
		if key == "" {
			continue
		}

		// Get parameters without key e.g. ab+cd+ef
		// TODO this is only correct if it is in "phpt mode". "Normal" GET will parse it differently
		// ab+cd+ef => array(1) { ["ab_cd_ef"]=> string(0) "" }
		if !strings.Contains(key, "=") && strings.Contains(key, "+") {
			parts := strings.Split(key, "+")
			for i := 0; i < len(parts); i++ {
				if len(parts[i]) > getPostMaxSize(interpreter.GetIni()) {
					interpreter.PrintError(phpError.NewWarning(
						"PHP Request Startup: POST Content-Length of %d bytes exceeds the limit of %d bytes in Unknown on line 0",
						len(parts[i]),
						getPostMaxSize(interpreter.GetIni()),
					))
					continue
				}
				if err := result.SetElement(nil, values.NewStr(parts[i])); err != nil {
					return result, err
				}
			}
			continue
		}

		if len(key) > getPostMaxSize(interpreter.GetIni()) {
			interpreter.PrintError(phpError.NewWarning(
				"PHP Request Startup: POST Content-Length of %d bytes exceeds the limit of %d bytes in Unknown on line 0",
				len(key),
				getPostMaxSize(interpreter.GetIni()),
			))
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
			result, err = parseQueryKey(key, value, result, interpreter.GetIni())
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

func parseQueryKey(key string, value string, result *values.Array, curIni *ini.Ini) (*values.Array, error) {
	// The parsing of a complex key with arrays is solved by using the interpreter itself:
	// The key and value is transformed into valid PHP code and executed.
	// Example:
	//   Input: 123[][12][de]=abc
	//   Key:   123[][12][de]
	//   Value: abc
	//   PHP:   $array[123][][12]["de"] = "abc";

	firstKey, key, _ := strings.Cut(key, "[")
	key = "[" + key

	maxDepth := curIni.GetInt("max_input_nesting_level")

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
	for depth, phpArrayKey := range phpArrayKeys {
		if depth+1 >= int(maxDepth) {
			return result, nil
		}
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

func getPostMaxSize(ini *ini.Ini) int {
	sizeStr := ini.GetStr("post_max_size")
	if common.IsDecimalLiteral(sizeStr, false) {
		size := int(common.DecimalLiteralToInt64(sizeStr, false))
		if size == 0 {
			return 8 * 1024 * 1024
		}
		return size
	}
	if strings.HasSuffix(sizeStr, "K") {
		return int(common.DecimalLiteralToInt64(strings.Replace(sizeStr, "K", "", 1), false) * 1024)
	}
	if strings.HasSuffix(sizeStr, "M") {
		return int(common.DecimalLiteralToInt64(strings.Replace(sizeStr, "M", "", 1), false) * 1024 * 1024)
	}
	if strings.HasSuffix(sizeStr, "G") {
		return int(common.DecimalLiteralToInt64(strings.Replace(sizeStr, "G", "", 1), false) * 1024 * 1024 * 1024)
	}
	return 8 * 1024 * 1024
}

func recode(input string, ini *ini.Ini) string {
	var decoder *transform.Reader
	var encoder transform.Transformer

	inputEncoding := ini.GetStr("input_encoding")
	if inputEncoding == "" {
		inputEncoding = ini.GetStr("default_charset")
	}
	switch inputEncoding {
	case "Shift_JIS":
		decoder = transform.NewReader(strings.NewReader(input), japanese.ShiftJIS.NewDecoder())
	case "UTF-8":
		decoder = transform.NewReader(strings.NewReader(input), unicode.UTF8.NewDecoder())
	default:
		fmt.Println("changeEncoding: Unsupported from encoding: ", inputEncoding)
		return ""
	}

	decodedBytes, err := io.ReadAll(decoder)
	if err != nil {
		fmt.Println("changeEncoding: Error decoding input: ", err)
		return ""
	}

	internalEncoding := ini.GetStr("internal_encoding")
	if internalEncoding == "" {
		internalEncoding = ini.GetStr("default_charset")
	}
	switch internalEncoding {
	case "Shift_JIS":
		encoder = japanese.ShiftJIS.NewEncoder()
	case "UTF-8":
		encoder = unicode.UTF8.NewEncoder()
	default:
		fmt.Println("changeEncoding: Unsupported from encoding: ", internalEncoding)
		return ""
	}

	encodedBytes, err := io.ReadAll(transform.NewReader(strings.NewReader(string(decodedBytes)), encoder))
	if err != nil {
		fmt.Println("changeEncoding: Error encoding output: ", err)
		return ""
	}

	return string(encodedBytes)
}

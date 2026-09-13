package main

import (
	"generator/fb"
	"strings"
	"unicode"
)

func CamelToSnakeCase(str string) string {
	var sb strings.Builder
	lastUpper := false

	for i, ch := range str {
		if unicode.IsDigit(ch) {
			sb.WriteRune(ch)
			lastUpper = true
		} else if unicode.IsUpper(ch) {
			if !lastUpper && i > 0 {
				sb.WriteRune('_')
			}
			sb.WriteRune(unicode.ToLower(ch))
			lastUpper = true
		} else {
			sb.WriteRune(ch)
			lastUpper = false
		}
	}

	return sb.String()
}

func SnakeToPascalCase(str string) string {
	var sb strings.Builder
	nextUpper := true

	for _, ch := range str {
		if ch == '_' {
			nextUpper = true
			continue
		}

		if nextUpper {
			sb.WriteRune(unicode.ToUpper(ch))
			nextUpper = false
		} else {
			sb.WriteRune(unicode.ToLower(ch))
		}

		if unicode.IsDigit(ch) {
			nextUpper = true
		}
	}

	return sb.String()
}

func GetDeclOrder(decl fb.Decl) int {
	switch decl.(type) {
	case *fb.Alias:
		return 1
	case *fb.Enum:
		return 2
	case *fb.Struct:
		return 3
	case *fb.Func:
		return 4

	default:
		panic("bindgen.GetDeclOrder() - Invalid declaration")
	}
}

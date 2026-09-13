package main

import (
	"generator/fb"
	"strconv"
	"strings"
)

func ParseCType(str string, custom func(str string) fb.Type) fb.Type {
	str = strings.TrimSpace(str)

	// Array
	if strings.HasSuffix(str, "]") {
		paren := strings.LastIndexByte(str, '[')

		sizeStr := str[paren+1 : len(str)-1]
		var size uint32 = 0

		if sizeStr != "" {
			parsedSize, err := strconv.ParseUint(sizeStr, 10, 32)
			if err != nil {
				panic(err.Error())
			}

			size = uint32(parsedSize)
		}

		return &fb.ArrayType{
			Size:    size,
			Element: ParseCType(str[:paren], custom),
		}
	}

	// Pointer
	str = strings.TrimSuffix(str, "const")
	str = strings.TrimSpace(str)

	if strings.HasSuffix(str, "*") {
		pointeeStr := strings.TrimSpace(str[:len(str)-1])
		isConst := strings.Contains(pointeeStr, "const ") || strings.HasSuffix(pointeeStr, "const")

		return &fb.PointerType{
			Mutable: !isConst,
			Pointee: ParseCType(pointeeStr, custom),
		}
	}

	// Function

	if idx := strings.Index(str, "(*)"); idx != -1 {
		returnsStr := strings.TrimSpace(str[:idx])
		paramsStr := strings.TrimSpace(str[idx+3:])

		if strings.HasPrefix(paramsStr, "(") && strings.HasSuffix(paramsStr, ")") {
			paramsStr = paramsStr[1 : len(paramsStr)-1]
		}

		var params []fb.Param

		if paramsStr != "" && paramsStr != "void" {
			for _, arg := range splitParams(paramsStr) {
				params = append(params, fb.Param{
					Name: "",
					Type: ParseCType(arg, custom),
				})
			}
		}

		return &fb.FuncType{
			Params:  params,
			Returns: ParseCType(returnsStr, custom),
		}
	}

	// Simple
	str = strings.ReplaceAll(str, "const ", "")
	str = strings.ReplaceAll(str, "volatile ", "")
	str = strings.ReplaceAll(str, "restrict ", "")
	str = strings.TrimSpace(str)

	str = strings.TrimPrefix(str, "struct ")
	str = strings.TrimPrefix(str, "enum ")
	str = strings.TrimPrefix(str, "union ")

	simple := mapSimpleType(str)

	if simple != "" {
		return &fb.SimpleType{Text: simple}
	}

	// Custom
	return custom(str)
}

func splitParams(str string) []string {
	var params []string

	var sb strings.Builder
	depth := 0

	for _, r := range str {
		switch r {
		case '(':
			depth++
		case ')':
			depth--
		case ',':
			if depth == 0 {
				params = append(params, strings.TrimSpace(sb.String()))
				sb.Reset()
				continue
			}
		}

		sb.WriteRune(r)
	}

	if sb.Len() > 0 {
		params = append(params, strings.TrimSpace(sb.String()))
	}

	return params
}

func mapSimpleType(cType string) string {
	switch cType {
	// -- 8-bit Types --
	case "char":
		return "u8"
	case "unsigned char":
		return "u8"
	case "uint8_t":
		return "u8"
	case "signed char":
		return "i8"
	case "int8_t":
		return "i8"

	// -- 16-bit Types --
	case "short", "signed short":
		return "i16"
	case "unsigned short":
		return "u16"
	case "int16_t":
		return "i16"
	case "uint16_t":
		return "u16"

	// -- 32-bit Types --
	case "int", "signed int":
		return "i32"
	case "unsigned int", "unsigned":
		return "u32"
	case "int32_t":
		return "i32"
	case "uint32_t":
		return "u32"

	// -- 64-bit Types --
	case "long long", "signed long long":
		return "i64"
	case "unsigned long long":
		return "u64"
	case "int64_t":
		return "i64"
	case "uint64_t":
		return "u64"

	// -- Floats --
	case "float":
		return "f32"
	case "double":
		return "f64"

	// -- Long Types --
	case "long", "signed long":
		return "c_long"
	case "unsigned long":
		return "c_ulong"

	// -- Pointer-sized Types --
	case "size_t":
		return "u64"
	case "ssize_t":
		return "u64"
	case "ptrdiff_t":
		return "u64"
	case "intptr_t":
		return "u64"
	case "uintptr_t":
		return "u64"

	// -- Misc --
	case "void":
		return "void"
	case "bool", "_Bool":
		return "bool"

	default:
		return ""
	}
}

package main

import (
	"cmp"
	"fmt"
	"generator/c"
	"generator/fb"
	"io"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

type BindGenOpts struct {
	Inputs  []string
	Defines []string

	Outputs []File

	NameMappings map[string]string
	MacroEnums   []MacroEnum

	FilterAlias  func(node *c.Node) bool
	FilterEnum   func(node *c.Node) bool
	FilterStruct func(node *c.Node) bool
	FilterFunc   func(node *c.Node) bool

	ParseType func(str string) fb.Type

	TransformAlias  func(a *fb.Alias)
	TransformEnum   func(e *fb.Enum)
	TransformStruct func(s *fb.Struct)
	TransformFunc   func(f *fb.Func)

	Order func(a, b fb.Decl) int
}

type File struct {
	Path   string
	Module string
}

type MacroEnum struct {
	Name     string
	Type     string
	Bitfield bool

	Prefix string
	Suffix string

	Exact []string
}

type generator struct {
	opts BindGenOpts

	nodes []*c.Node

	decls     []declNodePair
	fileDecls [][]fb.Decl

	structMethods map[*fb.Struct][]*fb.Func
}

type declNodePair struct {
	decl fb.Decl
	node *c.Node
}

func Generate(opts BindGenOpts) error {
	g := generator{
		opts:          opts,
		fileDecls:     make([][]fb.Decl, len(opts.Outputs)),
		structMethods: make(map[*fb.Struct][]*fb.Func),
	}

	g.SetupDefaults()

	if err := g.CollectNodes(); err != nil {
		return err
	}

	g.CreateDecls()
	g.ResolveDecls()
	g.TransformDecls()

	for i := range g.opts.Outputs {
		if err := g.WriteFile(i); err != nil {
			return err
		}
	}

	return nil
}

func (g *generator) CollectNodes() error {
	for _, input := range g.opts.Inputs {
		root, err := c.Parse(input, g.opts.Defines)
		if root == nil {
			return err
		}

		for i, node := range root.Inner {
			switch node.Kind {
			case "TypedefDecl":
				if g.opts.FilterAlias(node) && g.Node(node.Name) == nil {
					g.nodes = append(g.nodes, node)
				}

			case "EnumDecl":
				name := node.Name

				if name == "" && i+1 < len(root.Inner) {
					next := root.Inner[i+1]
					if next.Kind == "TypedefDecl" {
						name = next.Name
						node.Name = name
					}
				}

				if name != "" {
					if g.opts.FilterEnum(node) {
						existing := g.Node(name)

						if existing != nil && existing.Kind == "TypedefDecl" {
							idx := slices.Index(g.nodes, existing)
							g.nodes = slices.Delete(g.nodes, idx, idx+1)

							existing = nil
						}

						if existing == nil {
							g.nodes = append(g.nodes, node)
						}
					}
				}

			case "RecordDecl":
				if node.Tag == "struct" || node.Tag == "union" {
					name := node.Name

					if name == "" && i+1 < len(root.Inner) {
						next := root.Inner[i+1]

						if next.Kind == "TypedefDecl" {
							name = next.Name
							node.Name = name
						}
					}

					if name != "" {
						if g.opts.FilterStruct(node) {
							exiting := g.Node(name)

							if exiting != nil && exiting.Kind == "TypedefDecl" {
								idx := slices.Index(g.nodes, exiting)
								g.nodes = slices.Delete(g.nodes, idx, idx+1)

								exiting = nil
							}

							if exiting == nil {
								g.nodes = append(g.nodes, node)
							}
						}
					}
				}

			case "FunctionDecl":
				if g.opts.FilterFunc(node) && g.Node(node.Name) == nil {
					g.nodes = append(g.nodes, node)
				}

			case "MacroDefine":
				for _, enum := range g.opts.MacroEnums {
					if enum.contains(node.Name) {
						value := strings.Trim(node.Value, "()")
						value = strings.TrimRight(value, "uUlL")

						if strings.HasPrefix(value, "0x") {
							_, err := strconv.ParseUint(value[2:], 16, 64)
							if err != nil {
								break
							}
						} else if strings.HasPrefix(value, "0b") {
							_, err := strconv.ParseUint(value[2:], 2, 64)
							if err != nil {
								break
							}
						} else {
							_, err := strconv.ParseInt(value, 10, 64)
							if err != nil {
								break
							}
						}

						g.nodes = append(g.nodes, node)
						break
					}
				}
			}
		}
	}

	return nil
}

func (g *generator) CreateDecls() {
	g.decls = make([]declNodePair, 0, len(g.nodes)+len(g.opts.MacroEnums))

	for _, node := range g.nodes {
		var decl fb.Decl

		switch node.Kind {
		case "TypedefDecl":
			decl = &fb.Alias{
				Documentation: node.Documentation(),
				Name:          node.Name,
			}

		case "EnumDecl":
			decl = &fb.Enum{
				Documentation: node.Documentation(),
				Name:          node.Name,
			}

		case "RecordDecl":
			decl = &fb.Struct{
				Documentation: node.Documentation(),
				Name:          node.Name,
				Union:         node.Tag == "union",
			}

		case "FunctionDecl":
			decl = &fb.Func{
				Documentation: node.Documentation(),
				Name:          CamelToSnakeCase(g.MapName(node.Name)),
				LinkName:      node.Name,
				ReceiverIndex: -1,
			}
		}

		if decl != nil {
			g.decls = append(g.decls, declNodePair{
				decl: decl,
				node: node,
			})
		}
	}

	for _, enum := range g.opts.MacroEnums {
		var typ fb.Type

		if enum.Type != "" {
			typ = &fb.SimpleType{Text: enum.Type}
		}

		g.decls = append(g.decls, declNodePair{
			decl: &fb.Enum{
				Name:     enum.Name,
				Type:     typ,
				Bitfield: enum.Bitfield,
			},
			node: nil,
		})
	}
}

func (g *generator) ResolveDecls() {
	for _, pair := range g.decls {
		node := pair.node

		switch decl := pair.decl.(type) {
		case *fb.Alias:
			decl.Type = g.ParseType(node.Type.QualType)

		case *fb.Enum:
			macroEnumIdx := slices.IndexFunc(g.opts.MacroEnums, func(enum MacroEnum) bool {
				return enum.Name == decl.Name
			})

			if macroEnumIdx != -1 {
				enum := g.opts.MacroEnums[macroEnumIdx]

				for _, node := range g.nodes {
					if enum.contains(node.Name) {
						value := strings.Trim(node.Value, "()")
						value = strings.TrimRight(value, "uUlL")

						decl.Cases = append(decl.Cases, &fb.Case{
							Documentation: node.Documentation(),
							Name:          SnakeToPascalCase(enum.trim(node.Name)),
							Value:         strings.TrimRight(value, "uUlL"),
						})

						if decl.Documentation == "" && node.GroupText != "" {
							decl.Documentation = node.GroupText
						}
					}
				}
			} else {
				for _, node := range node.Inner {
					if node.Kind != "EnumConstantDecl" {
						continue
					}

					decl.Cases = append(decl.Cases, &fb.Case{
						Documentation: node.Documentation(),
						Name:          SnakeToPascalCase(node.Name),
						Value:         node.EnumValue(),
					})
				}
			}

			if decl.Type == nil {
				negative := false
				maxValue := uint64(0)

				for _, cas := range decl.Cases {
					str := cas.Value

					if strings.HasPrefix(str, "-") {
						negative = true
						str = str[1:]
					}

					if strings.HasPrefix(str, "0x") {
						num, _ := strconv.ParseUint(str[2:], 16, 64)
						maxValue = max(maxValue, num)
					} else if strings.HasPrefix(str, "0b") {
						num, _ := strconv.ParseUint(str[2:], 2, 64)
						maxValue = max(maxValue, num)
					} else {
						num, _ := strconv.ParseUint(str, 10, 64)
						maxValue = max(maxValue, num)
					}
				}

				if negative {
					if maxValue > math.MaxInt32 {
						decl.Type = &fb.SimpleType{Text: "i64"}
					} else {
						decl.Type = &fb.SimpleType{Text: "i32"}
					}
				} else {
					if maxValue > math.MaxUint32 {
						decl.Type = &fb.SimpleType{Text: "u64"}
					} else {
						decl.Type = &fb.SimpleType{Text: "u32"}
					}
				}
			}

		case *fb.Struct:
			for _, node := range node.Inner {
				if node.Kind == "FieldDecl" {
					decl.Fields = append(decl.Fields, &fb.Field{
						Documentation: node.Documentation(),
						Name:          node.Name,
						Type:          g.ParseType(node.Type.QualType),
					})
				}
			}

		case *fb.Func:
			for _, node := range node.Inner {
				if node.Kind == "ParmVarDecl" {
					decl.Params = append(decl.Params, &fb.Param{
						Name: CamelToSnakeCase(node.Name),
						Type: g.ParseType(node.Type.QualType),
					})
				}
			}

			paren := strings.IndexRune(node.Type.QualType, '(')
			decl.Returns = g.ParseType(node.Type.QualType[:paren])
		}
	}
}

func (g *generator) TransformDecls() {
	for _, pair := range g.decls {
		decl := pair.decl

		switch decl := decl.(type) {
		case *fb.Alias:
			g.opts.TransformAlias(decl)

		case *fb.Enum:
			g.opts.TransformEnum(decl)

		case *fb.Struct:
			g.opts.TransformStruct(decl)

		case *fb.Func:
			g.opts.TransformFunc(decl)

			if decl.ReceiverIndex >= 0 {
				s := decl.Params[decl.ReceiverIndex].Type.(*fb.PointerType).Pointee.(*fb.DeclType).Decl.(*fb.Struct)
				g.structMethods[s] = append(g.structMethods[s], decl)
			}
		}

		i := decl.OutputIndex_()
		g.fileDecls[i] = append(g.fileDecls[i], decl)
	}
}

func (g *generator) WriteFile(i int) error {
	output := g.opts.Outputs[i]

	err := os.MkdirAll(filepath.Dir(output.Path), 0750)
	if err != nil {
		return err
	}

	file, err := os.Create(output.Path)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(file, "mod %s;\n\n", output.Module)

	_, _ = fmt.Fprint(file, "// ----------------------------------------\n")
	_, _ = fmt.Fprint(file, "// | This file is automatically generated |\n")
	_, _ = fmt.Fprint(file, "// ----------------------------------------\n")

	decls := slices.SortedStableFunc(slices.Values(g.fileDecls[i]), g.opts.Order)

	for _, decl := range decls {
		_, _ = fmt.Fprint(file, "\n")
		decl.Write(file)

		if s, ok := decl.(*fb.Struct); ok {
			if methods, ok := g.structMethods[s]; ok {
				_, _ = fmt.Fprint(file, "\n")
				g.WriteMethods(file, s, methods)
			}
		}
	}

	_ = file.Close()

	return nil
}

func (g *generator) WriteMethods(w io.Writer, s *fb.Struct, methods []*fb.Func) {
	_, _ = fmt.Fprintf(w, "impl %s {\n", s.Name)

	for i, method := range methods {
		if i > 0 {
			_, _ = fmt.Fprint(w, "\n")
		}

		fb.WriteDocumentation(w, method.Documentation, "    ")

		// Signature
		name := method.MethodName
		if name == "" {
			name = method.Name
		}

		_, _ = fmt.Fprintf(w, "    pub func %s(", name)

		if method.Params[method.ReceiverIndex].Type.(*fb.PointerType).Mutable {
			_, _ = fmt.Fprint(w, "mut self")
		} else {
			_, _ = fmt.Fprint(w, "self")
		}

		for i, param := range method.Params {
			if i == method.ReceiverIndex {
				continue
			}

			_, _ = fmt.Fprintf(w, ", %s: ", param.Name)
			param.Type.Write(w)
		}

		_, _ = fmt.Fprint(w, ")")

		// Returns
		returns := false

		if s, ok := method.Returns.(*fb.SimpleType); !ok || s.Text != "void" {
			_, _ = fmt.Fprint(w, " ")
			method.Returns.Write(w)

			returns = true
		}

		// Body
		_, _ = fmt.Fprint(w, " {\n")

		if returns {
			_, _ = fmt.Fprint(w, "        return ")
		} else {
			_, _ = fmt.Fprint(w, "        ")
		}

		_, _ = fmt.Fprintf(w, "%s(", method.Name)

		for i, param := range method.Params {
			if i > 0 {
				_, _ = fmt.Fprint(w, ", ")
			}

			if i == method.ReceiverIndex {
				_, _ = fmt.Fprint(w, "self")
				continue
			}

			_, _ = fmt.Fprint(w, param.Name)
		}

		_, _ = fmt.Fprint(w, ");\n    }\n")
	}

	_, _ = fmt.Fprint(w, "}\n")
}

func (g *generator) Node(name string) *c.Node {
	for _, node := range g.nodes {
		if node.Name == name {
			return node
		}
	}

	return nil
}

func (g *generator) SetupDefaults() {
	// Filters
	if g.opts.FilterAlias == nil {
		g.opts.FilterAlias = func(node *c.Node) bool {
			return slices.Contains(g.opts.Inputs, node.Loc.File)
		}
	}

	if g.opts.FilterStruct == nil {
		g.opts.FilterStruct = func(node *c.Node) bool {
			return slices.Contains(g.opts.Inputs, node.Loc.File)
		}
	}

	if g.opts.FilterEnum == nil {
		g.opts.FilterEnum = func(node *c.Node) bool {
			return slices.Contains(g.opts.Inputs, node.Loc.File)
		}
	}

	if g.opts.FilterFunc == nil {
		g.opts.FilterFunc = func(node *c.Node) bool {
			return slices.Contains(g.opts.Inputs, node.Loc.File)
		}
	}

	// Parse type
	if g.opts.ParseType == nil {
		g.opts.ParseType = func(str string) fb.Type {
			return nil
		}
	}

	// Transforms
	transformAlias := g.opts.TransformAlias
	transformStruct := g.opts.TransformStruct
	transformEnum := g.opts.TransformEnum
	transformFunc := g.opts.TransformFunc

	g.opts.TransformAlias = func(a *fb.Alias) {
		a.Name = g.MapName(a.Name)

		if transformAlias != nil {
			transformAlias(a)
		}
	}

	g.opts.TransformStruct = func(s *fb.Struct) {
		s.Name = g.MapName(s.Name)

		if transformStruct != nil {
			transformStruct(s)
		}
	}

	g.opts.TransformEnum = func(e *fb.Enum) {
		e.Name = g.MapName(e.Name)

		if transformEnum != nil {
			transformEnum(e)
		}

		for _, cas := range e.Cases {
			if cas.Name[0] >= '0' && cas.Name[0] <= '9' {
				cas.Name = e.Name + cas.Name
			}
		}
	}

	g.opts.TransformFunc = func(f *fb.Func) {
		f.Name = g.MapName(f.Name)

		if transformFunc != nil {
			transformFunc(f)
		}
	}

	if g.opts.TransformFunc == nil {
		g.opts.TransformFunc = func(f *fb.Func) {}
	}

	// Order
	if g.opts.Order == nil {
		g.opts.Order = func(a, b fb.Decl) int {
			aOrder := GetDeclOrder(a)
			bOrder := GetDeclOrder(b)

			return cmp.Compare(aOrder, bOrder)
		}
	}
}

func (g *generator) MapName(name string) string {
	if g.opts.NameMappings != nil {
		if newName, ok := g.opts.NameMappings[name]; ok {
			return newName
		}
	}

	return name
}

func (g *generator) ParseType(str string) fb.Type {
	typ := ParseCType(str, func(str string) fb.Type {
		// Declaration
		for _, pair := range g.decls {
			decl := pair.decl

			if decl.Name_() == str {
				return &fb.DeclType{Decl: decl}
			}
		}

		// Custom
		return g.opts.ParseType(str)
	})

	if typ == nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to parse type: '%s'\n", str)
		os.Exit(1)
	}

	return typ
}

func (m MacroEnum) contains(name string) bool {
	if len(m.Exact) > 0 {
		return slices.Contains(m.Exact, name)
	}

	if m.Prefix != "" && !strings.HasPrefix(name, m.Prefix) {
		return false
	}
	if m.Suffix != "" && !strings.HasSuffix(name, m.Suffix) {
		return false
	}

	return true
}

func (m MacroEnum) trim(name string) string {
	if m.Prefix != "" {
		name = strings.TrimPrefix(name, m.Prefix)
	}
	if m.Suffix != "" {
		name = strings.TrimSuffix(name, m.Suffix)
	}

	return name
}

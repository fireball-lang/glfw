package fb

import (
	"fmt"
	"io"
	"strings"
)

type Decl interface {
	OutputIndex_() int
	Name_() string

	Write(w io.Writer)
}

// Alias

type Alias struct {
	OutputIndex int

	Documentation string

	Name string
	Type Type
}

func (a *Alias) OutputIndex_() int {
	return a.OutputIndex
}

func (a *Alias) Name_() string {
	return a.Name
}

func (a *Alias) Write(w io.Writer) {
	WriteDocumentation(w, a.Documentation, "")

	_, _ = fmt.Fprintf(w, "pub type %s = ", a.Name)
	a.Type.Write(w)
	_, _ = fmt.Fprint(w, ";\n")
}

// Enum

type Case struct {
	Documentation string

	Name  string
	Value string
}

type Enum struct {
	OutputIndex int

	Documentation string

	Name string
	Type Type

	Cases []*Case

	Bitfield bool
}

func (e *Enum) Case(name string) *Case {
	for _, cas := range e.Cases {
		if cas.Name == name {
			return cas
		}
	}

	return nil
}

func (e *Enum) OutputIndex_() int {
	return e.OutputIndex
}

func (e *Enum) Name_() string {
	return e.Name
}

func (e *Enum) Write(w io.Writer) {
	WriteDocumentation(w, e.Documentation, "")

	_, _ = fmt.Fprintf(w, "pub enum %s", e.Name)

	if e.Type != nil {
		_, _ = fmt.Fprint(w, " : ")
		e.Type.Write(w)
	}

	_, _ = fmt.Fprint(w, " {\n")

	for _, cas := range e.Cases {
		WriteDocumentation(w, cas.Documentation, "    ")

		if cas.Value == "" {
			_, _ = fmt.Fprintf(w, "    %s,\n", cas.Name)
		} else {
			_, _ = fmt.Fprintf(w, "    %s = %s,\n", cas.Name, cas.Value)
		}
	}

	_, _ = fmt.Fprint(w, "}\n")

	// Bitfield
	if e.Bitfield {
		var underlying strings.Builder
		e.Type.Write(&underlying)

		_, _ = fmt.Fprintf(w, `
impl %[1]s : BitNot {
    type Result = %[1]s;

    pub func bit_not(self) Self {
        return ~(*self as %[2]s) as Self;
    }
}

impl %[1]s : BitOr[%[1]s] {
    type Result = %[1]s;

    pub func bit_or(self, rhs: Self) Self {
        return (*self as %[2]s | rhs as %[2]s) as Self;
    }
}

impl %[1]s : BitAnd[%[1]s] {
    type Result = %[1]s;

    pub func bit_and(self, rhs: Self) Self {
        return (*self as %[2]s & rhs as %[2]s) as Self;
    }
}
`, e.Name, underlying.String())
	}
}

// Struct

type Field struct {
	Documentation string

	Name string
	Type Type
}

type Struct struct {
	OutputIndex int

	Documentation string

	Name   string
	Fields []*Field

	Union bool
}

func (s *Struct) OutputIndex_() int {
	return s.OutputIndex
}

func (s *Struct) Name_() string {
	return s.Name
}

func (s *Struct) Write(w io.Writer) {
	WriteDocumentation(w, s.Documentation, "")

	if s.Union {
		_, _ = fmt.Fprint(w, "#[repr(Union)]\n")
	} else {
		_, _ = fmt.Fprint(w, "#[repr(C)]\n")
	}

	if len(s.Fields) == 0 {
		_, _ = fmt.Fprintf(w, "pub struct %s {}\n", s.Name)
		return
	}

	_, _ = fmt.Fprintf(w, "pub struct %s {\n", s.Name)

	for _, field := range s.Fields {
		WriteDocumentation(w, field.Documentation, "    ")

		_, _ = fmt.Fprintf(w, "    pub %s: ", field.Name)
		field.Type.Write(w)
		_, _ = fmt.Fprint(w, ",\n")
	}

	_, _ = fmt.Fprint(w, "}\n")
}

// Func

type Param struct {
	Name string
	Type Type
}

type Func struct {
	OutputIndex int

	Documentation string

	Name     string
	LinkName string

	Params  []*Param
	Returns Type

	MethodName    string
	ReceiverIndex int
}

func (f *Func) OutputIndex_() int {
	return f.OutputIndex
}

func (f *Func) Name_() string {
	return f.Name
}

func (f *Func) Write(w io.Writer) {
	WriteDocumentation(w, f.Documentation, "")

	// Attributes
	_, _ = fmt.Fprint(w, "#[extern")

	if f.Name == f.LinkName {
		_, _ = fmt.Fprint(w, "]\n")
	} else {
		_, _ = fmt.Fprintf(w, ", link_name(\"%s\")]\n", f.LinkName)
	}

	// Signature
	_, _ = fmt.Fprintf(w, "pub func %s(", f.Name)

	for i, param := range f.Params {
		if i > 0 {
			_, _ = fmt.Fprint(w, ", ")
		}

		_, _ = fmt.Fprintf(w, "%s: ", param.Name)
		param.Type.Write(w)
	}

	_, _ = fmt.Fprint(w, ")")

	// Returns
	if s, ok := f.Returns.(*SimpleType); !ok || s.Text != "void" {
		_, _ = fmt.Fprint(w, " ")
		f.Returns.Write(w)
	}

	_, _ = fmt.Fprint(w, ";\n")
}

// utils

func WriteDocumentation(w io.Writer, docs string, indent string) {
	for line := range strings.Lines(docs) {
		line = strings.TrimSpace(line)

		if indent != "" {
			_, _ = fmt.Fprint(w, indent)
		}

		if line == "" {
			_, _ = fmt.Fprint(w, "///\n")
			continue
		}

		_, _ = fmt.Fprintf(w, "/// %s\n", line)
	}
}

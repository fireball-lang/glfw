package fb

import (
	"fmt"
	"io"
)

type Type interface {
	isType()

	Write(w io.Writer)
}

// Simple

type SimpleType struct {
	Text string
}

func (s *SimpleType) isType() {}

func (s *SimpleType) Write(w io.Writer) {
	_, _ = fmt.Fprint(w, s.Text)
}

// DeclType

type DeclType struct {
	Decl Decl
}

func (d *DeclType) isType() {}

func (d *DeclType) Write(w io.Writer) {
	_, _ = fmt.Fprint(w, d.Decl.Name_())
}

// ArrayType

type ArrayType struct {
	Size    uint32
	Element Type
}

func (a *ArrayType) isType() {}

func (a *ArrayType) Write(w io.Writer) {
	_, _ = fmt.Fprintf(w, "[%d]", a.Size)
	a.Element.Write(w)
}

// PointerType

type PointerType struct {
	Mutable bool
	Pointee Type
}

func (p *PointerType) isType() {}

func (p *PointerType) Write(w io.Writer) {
	if p.Mutable {
		_, _ = fmt.Fprint(w, "mut ")
	}

	_, _ = fmt.Fprint(w, "*")
	p.Pointee.Write(w)
}

// FuncType

type FuncType struct {
	Params  []Param
	Returns Type
}

func (f *FuncType) isType() {}

func (f *FuncType) Write(w io.Writer) {
	_, _ = fmt.Fprint(w, "func(")

	for i, param := range f.Params {
		if i > 0 {
			_, _ = fmt.Fprint(w, ", ")
		}

		if param.Name != "" {
			_, _ = fmt.Fprintf(w, "%s: ", param.Name)
		}

		param.Type.Write(w)
	}

	_, _ = fmt.Fprint(w, ")")

	if s, ok := f.Returns.(*SimpleType); !ok || s.Text != "void" {
		_, _ = fmt.Fprint(w, " ")
		f.Returns.Write(w)
	}
}

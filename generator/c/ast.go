package c

import "strings"

type Node struct {
	Kind string `json:"kind"`

	Loc Location `json:"loc"`

	Name string `json:"name"`

	Tag string `json:"tagUsed"`

	Type Type `json:"type"`

	Value string `json:"value"`

	Text      string `json:"text"`
	GroupText string `json:"-"`

	Args []string `json:"args"`

	Inner []*Node `json:"inner"`
}

type Location struct {
	File         string  `json:"file"`
	IncludedFrom Include `json:"includedFrom"`
}

type Include struct {
	File string `json:"file"`
}

type Type struct {
	QualType string `json:"qualType"`
}

func (n *Node) EnumValue() string {
	if n.Value != "" {
		return n.Value
	}

	for _, child := range n.Inner {
		if value := child.EnumValue(); value != "" {
			return value
		}
	}

	return ""
}

func (n *Node) Documentation() string {
	if n.Kind == "MacroDefine" && n.Text != "" {
		return n.Text
	}

	var sb strings.Builder

	for _, node := range n.Inner {
		if node.Kind == "FullComment" {
			extractDocumentationText(&sb, node)
		}
	}

	return strings.TrimSpace(sb.String())
}

func extractDocumentationText(sb *strings.Builder, node *Node) {
	switch node.Kind {
	case "TextComment":
		sb.WriteString(node.Text)

	case "BlockCommandComment":
		if sb.Len() > 0 && !strings.HasSuffix(sb.String(), "\n") {
			sb.WriteRune('\n')
		}

		sb.WriteRune('@')
		sb.WriteString(node.Name)

		if len(node.Args) > 0 {
			sb.WriteRune(' ')
			sb.WriteString(strings.Join(node.Args, " "))
		}

	case "InlineCommandComment":
		sb.WriteRune('@')
		sb.WriteString(node.Name)

		if len(node.Args) > 0 {
			sb.WriteRune(' ')
			sb.WriteString(strings.Join(node.Args, " "))
		}
	}

	for _, child := range node.Inner {
		extractDocumentationText(sb, child)
	}

	if node.Kind == "ParagraphComment" {
		if sb.Len() > 0 && !strings.HasSuffix(sb.String(), "\n") {
			sb.WriteRune('\n')
		}
	}
}

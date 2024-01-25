package sitterutils

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

func Format(node *sitter.Node, depth int) string {
	if node.ChildCount() == 0 {
		return "<" + node.Type() + " />\n"
	}

	text := "<" + node.Type() + ">\n  "
	for i := 0; i < int(node.ChildCount()); i++ {
		fieldName := node.FieldNameForChild(i)

		inner := Format(node.Child(i), depth+1)
		if fieldName != "" {
			inner = fieldName + ": " + inner
		}

		text += strings.ReplaceAll(inner, "\n", "\n  ")
	}
	text = strings.TrimRight(text, " ")
	text += "</" + node.Type() + ">\n"

	return text
}

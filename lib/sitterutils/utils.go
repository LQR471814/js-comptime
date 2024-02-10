package sitterutils

import (
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

func Format(node *sitter.Node, depth int, onlyNamed bool) string {
	if node.ChildCount() == 0 {
		return "<" + node.Type() + " />\n"
	}

	text := "<" + node.Type() + ">\n  "
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if onlyNamed && !child.IsNamed() {
			continue
		}

		fieldName := node.FieldNameForChild(i)
		inner := Format(child, depth+1, onlyNamed)
		if fieldName != "" {
			inner = fmt.Sprintf("%d|%s: %s", i, fieldName, inner)
		} else {
			inner = fmt.Sprintf("%d: %s", i, inner)
		}

		text += strings.ReplaceAll(inner, "\n", "\n  ")
	}
	text = strings.TrimRight(text, " ")
	text += "</" + node.Type() + ">\n"

	return text
}

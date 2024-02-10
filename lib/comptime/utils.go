package comptime

import (
	sitter "github.com/smacker/go-tree-sitter"
)

const COMPTIME_KEYWORD = "$comptime"

// node must be "array_pattern" | "identifier" | "object_pattern"
func getDeclaredVars(node *sitter.Node, buff []byte) []string {
	if node == nil {
		return nil
	}

	declared := []string{}
	switch node.Type() {
	case "identifier":
		declared = append(declared, node.Content(buff))
	case "shorthand_property_identifier_pattern":
		declared = append(declared, node.Content(buff))
	case "pair_pattern":
		inner := getDeclaredVars(node.ChildByFieldName("value"), buff)
		declared = append(declared, inner...)
	case "object_pattern", "array_pattern":
		for i := 0; i < int(node.ChildCount()); i++ {
			inner := getDeclaredVars(node.Child(i), buff)
			declared = append(declared, inner...)
		}
	}
	return declared
}

func parseLexicalDecl(node *sitter.Node, buff []byte) []string {
	// this assumes lexical_declaration
	declarations := []string{}

	for declIdx := 0; declIdx < int(node.ChildCount()); declIdx++ {
		decl := node.Child(declIdx)
		if decl.Type() != "variable_declarator" {
			continue
		}

		vars := getDeclaredVars(decl.ChildByFieldName("name"), buff)
		declarations = append(declarations, vars...)
	}

	return declarations
}

func resolve(id string, scope *Scope) bool {
	for scope != nil {
		for _, otherId := range scope.RuntimeDeclarations {
			if otherId == id {
				return false
			}
		}
		for _, decl := range scope.ComptimeDeclarations {
			for _, otherId := range decl.Identifiers {
				if otherId == id {
					return true
				}
			}
		}
		scope = scope.Parent
	}
	return false
}

func stringFromId(source []byte, node *sitter.Node) string {
	if node.Type() == "identifier" {
		return node.Content(source)
	}
	return ""
}

// returns the defined identifiers
func getDefinedIdentifiers(node *sitter.Node, source []byte) []string {
	switch node.Type() {
	case "class_declaration",
		"function_declaration",
		"generator_function_declaration":
		name := stringFromId(source, node.ChildByFieldName("name"))
		return []string{name}
	// variable declarations
	case "lexical_declaration", "variable_declaration":
		return parseLexicalDecl(node, source)
	case "import_clause":
		// todo
	}
	return nil
}

func getParameterIdentifiers(node *sitter.Node, source []byte) []string {
	singleParam := node.ChildByFieldName("parameter")
	if singleParam != nil {
		return []string{singleParam.Content(source)}
	} else {
		params := node.ChildByFieldName("parameters")
		var ids []string
		for i := 0; i < int(params.NamedChildCount()); i++ {
			ids = append(ids, getDeclaredVars(node, source)...)
		}
		return ids
	}
}

package comptime

import sitter "github.com/smacker/go-tree-sitter"

// node is the "name" of variable_declarator
func getDeclaredVars(node *sitter.Node, buff []byte) []string {
	if node == nil {
		return nil
	}

	declared := []string{}
	switch node.Type() {
	case "identifier":
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

// var constants = map[string]struct{}{
// 	// primitives
// 	"number":    {},
// 	"string":    {},
// 	"null":      {},
// 	"undefined": {},
// 	"true":      {},
// 	"false":     {},

// 	// assignment operators
// 	"=":    {},
// 	"+=":   {},
// 	"-=":   {},
// 	"*=":   {},
// 	"/=":   {},
// 	"%=":   {},
// 	"**=":  {},
// 	"++":   {},
// 	"--":   {},
// 	"<<=":  {},
// 	">>=":  {},
// 	">>>=": {},
// 	"&=":   {},
// 	"^=":   {},
// 	"|=":   {},
// 	"&&=":  {},
// 	"||=":  {},
// 	"??=":  {},

// 	// comparison operators
// 	"==":  {},
// 	"!=":  {},
// 	"===": {},
// 	"!==": {},
// 	">":   {},
// 	"<":   {},
// 	">=":  {},
// 	"<=":  {},

// 	// logical operators
// 	"&&": {},
// 	"||": {},
// 	"!":  {},

// 	// arithmetic operators
// 	"+":  {},
// 	"-":  {},
// 	"*":  {},
// 	"/":  {},
// 	"%":  {},
// 	"**": {},

// 	// bitwise operators
// 	"&":   {},
// 	"|":   {},
// 	"^":   {},
// 	"~":   {},
// 	"<<":  {},
// 	">>":  {},
// 	">>>": {},

// 	// misc operators
// 	"ternary_expression": {},
// 	"unary_expression":   {},
// }

// func isConstant(node *sitter.Node) bool {
// 	_, isConstant := constants[node.Type()]
// 	return isConstant
// }

func isConstant(node *sitter.Node) bool {
	switch node.Type() {
	case "number", "string", "null", "undefined", "true", "false":
		return true
	}
	return false
}

package main

import (
	"context"
	"fmt"
	"io"
	sitterutils "js-comptime/lib/sitter-utils"
	"log"
	"os"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/javascript"
)

type comptimeNode interface {
	comptimeNode()
}

type comptimeVarDeclaration struct {
	text         string
	declarations []string
}

type comptimeStatement struct {
	text string
}

type comptimeScope struct {
	nodes []comptimeNode
}

func (comptimeStatement) comptimeNode()      {}
func (comptimeScope) comptimeNode()          {}
func (comptimeVarDeclaration) comptimeNode() {}

type traversalContext struct {
	inComptime bool
	source     []byte
}

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

func traverse(node *sitter.Node, ctx traversalContext) comptimeNode {
	fmt.Println("---------")
	fmt.Println(node.Type(), node.Content(ctx.source))

	if node.Type() == "labeled_statement" {
		label := node.ChildByFieldName("label").Content(ctx.source)
		body := node.ChildByFieldName("body")
		if label == "$comptime" && !ctx.inComptime {
			ctx.inComptime = true

			switch body.Type() {
			case "lexical_declaration":
				return comptimeVarDeclaration{
					declarations: parseLexicalDecl(body, ctx.source),
					text:         body.Content(ctx.source),
				}
			case "function_declaration":
				return comptimeVarDeclaration{
					declarations: []string{body.ChildByFieldName("name").Content(ctx.source)},
					text:         body.Content(ctx.source),
				}
			}

			return comptimeStatement{text: body.Content(ctx.source)}
		}
	}

	nodes := []comptimeNode{}
	for i := 0; i < int(node.ChildCount()); i++ {
		rendered := traverse(node.Child(i), ctx)
		switch typedNode := rendered.(type) {
		case comptimeScope:
			if len(typedNode.nodes) == 0 {
				continue
			}
		}
		nodes = append(nodes, rendered)
	}

	return comptimeScope{nodes: nodes}
}

func renderComptime(n comptimeNode) string {
	text := ""
	switch typedNode := n.(type) {
	case comptimeVarDeclaration:
		text += typedNode.text
	case comptimeStatement:
		text = typedNode.text
	case comptimeScope:
		text = "{"
		for i, child := range typedNode.nodes {
			text += renderComptime(child)
			if i < len(typedNode.nodes)-1 {
				text += ";\n"
			}
		}
		text += "}"
	}
	return text
}

/*
a comptime value can only be brought into runtime code in only 2 ways:
- a comptime variable is referenced directly (by identifier)
- a comptime function is called
*/

func main() {
	buff, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	parser := sitter.NewParser()
	parser.SetLanguage(javascript.GetLanguage())
	tree, err := parser.ParseCtx(context.Background(), nil, buff)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(sitterutils.Format(tree.RootNode(), 0))

	rendered := traverse(tree.RootNode(), traversalContext{
		source: buff,
	})
	fmt.Println("===============")
	fmt.Println(renderComptime(rendered))
}

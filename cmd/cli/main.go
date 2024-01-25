package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/javascript"
)

type comptimeNode interface {
	comptimeNode()
}

type comptimeStatement struct {
	text string
}

type comptimeScope struct {
	nodes []comptimeNode
}

func (comptimeStatement) comptimeNode() {}
func (comptimeScope) comptimeNode()     {}

// a comptime value can be referenced in the following ways:
// value is 
// variable access
type comptimeRef struct {
	expr *sitter.Node
}

type traversalContext struct {
	inComptime bool
	source     []byte
	refs       *[]comptimeRef
}

func traverse(node *sitter.Node, ctx traversalContext) comptimeNode {
	// if node.ChildCount() != 0 {
	//   fmt.Println("---------")
	//   fmt.Println(node.Type(), node.Content(ctx.source))
	// }

	if node.Type() == "labeled_statement" {
		label := node.ChildByFieldName("label").Content(ctx.source)
		body := node.ChildByFieldName("body")
		if label == "$comptime" && !ctx.inComptime {
			ctx.inComptime = true
			return comptimeStatement{text: body.Content(ctx.source)}
		}
	}

	nodes := []comptimeNode{}
	for i := 0; i < int(node.ChildCount()); i++ {
		rendered := traverse(node.Child(i), ctx)
		switch rendered.(type) {
		case comptimeScope:
			if len(rendered.(comptimeScope).nodes) == 0 {
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

	rendered := traverse(tree.RootNode(), traversalContext{
		source: buff,
	})
	fmt.Println(renderComptime(rendered))
}

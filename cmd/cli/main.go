package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/javascript"
	v8 "rogchap.com/v8go"
)

func main() {
	jsctx := v8.NewContext()

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

	iterator := sitter.NewIterator(tree.RootNode(), sitter.DFSMode)
	err = iterator.ForEach(func(n *sitter.Node) error {
		if n.ChildCount() == 0 {
			return nil
		}

		// fmt.Println("---------")
		// fmt.Println(n.Type(), n.Content(buff))

		if n.Type() == "labeled_statement" {
			label := n.ChildByFieldName("label").Content(buff)
			body := n.ChildByFieldName("body")
			if label == "$comptime" {
				fmt.Println("COMPTIME ---------")
				fmt.Println(body.Type(), body.Content(buff))

				switch body.Type() {
				case "expression_statement":
					// TODO: add more meaningful errors later
					_, err := jsctx.RunScript(body.Content(buff), "<comptime code>")
					if err != nil {
						return err
					}
				}
			}
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}

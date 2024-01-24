package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/javascript"
)

func main() {
	// jsctx := v8.NewContext()

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

	runBuff := bytes.NewBuffer(nil)

	iterator := sitter.NewIterator(tree.RootNode(), sitter.DFSMode)
	iterator.ForEach(func(n *sitter.Node) error {
		if n.ChildCount() == 0 {
			return nil
		}

		// fmt.Println("---------")
		// fmt.Println(n.Type(), n.Content(buff))

		if n.Type() == "labeled_statement" {
			label := n.ChildByFieldName("label").Content(buff)
			body := n.ChildByFieldName("body")
			if label == "$comptime" {
				bodyType := body.Type()
				bodyContent := body.Content(buff)

				// fmt.Println("COMPTIME ---------")
				// fmt.Println(bodyType, bodyContent)

				switch bodyType {
				case "expression_statement":
					runBuff.Write([]byte(bodyContent + "\n"))
					// // TODO: add more meaningful errors later
					// _, err := jsctx.RunScript(bodyContent, "<comptime code>")
					// if err != nil {
					// 	return err
					// }
				}
			}
		}

		return nil
	})

	fmt.Println(runBuff.String())
}

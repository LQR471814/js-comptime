package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	sitterutils "jscomptime/lib/sitterutils"
	"log"
	"os"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/javascript"
)

func main() {
	buff, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	onlyNamed := flag.Bool("only-named", false, "Only show named nodes.")
	flag.Parse()

	parser := sitter.NewParser()
	parser.SetLanguage(javascript.GetLanguage())
	tree, err := parser.ParseCtx(context.Background(), nil, buff)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(sitterutils.Format(tree.RootNode(), 0, *onlyNamed))
}

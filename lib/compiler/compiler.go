package compiler

import (
	"context"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/javascript"
)

type Compiler struct {
	source []byte
	tree   *sitter.Tree
}

func NewCompiler(source []byte) (Compiler, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(javascript.GetLanguage())
	tree, err := parser.ParseCtx(context.Background(), nil, source)
	if err != nil {
		return Compiler{}, err
	}
	return Compiler{
		source: source,
		tree:   tree,
	}, nil
}

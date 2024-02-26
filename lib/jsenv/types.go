package jsenv

import (
	"context"

	sitter "github.com/smacker/go-tree-sitter"
)

type Eval struct {
	Node   *sitter.Node
	Result string
}

type Env interface {
	Eval(ctx context.Context, code string, results []Eval) error
}


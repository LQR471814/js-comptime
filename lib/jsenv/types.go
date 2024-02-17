package jsenv

import (
	"context"

	sitter "github.com/smacker/go-tree-sitter"
)

type EvalResult struct {
	Node   *sitter.Node
	Result string
}

type Env interface {
	Eval(ctx context.Context, code string, results []EvalResult) error
}


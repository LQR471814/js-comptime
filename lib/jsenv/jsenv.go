package jsenv

import sitter "github.com/smacker/go-tree-sitter"

type EvalResult struct {
	Node   *sitter.Node
	Result string
}

type Env interface {
	Eval(code string, results []EvalResult) error
}

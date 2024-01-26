package comptime

import (
	"context"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/javascript"
)

type StatementType = uint8

const (
	STATEMENT_REGULAR StatementType = iota
	STATEMENT_DECLARATION
	STATEMENT_REGION
	STATEMENT_SCOPE
)

type RegularStatement struct {
	Text string
}

type VarDeclarations struct {
	Identifiers []string
	Text        string
}

type Region struct {
	// these only refer to comptime dependencies
	Dependencies []string
	Text         string
}

type StatementRef struct {
	Type  StatementType
	Index int
}

type Scope struct {
	Parent        *Scope
	Children      []Scope
	StatementList []StatementRef

	Parameters   []string
	Regular      []RegularStatement
	Declarations []VarDeclarations
	Regions      []Region
}

func recurse(n *sitter.Node, scope *Scope, source []byte) *Region {
	isExpression := false
	nodeType := n.Type()
	if nodeType == "labeled_statement" {
		// collect "comptime values"
		label := n.ChildByFieldName("label").Content(source)
		body := n.ChildByFieldName("body")
		if label == "$comptime" {
			content := body.Content(source)

			switch body.Type() {
			case "lexical_declaration":
				scope.Declarations = append(scope.Declarations, VarDeclarations{
					Identifiers: parseLexicalDecl(body, source),
					Text:        content,
				})
				scope.StatementList = append(scope.StatementList, StatementRef{
					Type:  STATEMENT_DECLARATION,
					Index: len(scope.Declarations) - 1,
				})
			case "function_declaration":
				scope.Declarations = append(scope.Declarations, VarDeclarations{
					Identifiers: []string{body.ChildByFieldName("name").Content(source)},
					Text:        content,
				})
				scope.StatementList = append(scope.StatementList, StatementRef{
					Type:  STATEMENT_DECLARATION,
					Index: len(scope.Declarations) - 1,
				})
			default:
				scope.Regular = append(scope.Regular, RegularStatement{
					Text: content,
				})
				scope.StatementList = append(scope.StatementList, StatementRef{
					Type:  STATEMENT_REGULAR,
					Index: len(scope.Declarations) - 1,
				})
			}
		}
	} else if strings.HasSuffix(nodeType, "expression") {
		isExpression = true
	}

	constantChildren := true
	for i := 0; i < int(n.ChildCount()); i++ {
		childScope := scope
		childNode := n.Child(i)

		switch childNode.Type() {
		case "statement_block":
			childScope = &Scope{
				Parent: scope,
			}
		case "function_declaration":
			params := []string{}
			for i := 0; i < int(childNode.ChildCount()); i++ {
				param := childNode.Child(i)
				if param.Type() == "identifier" {
					params = append(params, param.Content(source))
				}
			}
			childScope = &Scope{
				Parent:     scope,
				Parameters: params,
			}
		}

		comptimeRegion := recurse(childNode, childScope, source)
		if comptimeRegion != nil {

		} else if !isConstant(childNode) {
			constantChildren = false
		}
	}

	if constantChildren && isExpression {
		return &Region{}
	}

	return nil
}

func Compile(source []byte) error {
	parser := sitter.NewParser()
	parser.SetLanguage(javascript.GetLanguage())
	tree, err := parser.ParseCtx(context.Background(), nil, source)
	if err != nil {
		return err
	}

	root := &Scope{}
	recurse(tree.RootNode(), root, source)
	return nil
}

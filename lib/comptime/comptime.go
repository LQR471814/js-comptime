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
	Text *sitter.Node
}

type VarDeclarations struct {
	Identifiers []string
	Text        *sitter.Node
}

type Region struct {
	// these only refer to comptime dependencies
	Dependencies []string
	Node         *sitter.Node
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
  if isConstant(n) {
    
  }

	isExpression := false
	nodeType := n.Type()
	if nodeType == "labeled_statement" {
		// collect "comptime values"
		label := n.ChildByFieldName("label").Content(source)
		body := n.ChildByFieldName("body")
		if label == "$comptime" {
			switch body.Type() {
			case "lexical_declaration":
				scope.Declarations = append(scope.Declarations, VarDeclarations{
					Identifiers: parseLexicalDecl(body, source),
					Text:        body,
				})
				scope.StatementList = append(scope.StatementList, StatementRef{
					Type:  STATEMENT_DECLARATION,
					Index: len(scope.Declarations) - 1,
				})
			case "function_declaration":
				scope.Declarations = append(scope.Declarations, VarDeclarations{
					Identifiers: []string{body.ChildByFieldName("name").Content(source)},
					Text:        body,
				})
				scope.StatementList = append(scope.StatementList, StatementRef{
					Type:  STATEMENT_DECLARATION,
					Index: len(scope.Declarations) - 1,
				})
			default:
				scope.Regular = append(scope.Regular, RegularStatement{
					Text: body,
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

  comptimeRegions := []Region{}
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

		region := recurse(childNode, childScope, source)
		if region != nil {
      // on comptime region
      comptimeRegions = append(comptimeRegions, *region)
		} else if !isConstant(childNode) {
      // on non-constant non-comptime child
			constantChildren = false
		}
	}

  // boundary is reached
  if !constantChildren  {
    
  }

  // expression can be comptime evaluated
  

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

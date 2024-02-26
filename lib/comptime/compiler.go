package comptime

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"jscomptime/lib/jsenv"
	"os"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/javascript"
)

func handleComptimeBody(node *sitter.Node, scope *Scope, source []byte) {
	ids := definedIdentifiers(node, source)
	if len(ids) > 0 {
		scope.addComptimeDeclaration(VarDeclarations{
			Identifiers: ids,
			Node:        node,
		})
		return
	}
	scope.addComptimeStatement(node)
}

func recurse(node *sitter.Node, scope *Scope, source []byte) childType {
	nodeType := node.Type()
	childScope := scope

	// handle comptime labels
	if nodeType == "labeled_statement" {
		label := node.ChildByFieldName("label").Content(source)
		if label == COMPTIME_KEYWORD {
			comptimeNode := node.ChildByFieldName("body")
			handleComptimeBody(comptimeNode, scope, source)
			// don't handle comptime body children
			return type_comptime
		}
	}

	// handle runtime var declarations
	ids := definedIdentifiers(node, source)
	if len(ids) > 0 {
		scope.RuntimeDeclarations = append(scope.RuntimeDeclarations, ids...)
	} else {
		// handle scope
		switch nodeType {
		case "statement_block":
			childScope = &Scope{Parent: scope}
		case "generator_function",
			"generator_function_declaration",
			"function",
			"function_declaration",
			"arrow_function",
			"method_definition":
			childScope = &Scope{
				Parent:              scope,
				RuntimeDeclarations: getParameterIdentifiers(node, source),
			}
		}

		// handle regions (expressions that can be replaced with an evaluated comptime value)
		switch nodeType {
		case "expression_statement":
			return recurse(node.NamedChild(0), childScope, source)
		case "generator_function",
			"function",
			"arrow_function":
			return recurse(node.ChildByFieldName("body"), childScope, source)
		case "await_expression",
			"binary_expression",
			"new_expression",
			"ternary_expression",
			"array",
			"call_expression",
			"member_expression",
			"object",
			"parenthesized_expression",
			"subscript_expression",
			"template_string":
			comptimeExprs := []*sitter.Node{}
			hasRuntime := false

			// is comptime-able expression
			for i := 0; i < int(node.NamedChildCount()); i++ {
				child := node.NamedChild(i)
				childType := recurse(child, childScope, source)
				switch childType {
				case type_runtime:
					hasRuntime = true
				case type_comptime:
					comptimeExprs = append(comptimeExprs, child)
				}
			}

			if !hasRuntime {
				return type_comptime
			}
			for _, e := range comptimeExprs {
				scope.addRegion(e)
			}
			return type_runtime
		case "true",
			"false",
			"null",
			"number",
			"string",
			"regex",
			"undefined":
			return type_comptime
		case "identifier":
			id := node.Content(source)
			resolved := resolve(id, scope)
			if resolved {
				return type_comptime
			}
			return type_runtime
		}
	}

	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		childType := recurse(child, childScope, source)
		if childType == type_comptime {
			switch child.Type() {
			case "identifier",
				"generator_function",
				"function",
				"arrow_function",
				"await_expression",
				"binary_expression",
				"new_expression",
				"ternary_expression",
				"array",
				"call_expression",
				"member_expression",
				"object",
				"parenthesized_expression",
				"subscript_expression",
				"template_string":
				scope.addRegion(child)
			}
		}
	}

	if childScope != scope && len(childScope.DefinitionOrder) > 0 {
		scope.addScope(childScope)
	}

	return type_invalid
}

type comptimeResultType = uint8

const (
	ct_type_region comptimeResultType = iota
	ct_type_statement
)

type nodeRef struct {
	resultType comptimeResultType
	index      int
}

type comptimeResults struct {
	defOrder   []nodeRef
	statements []*sitter.Node
	regions    []jsenv.EvalResult
}

func renderComptimeCode(
	scope *Scope,
	source []byte,
	results *comptimeResults,
	out io.Writer,
) error {
	var err error
	for _, ref := range scope.DefinitionOrder {
		switch ref.Type {
		case DEF_SCOPE:
			_, err = out.Write([]byte("{\n"))
			if err != nil {
				return err
			}
			err = renderComptimeCode(scope.Scopes[ref.Index], source, results, out)
			if err != nil {
				return err
			}
			_, err = out.Write([]byte("}"))
			if err != nil {
				return err
			}
		case DEF_COMPTIME_STATEMENT:
			node := scope.ComptimeStatements[ref.Index]

			results.statements = append(results.statements, node.Parent())
			results.defOrder = append(results.defOrder, nodeRef{
				resultType: ct_type_statement,
				index:      len(results.statements) - 1,
			})

			_, err = out.Write([]byte(node.Content(source)))
			if err != nil {
				return err
			}
		case DEF_COMPTIME_DECLARATION:
			node := scope.ComptimeDeclarations[ref.Index].Node

			results.statements = append(results.statements, node.Parent())
			results.defOrder = append(results.defOrder, nodeRef{
				resultType: ct_type_statement,
				index:      len(results.statements) - 1,
			})

			_, err = out.Write([]byte(node.Content(source)))
			if err != nil {
				return err
			}
		case DEF_REGION:
			regionNode := scope.Regions[ref.Index]

			results.regions = append(results.regions, jsenv.EvalResult{
				Node: regionNode,
			})
			regionId := len(results.regions) - 1
			results.defOrder = append(results.defOrder, nodeRef{
				resultType: ct_type_region,
				index:      regionId,
			})

			export := fmt.Sprintf(
				"__jscomptime_export_value(%d, %s)",
				regionId, regionNode.Content(source),
			)
			if err != nil {
				return err
			}

			_, err = out.Write([]byte(export))
			if err != nil {
				return err
			}
		}
		_, err = out.Write([]byte("\n"))
		if err != nil {
			return err
		}
	}
	return nil
}

func Compile(ctx context.Context, source []byte, env jsenv.Env) (string, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(javascript.GetLanguage())
	tree, err := parser.ParseCtx(context.Background(), nil, source)
	if err != nil {
		return "", err
	}

	root := &Scope{}
	recurse(tree.RootNode(), root, source)

	// Debug info
	jsonScope := TransformToJSONScope(root, source)
	serialized, err := json.Marshal(jsonScope)
	if err != nil {
		return "", err
	}
	err = os.WriteFile("debug.json", serialized, 0777)
	if err != nil {
		return "", err
	}

	code := bytes.NewBuffer(nil)
	results := comptimeResults{}
	err = renderComptimeCode(root, source, &results, code)
	if err != nil {
		return "", err
	}

	err = env.Eval(ctx, code.String(), results.regions)
	if err != nil {
		return "", err
	}

	output := bytes.NewBuffer(nil)
	cursor := 0

	for _, res := range results.defOrder {
		switch res.resultType {
		case ct_type_region:
			region := results.regions[res.index]

			fmt.Printf("%s\t%s\n", region.Node.Content(source), region.Result)
			fmt.Println(cursor, region.Node.StartByte())

			_, err := output.Write(source[cursor:region.Node.StartByte()])
			if err != nil {
				return "", err
			}
			_, err = output.Write([]byte(region.Result))
			if err != nil {
				return "", err
			}
			cursor = int(region.Node.EndByte())
		case ct_type_statement:
			statement := results.statements[res.index]
			_, err := output.Write(source[cursor:statement.StartByte()])
			if err != nil {
				return "", err
			}
			cursor = int(statement.EndByte())
		}
	}
	_, err = output.Write(source[cursor:])
	if err != nil {
		return "", err
	}

	return output.String(), nil
}

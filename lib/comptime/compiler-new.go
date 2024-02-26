package comptime

import (
	"bytes"
	"context"
	"fmt"
	"jscomptime/lib/jsenv"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/javascript"
)

type comptimeCode struct {
	source []byte
	buff   *bytes.Buffer
}

func (c comptimeCode) addInlineExport(id int, node *sitter.Node) error {
	_, err := c.buff.WriteString(fmt.Sprintf(
		"__jscomptime_export_value(%d, %s)",
		id, node.Content(c.source),
	))
	return err
}
func (c comptimeCode) addStatement(node *sitter.Node) error {
	_, err := c.buff.Write(c.source[node.StartByte():node.EndByte()])
	return err
}
func (c comptimeCode) startBlock() error {
	_, err := c.buff.WriteString("{\n")
	return err
}
func (c comptimeCode) endBlock() error {
	_, err := c.buff.WriteString("}")
	return err
}

type expungeMod struct {
	node *sitter.Node
}

// if evalId < 0, the code will be expunged
type inlineMod struct {
	replace *sitter.Node
	evalId  int
}

type varScope struct {
	parent   *varScope
	comptime map[string]struct{}
	runtime  map[string]struct{}
}

func newVarScope(parent *varScope) *varScope {
	return &varScope{
		parent:   parent,
		comptime: map[string]struct{}{},
		runtime:  map[string]struct{}{},
	}
}

func (s varScope) addComptimeVars(ids []string) {
	for _, i := range ids {
		s.comptime[i] = struct{}{}
	}
}
func (s varScope) addRuntimeVars(ids []string) {
	for _, i := range ids {
		s.runtime[i] = struct{}{}
	}
}

// returns true if resolved comptime variable, false if resolved runtime or unresolved
func (s *varScope) resolve(id string) bool {
	current := s
	for current != nil {
		_, ok := current.runtime[id]
		if ok {
			return false
		}
		_, ok = current.comptime[id]
		if ok {
			return true
		}
		current = current.parent
	}
	return false
}

type evalResults struct {
	results []jsenv.Eval
}

func (r *evalResults) addEval(node *sitter.Node) int {
	newIdx := len(r.results)
	r.results = append(r.results, jsenv.Eval{
		Node: node,
	})
	return newIdx
}

type compileContext struct {
	source   []byte
	inline   *[]inlineMod
	comptime *comptimeCode
	evals    *evalResults
}

func (c compileContext) expungeCode(node *sitter.Node) {
	err := c.comptime.addStatement(node)
	if err != nil {
		panic(err)
	}
	*c.inline = append(*c.inline, inlineMod{replace: node, evalId: -1})
}

func (c compileContext) addComptimeExpr(node *sitter.Node) {
	inlineable := node
	switch node.Type() {
	case "spread_element", "computed_property_name":
		inlineable = node.NamedChild(0)
	}

	evalId := c.evals.addEval(inlineable)
	*c.inline = append(*c.inline, inlineMod{replace: inlineable, evalId: evalId})
	err := c.comptime.addInlineExport(evalId, inlineable)
	if err != nil {
		panic(err)
	}
}

// handles expressions with multiple child expressions
func handleCompositeExpression(c compileContext, scope *varScope, node *sitter.Node) bool {
	comptime := true
	comptimeNodes := make([]bool, node.NamedChildCount())
	for i := 0; i < int(node.NamedChildCount()); i++ {
		child := node.NamedChild(i)
		comptime := compile(c, scope, child)
		if !comptime {
			comptime = false
		}
		comptimeNodes[i] = comptime
	}
	if comptime {
		return true
	}
	for i := 0; i < int(node.NamedChildCount()); i++ {
		if comptimeNodes[i] {
			c.addComptimeExpr(node.NamedChild(i))
		}
	}
	return false
}

// handles expressions with multiple child expressions, but with a list of children passed in
func handleCompositeExpressionList(c compileContext, scope *varScope, children []*sitter.Node) bool {
	comptime := true
	comptimeNodes := make([]bool, len(children))
	for i, node := range children {
		comptime := compile(c, scope, node)
		if !comptime {
			comptime = false
		}
		comptimeNodes[i] = comptime
	}
	if comptime {
		return true
	}
	for i, child := range children {
		if comptimeNodes[i] {
			c.addComptimeExpr(child)
		}
	}
	return false
}

func compile(
	c compileContext,
	scope *varScope,
	node *sitter.Node,
) bool {
	nodeType := node.Type()

	switch nodeType {
	case "labeled_statement":
		// label: const x = 21
		label := node.ChildByFieldName("label").Content(c.source)
		body := node.ChildByFieldName("body")
		if label == COMPTIME_KEYWORD {
			scope.addComptimeVars(definedIdentifiers(body, c.source))
			c.expungeCode(node)
		}
	case "identifier", "shorthand_property_identifier":
		// some_variable
		return scope.resolve(node.Content(c.source))
	case "string", "true", "false", "number", "null", "undefined", "regex":
		// "1", true, false, 12, null, undefined
		return true
	// single child, and that child is an expression
	case "await_expression", "spread_element",
		"parenthesized_expression":
		// await 1, [1, 2]..., (1, 2)
		return compile(c, scope, node.NamedChild(0))
	// single child, not inlineable
	case "yield_expression":
		// yield "what"
		compile(c, scope, node.NamedChild(0))
		return false
	// some children are not expressions, some are
	case "update_expression", "unary_expression":
		// i++, typeof i
		return compile(c, scope, node.ChildByFieldName("argument"))
	case "binary_expression":
		// a + b
		return handleCompositeExpressionList(c, scope, []*sitter.Node{
			node.ChildByFieldName("left"),
			node.ChildByFieldName("right"),
		})
	case "call_expression":
		// func(1, 2)
		return handleCompositeExpressionList(c, scope, []*sitter.Node{
			node.ChildByFieldName("function"),
			node.ChildByFieldName("arguments"),
		})
	case "object":
		// { key: value }
		comptime := true
		comptimeNodes := make([]bool, node.NamedChildCount())
		pairResults := make([]struct {
			key   bool
			value bool
		}, node.NamedChildCount())

		for i := 0; i < int(node.NamedChildCount()); i++ {
			child := node.NamedChild(i)

			if child.Type() == "pair" {
				keyNode := node.ChildByFieldName("key")
				value := compile(c, scope, node.ChildByFieldName("value"))
				if keyNode.Type() != "computed_property_name" {
					pairResults[i] = struct {
						key   bool
						value bool
					}{
						key:   true,
						value: value,
					}
					comptimeNodes[i] = value
					continue
				}
				key := compile(c, scope, keyNode.NamedChild(0))
				pairResults[i] = struct {
					key   bool
					value bool
				}{
					key:   key,
					value: value,
				}
				comptimeNodes[i] = key && value
				continue
			}

			childComptime := compile(c, scope, child)
			if !childComptime {
				comptime = false
			}
			comptimeNodes[i] = childComptime
		}
		if comptime {
			return true
		}

		for i := 0; i < int(node.NamedChildCount()); i++ {
			child := node.NamedChild(i)

			if child.Type() == "pair" {
				keyNode := child.ChildByFieldName("key")
				valueNode := child.ChildByFieldName("value")
				result := pairResults[i]
				if result.key && keyNode.Type() == "computed_property_name" {
					c.addComptimeExpr(keyNode)
				}
				if result.value {
					c.addComptimeExpr(valueNode)
				}
				continue
			}

			if !comptimeNodes[i] {
				continue
			}

			c.addComptimeExpr(child)
		}
		return false
	// all children are expressions
	case
		// all fields: left, right
		"assignment_expression", "augmented_assignment_expression",
		"sequence_expression",
		// all children
		"arguments", "array",
		// all fields: constructor, arguments
		"new_expression",
		// all fields: consequence, alternative, condition
		"ternary_expression":
		return handleCompositeExpression(c, scope, node)
	}

	return false
}

type expansionCtx struct {
	source         []byte
	originalCursor *int
	newCursor      *int
	output         *bytes.Buffer
}

func expandShorthandProperties(source []byte, root *sitter.Node) []byte {
	buff := bytes.NewBuffer(make([]byte, 0, len(source)))
	originalCursor := 0
	newCursor := 0
	ctx := expansionCtx{
		source:         source,
		originalCursor: &originalCursor,
		newCursor:      &newCursor,
		output:         buff,
	}
	ctx.recurse(root)
	return buff.Bytes()
}

func (c expansionCtx) recurse(node *sitter.Node) {
	if node.Type() == "shorthand_property_identifier" {
		n1, err := c.output.Write(c.source[*c.originalCursor:node.StartByte()])
		if err != nil {
			panic(err)
		}

		id := node.Content(c.source)
		n2, err := c.output.Write([]byte(id + ": " + id))
		if err != nil {
			panic(err)
		}

		newCursor := *c.newCursor
		node.Edit(sitter.EditInput{
			StartIndex: uint32(newCursor + n1),
			OldEndIndex: uint32(node.EndByte()),
			NewEndIndex: uint32(newCursor + n1),
		})

		*c.newCursor += n1 + n2
		*c.originalCursor = int(node.EndByte())

		return
	}
	for i := 0; i < int(node.NamedChildCount()); i++ {
		c.recurse(node.Child(i))
	}
}

func CompileNew(ctx context.Context, source []byte, env jsenv.Env) (string, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(javascript.GetLanguage())

	tree, err := parser.ParseCtx(ctx, nil, source)
	if err != nil {
		return "", err
	}

	source = expandShorthandProperties(source, tree.RootNode())
	tree, err = parser.ParseCtx(ctx, tree, source)
	if err != nil {
		return "", err
	}

	mods := make([]inlineMod, 0, 1024)
	scope := newVarScope(nil)
	comptimeCode := comptimeCode{
		source: source,
		buff:   bytes.NewBuffer(make([]byte, 0, 2048)),
	}
	evals := make([]jsenv.Eval, 0, 1024)
	compileCtx := compileContext{
		source:   source,
		inline:   &mods,
		comptime: &comptimeCode,
		evals:    &evalResults{results: evals},
	}
	comptime := compile(compileCtx, scope, tree.RootNode())
	if comptime {
		compileCtx.addComptimeExpr(tree.RootNode())
	}

	return "", nil
}

package comptime

import (
	sitter "github.com/smacker/go-tree-sitter"
)

type DefinitionType = uint8

const (
	DEF_SCOPE DefinitionType = iota
	DEF_COMPTIME_STATEMENT
	DEF_COMPTIME_DECLARATION
	DEF_REGION
)

type VarDeclarations struct {
	Identifiers []string
	// this should include the entire variable declaration
	Node *sitter.Node
}

type StatementRef struct {
	Type  DefinitionType
	Index int
}

type Scope struct {
	// the parent scope
	Parent *Scope
	// a list of children scopes
	Scopes []*Scope
	// a list that keeps order in which statements, declarations,
	// and regions were defined
	DefinitionOrder []StatementRef
	// comptime statements
	ComptimeStatements []*sitter.Node
	// runtime variable declarations (to check if comptime variables
	// are shadowed by runtime declarations)
	RuntimeDeclarations []string
	// comptime variable declarations
	ComptimeDeclarations []VarDeclarations
	// expressions in which comptime variables are used
	Regions []*sitter.Node
}

func (s *Scope) addScope(scope *Scope) {
	s.Scopes = append(s.Scopes, scope)
	s.DefinitionOrder = append(s.DefinitionOrder, StatementRef{
		Type:  DEF_SCOPE,
		Index: len(s.Scopes) - 1,
	})
}

func (s *Scope) addComptimeStatement(statement *sitter.Node) {
	s.ComptimeStatements = append(s.ComptimeStatements, statement)
	s.DefinitionOrder = append(s.DefinitionOrder, StatementRef{
		Type:  DEF_COMPTIME_STATEMENT,
		Index: len(s.ComptimeStatements) - 1,
	})
}

func (s *Scope) addComptimeDeclaration(decl VarDeclarations) {
	s.ComptimeDeclarations = append(s.ComptimeDeclarations, decl)
	s.DefinitionOrder = append(s.DefinitionOrder, StatementRef{
		Type:  DEF_COMPTIME_DECLARATION,
		Index: len(s.ComptimeDeclarations) - 1,
	})
}

func (s *Scope) addRegion(r *sitter.Node) {
	s.Regions = append(s.Regions, r)
	s.DefinitionOrder = append(s.DefinitionOrder, StatementRef{
		Type:  DEF_REGION,
		Index: len(s.Regions) - 1,
	})
}

type childType = uint8

const (
	type_runtime childType = iota
	type_comptime
	type_invalid
)

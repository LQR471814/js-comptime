package comptime

type JSONVarDeclarations struct {
	Identifiers []string `json:"identifiers"`
	Text        string   `json:"text"`
}

type JSONStatementRef struct {
	Type  string `json:"type"`
	Index int    `json:"index"`
}

type JSONScope struct {
	Children        []JSONScope           `json:"children"`
	DefinitionOrder []JSONStatementRef    `json:"def_order"`
	Statements      []string              `json:"statements"`
	Declarations    []JSONVarDeclarations `json:"declarations"`
	Regions         []string              `json:"regions"`
}

func TransformToJSONScope(scope *Scope, source []byte) JSONScope {
	definitionOrder := make([]JSONStatementRef, len(scope.DefinitionOrder))
	for i, s := range scope.DefinitionOrder {
		t := ""
		switch s.Type {
		case DEF_COMPTIME_STATEMENT:
			t = "statement"
		case DEF_COMPTIME_DECLARATION:
			t = "declaration"
		case DEF_REGION:
			t = "region"
		}
		definitionOrder[i] = JSONStatementRef{
			Type:  t,
			Index: s.Index,
		}
	}

	statements := make([]string, len(scope.ComptimeStatements))
	for i, s := range scope.ComptimeStatements {
		statements[i] = s.Content(source)
	}

	declarations := make([]JSONVarDeclarations, len(scope.ComptimeDeclarations))
	for i, d := range scope.ComptimeDeclarations {
		declarations[i] = JSONVarDeclarations{
			Identifiers: d.Identifiers,
			Text:        d.Node.Content(source),
		}
	}

	regions := make([]string, len(scope.Regions))
	for i, r := range scope.Regions {
		regions[i] = r.Content(source)
	}

	children := make([]JSONScope, len(scope.Scopes))
	for i, c := range scope.Scopes {
		children[i] = TransformToJSONScope(c, source)
	}

	return JSONScope{
		Statements:      statements,
		Declarations:    declarations,
		Regions:         regions,
		DefinitionOrder: definitionOrder,
	}
}

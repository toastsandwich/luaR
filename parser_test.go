package luar

import (
	"testing"
)

func TestParser_ParseAssignment(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(*testing.T, *Program)
	}{
		{
			name:  "simple assignment",
			input: `x = 10`,
			check: func(t *testing.T, p *Program) {
				if len(p.Statements) != 1 {
					t.Fatalf("expected 1 statement, got %d", len(p.Statements))
				}
				stmt, ok := p.Statements[0].(*AssignmentStatement)
				if !ok {
					t.Fatal("expected AssignmentStatement")
				}
				if stmt.Names[0].Name != "x" {
					t.Errorf("expected name 'x', got %s", stmt.Names[0].Name)
				}
			},
		},
		{
			name:  "string assignment",
			input: `name = "hello"`,
			check: func(t *testing.T, p *Program) {
				stmt := p.Statements[0].(*AssignmentStatement)
				lit, ok := stmt.Values[0].(*StringLiteral)
				if !ok {
					t.Fatal("expected StringLiteral")
				}
				if lit.Value != "hello" {
					t.Errorf("expected 'hello', got %s", lit.Value)
				}
			},
		},
		{
			name:  "boolean assignment",
			input: `debug = true`,
			check: func(t *testing.T, p *Program) {
				stmt := p.Statements[0].(*AssignmentStatement)
				lit, ok := stmt.Values[0].(*BooleanLiteral)
				if !ok {
					t.Fatal("expected BooleanLiteral")
				}
				if !lit.Value {
					t.Error("expected true")
				}
			},
		},
		{
			name:  "float assignment",
			input: `pi = 3.14`,
			check: func(t *testing.T, p *Program) {
				stmt := p.Statements[0].(*AssignmentStatement)
				lit, ok := stmt.Values[0].(*NumberLiteral)
				if !ok {
					t.Fatal("expected NumberLiteral")
				}
				if lit.Value != 3.14 {
					t.Errorf("expected 3.14, got %f", lit.Value)
				}
			},
		},
		{
			name:  "nil assignment",
			input: `x = nil`,
			check: func(t *testing.T, p *Program) {
				stmt := p.Statements[0].(*AssignmentStatement)
				_, ok := stmt.Values[0].(*NilLiteral)
				if !ok {
					t.Fatal("expected NilLiteral")
				}
			},
		},
		{
			name:  "multiple assignments",
			input: `a = 1 b = 2 c = 3`,
			check: func(t *testing.T, p *Program) {
				if len(p.Statements) != 3 {
					t.Fatalf("expected 3 statements, got %d", len(p.Statements))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.input)
			p, err := parser.Parse()
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}
			tt.check(t, p)
		})
	}
}

func TestParser_ParseTable(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(*testing.T, *TableLiteral)
	}{
		{
			name:  "empty table",
			input: `t = {}`,
			check: func(t *testing.T, tbl *TableLiteral) {
				if len(tbl.Fields) != 0 {
					t.Errorf("expected 0 fields, got %d", len(tbl.Fields))
				}
			},
		},
		{
			name:  "table with values",
			input: `t = {1, 2, 3}`,
			check: func(t *testing.T, tbl *TableLiteral) {
				if len(tbl.Fields) != 3 {
					t.Fatalf("expected 3 fields, got %d", len(tbl.Fields))
				}
			},
		},
		{
			name:  "table with key-value",
			input: `t = {name = "test", count = 5}`,
			check: func(t *testing.T, tbl *TableLiteral) {
				if len(tbl.Fields) != 2 {
					t.Fatalf("expected 2 fields, got %d", len(tbl.Fields))
				}
				field := tbl.Fields[0]
				ident, ok := field.Key.(*Identifier)
				if !ok {
					t.Fatal("expected Identifier as key")
				}
				if ident.Name != "name" {
					t.Errorf("expected key 'name', got %s", ident.Name)
				}
			},
		},
		{
			name:  "nested table",
			input: `t = {inner = {a = 1}}`,
			check: func(t *testing.T, tbl *TableLiteral) {
				if len(tbl.Fields) != 1 {
					t.Fatalf("expected 1 field, got %d", len(tbl.Fields))
				}
				_, ok := tbl.Fields[0].Value.(*TableLiteral)
				if !ok {
					t.Error("expected nested TableLiteral")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.input)
			p, err := parser.Parse()
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}
			stmt := p.Statements[0].(*AssignmentStatement)
			tbl, ok := stmt.Values[0].(*TableLiteral)
			if !ok {
				t.Fatalf("expected TableLiteral, got %T", stmt.Values[0])
			}
			tt.check(t, tbl)
		})
	}
}

func TestParser_ParseIfStatement(t *testing.T) {
	input := `
if x == 1 then
    a = 10
elseif y == 2 then
    a = 20
else
    a = 0
end
`
	parser := NewParser(input)
	p, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(p.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(p.Statements))
	}

	ifStmt, ok := p.Statements[0].(*IfStatement)
	if !ok {
		t.Fatal("expected IfStatement")
	}

	if len(ifStmt.ElseIfs) != 1 {
		t.Errorf("expected 1 elseif, got %d", len(ifStmt.ElseIfs))
	}

	if len(ifStmt.Then) != 1 {
		t.Errorf("expected 1 statement in then block, got %d", len(ifStmt.Then))
	}

	if len(ifStmt.Else) != 1 {
		t.Errorf("expected 1 statement in else block, got %d", len(ifStmt.Else))
	}
}

func TestParser_ParseWhileStatement(t *testing.T) {
	input := `
while x < 10 do
    x = x + 1
end
`
	parser := NewParser(input)
	p, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	whileStmt, ok := p.Statements[0].(*WhileStatement)
	if !ok {
		t.Fatal("expected WhileStatement")
	}

	if len(whileStmt.Body) != 1 {
		t.Errorf("expected 1 statement in body, got %d", len(whileStmt.Body))
	}
}

func TestParser_ParseForStatement(t *testing.T) {
	input := `for i = 1, 10, 2 do
    x = i
end
`
	parser := NewParser(input)
	p, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	_, ok := p.Statements[0].(*ForStatement)
	if !ok {
		t.Error("expected ForStatement")
	}
}

func TestParser_ParseForInStatement(t *testing.T) {
	input := `for k, v in pairs(t) do
    print(k, v)
end
`
	parser := NewParser(input)
	p, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	forIn, ok := p.Statements[0].(*ForInStatement)
	if !ok {
		t.Error("expected ForInStatement")
	}

	if len(forIn.Names) != 2 {
		t.Errorf("expected 2 names, got %d", len(forIn.Names))
	}
}

func TestParser_ParseFunction(t *testing.T) {
	input := `function foo(a, b)
    return a + b
end
`
	parser := NewParser(input)
	p, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	funcStmt, ok := p.Statements[0].(*FunctionStatement)
	if !ok {
		t.Fatal("expected FunctionStatement")
	}

	if funcStmt.Name.Name.Name != "foo" {
		t.Errorf("expected function name 'foo', got %s", funcStmt.Name.Name.Name)
	}

	if len(funcStmt.Parameters) != 2 {
		t.Errorf("expected 2 parameters, got %d", len(funcStmt.Parameters))
	}
}

func TestParser_ParseLocalAssignment(t *testing.T) {
	input := `local x = 10
local y, z = 1, 2
`
	parser := NewParser(input)
	p, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(p.Statements) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(p.Statements))
	}

	localStmt, ok := p.Statements[0].(*LocalAssignmentStatement)
	if !ok {
		t.Fatal("expected LocalAssignmentStatement")
	}

	if localStmt.Names[0].Name != "x" {
		t.Errorf("expected 'x', got %s", localStmt.Names[0].Name)
	}
}

func TestParser_ParseReturn(t *testing.T) {
	input := `function foo()
    return 1, 2
end
`
	parser := NewParser(input)
	p, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	funcStmt := p.Statements[0].(*FunctionStatement)
	returnStmt := funcStmt.Body[0].(*ReturnStatement)

	if len(returnStmt.Results) != 2 {
		t.Errorf("expected 2 return values, got %d", len(returnStmt.Results))
	}
}

func TestParser_ParseExpressions(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"binary add", "x = 1 + 2"},
		{"binary sub", "x = 1 - 2"},
		{"binary mul", "x = 1 * 2"},
		{"binary div", "x = 1 / 2"},
		{"unary not", "x = not true"},
		{"unary minus", "x = -5"},
		{"parentheses", "x = (1 + 2) * 3"},
		{"chained comparison", "x = 1 < 2 <= 3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.input)
			p, err := parser.Parse()
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}
			if len(p.Statements) != 1 {
				t.Fatalf("expected 1 statement, got %d", len(p.Statements))
			}
		})
	}
}

func TestParser_MethodCall(t *testing.T) {
	input := `obj:method(arg1, arg2)`
	parser := NewParser(input)
	p, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	stmt := p.Statements[0].(*FunctionCallStatement)
	call := stmt.Function
	if call == nil {
		t.Fatal("expected FunctionCall, got nil")
	}

	if call.Method != "method" {
		t.Errorf("expected method 'method', got %s", call.Method)
	}

	if len(call.Arguments) != 2 {
		t.Errorf("expected 2 arguments, got %d", len(call.Arguments))
	}
}

func TestParser_IndexExpression(t *testing.T) {
	input := `x = t.key`
	parser := NewParser(input)
	p, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	stmt := p.Statements[0].(*AssignmentStatement)
	member, ok := stmt.Values[0].(*MemberExpression)
	if !ok {
		t.Fatal("expected MemberExpression")
	}

	if member.Member != "key" {
		t.Errorf("expected member 'key', got %s", member.Member)
	}
}

func TestParser_IndexBracket(t *testing.T) {
	input := `x = t["key"]`
	parser := NewParser(input)
	p, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	stmt := p.Statements[0].(*AssignmentStatement)
	index, ok := stmt.Values[0].(*IndexExpression)
	if !ok {
		t.Fatal("expected IndexExpression")
	}

	if _, ok := index.Index.(*StringLiteral); !ok {
		t.Error("expected StringLiteral as index")
	}
}

func TestParser_Errors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"unexpected token", "x = @"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.input)
			_, err := parser.Parse()
			if err == nil {
				t.Error("expected parse error")
			}
		})
	}
}

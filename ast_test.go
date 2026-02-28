package luar

import (
	"testing"
)

func TestAST_NodeTypes(t *testing.T) {
	program := &Program{
		Statements: []Statement{
			&AssignmentStatement{
				Names:  []*Identifier{{Name: "x"}},
				Values: []Expression{&NumberLiteral{Value: 10}},
			},
		},
	}

	if program.NodeType() != "Program" {
		t.Errorf("expected Program, got %s", program.NodeType())
	}
}

func TestAST_Identifier(t *testing.T) {
	ident := &Identifier{Name: "myVar", TokenLine: 1}
	if ident.Name != "myVar" {
		t.Errorf("expected 'myVar', got %s", ident.Name)
	}
	ident.ExpressionNode()
}

func TestAST_NumberLiteral(t *testing.T) {
	num := &NumberLiteral{Value: 3.14, IsInt: false, TokenLine: 1}
	if num.Value != 3.14 {
		t.Errorf("expected 3.14, got %f", num.Value)
	}
	if num.IsInt {
		t.Error("expected IsInt to be false")
	}
	num.ExpressionNode()
}

func TestAST_StringLiteral(t *testing.T) {
	str := &StringLiteral{Value: "hello", TokenLine: 1}
	if str.Value != "hello" {
		t.Errorf("expected 'hello', got %s", str.Value)
	}
	str.ExpressionNode()
}

func TestAST_BooleanLiteral(t *testing.T) {
	bl := &BooleanLiteral{Value: true, TokenLine: 1}
	if !bl.Value {
		t.Error("expected true")
	}
	bl.ExpressionNode()

	bf := &BooleanLiteral{Value: false, TokenLine: 1}
	if bf.Value {
		t.Error("expected false")
	}
	bf.ExpressionNode()
}

func TestAST_NilLiteral(t *testing.T) {
	nilLit := &NilLiteral{TokenLine: 1}
	nilLit.ExpressionNode()
}

func TestAST_TableLiteral(t *testing.T) {
	tbl := &TableLiteral{
		Fields: []*TableField{
			{
				Key:   &Identifier{Name: "key"},
				Value: &StringLiteral{Value: "value"},
			},
			{
				Value: &NumberLiteral{Value: 1},
			},
		},
		TokenLine: 1,
	}

	if len(tbl.Fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(tbl.Fields))
	}

	if tbl.Fields[0].Key == nil {
		t.Error("first field should have a key")
	}

	if tbl.Fields[1].Key != nil {
		t.Error("second field should not have a key")
	}

	tbl.ExpressionNode()
}

func TestAST_BinaryExpression(t *testing.T) {
	expr := &BinaryExpression{
		Operator:  PLUS,
		Left:      &NumberLiteral{Value: 1},
		Right:     &NumberLiteral{Value: 2},
		TokenLine: 1,
	}

	if expr.Operator != PLUS {
		t.Errorf("expected PLUS, got %v", expr.Operator)
	}

	if _, ok := expr.Left.(*NumberLiteral); !ok {
		t.Error("expected Left to be NumberLiteral")
	}

	if _, ok := expr.Right.(*NumberLiteral); !ok {
		t.Error("expected Right to be NumberLiteral")
	}

	expr.ExpressionNode()
}

func TestAST_UnaryExpression(t *testing.T) {
	expr := &UnaryExpression{
		Operator:  NOT,
		Right:     &BooleanLiteral{Value: false},
		TokenLine: 1,
	}

	if expr.Operator != NOT {
		t.Errorf("expected NOT, got %v", expr.Operator)
	}

	expr.ExpressionNode()
}

func TestAST_FunctionCall(t *testing.T) {
	call := &FunctionCall{
		Function: &Identifier{Name: "print"},
		Arguments: []Expression{
			&StringLiteral{Value: "hello"},
		},
		Method:    "",
		TokenLine: 1,
	}

	if call.Function.(*Identifier).Name != "print" {
		t.Error("expected function name 'print'")
	}

	if len(call.Arguments) != 1 {
		t.Errorf("expected 1 argument, got %d", len(call.Arguments))
	}

	call.ExpressionNode()
}

func TestAST_MemberExpression(t *testing.T) {
	expr := &MemberExpression{
		Object:    &Identifier{Name: "obj"},
		Member:    "field",
		TokenLine: 1,
	}

	if expr.Member != "field" {
		t.Errorf("expected member 'field', got %s", expr.Member)
	}

	expr.ExpressionNode()
}

func TestAST_IndexExpression(t *testing.T) {
	expr := &IndexExpression{
		Object:    &Identifier{Name: "arr"},
		Index:     &NumberLiteral{Value: 0},
		TokenLine: 1,
	}

	if expr.Index == nil {
		t.Error("expected index to be set")
	}

	expr.ExpressionNode()
}

func TestAST_FunctionLiteral(t *testing.T) {
	fn := &FunctionLiteral{
		Parameters: []*Identifier{
			{Name: "a"},
			{Name: "b"},
		},
		Body: []Statement{
			&ReturnStatement{
				Results: []Expression{
					&BinaryExpression{
						Operator: PLUS,
						Left:     &Identifier{Name: "a"},
						Right:    &Identifier{Name: "b"},
					},
				},
			},
		},
		TokenLine: 1,
	}

	if len(fn.Parameters) != 2 {
		t.Errorf("expected 2 parameters, got %d", len(fn.Parameters))
	}

	if len(fn.Body) != 1 {
		t.Errorf("expected 1 statement in body, got %d", len(fn.Body))
	}

	fn.ExpressionNode()
}

func TestAST_Statements(t *testing.T) {
	stmt := &AssignmentStatement{
		Names:  []*Identifier{{Name: "x"}},
		Values: []Expression{&NumberLiteral{Value: 10}},
	}
	stmt.StatementNode()

	local := &LocalAssignmentStatement{
		Names:  []*Identifier{{Name: "y"}},
		Values: []Expression{&StringLiteral{Value: "test"}},
	}
	local.StatementNode()

	ifStmt := &IfStatement{
		Condition: &BooleanLiteral{Value: true},
		Then:      []Statement{},
	}
	ifStmt.StatementNode()

	whileStmt := &WhileStatement{
		Condition: &BooleanLiteral{Value: true},
		Body:      []Statement{},
	}
	whileStmt.StatementNode()

	forStmt := &ForStatement{
		Body: []Statement{},
	}
	forStmt.StatementNode()

	fnStmt := &FunctionStatement{
		Name:       &FunctionName{Name: &Identifier{Name: "foo"}},
		Parameters: []*Identifier{},
		Body:       []Statement{},
	}
	fnStmt.StatementNode()

	returnStmt := &ReturnStatement{
		Results: []Expression{},
	}
	returnStmt.StatementNode()

	breakStmt := &BreakStatement{}
	breakStmt.StatementNode()
}

func TestAST_FunctionName(t *testing.T) {
	fnName := &FunctionName{
		Name:   &Identifier{Name: "foo"},
		Method: "bar",
	}

	if fnName.Name.Name != "foo" {
		t.Errorf("expected Name 'foo', got %s", fnName.Name.Name)
	}

	if fnName.Method != "bar" {
		t.Errorf("expected Method 'bar', got %s", fnName.Method)
	}

	fnName2 := &FunctionName{
		Table: &IndexExpression{
			Object: &Identifier{Name: "obj"},
			Index:  &StringLiteral{Value: "key"},
		},
		Name: &Identifier{Name: "method"},
	}

	if fnName2.Table == nil {
		t.Error("expected Table to be set")
	}
}

func TestAST_TableField(t *testing.T) {
	field := &TableField{
		Key:       &Identifier{Name: "key"},
		Value:     &StringLiteral{Value: "value"},
		TokenLine: 1,
	}

	if field.Key.(*Identifier).Name != "key" {
		t.Error("expected key 'key'")
	}

	field = &TableField{
		Value:     &NumberLiteral{Value: 42},
		TokenLine: 1,
	}

	if field.Key != nil {
		t.Error("expected no key for sequential field")
	}
}

func TestAST_ErrorNode(t *testing.T) {
	err := &ErrorNode{
		Message:   "test error",
		TokenLine: 1,
	}

	if err.Message != "test error" {
		t.Errorf("expected 'test error', got %s", err.Message)
	}

	err.ExpressionNode()
}

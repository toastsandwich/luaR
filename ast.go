package luar

type Node interface {
	NodeType() string
}

type Program struct {
	Statements []Statement
}

func (p *Program) NodeType() string { return "Program" }

type Statement interface {
	StatementNode()
}

type Expression interface {
	ExpressionNode()
}

func (s *AssignmentStatement) StatementNode()      {}
func (s *FunctionCallStatement) StatementNode()    {}
func (s *IfStatement) StatementNode()              {}
func (s *WhileStatement) StatementNode()           {}
func (s *RepeatStatement) StatementNode()          {}
func (s *ForStatement) StatementNode()             {}
func (s *ForInStatement) StatementNode()           {}
func (s *FunctionStatement) StatementNode()        {}
func (s *LocalAssignmentStatement) StatementNode() {}
func (s *LocalFunctionStatement) StatementNode()   {}
func (s *ReturnStatement) StatementNode()          {}
func (s *BreakStatement) StatementNode()           {}
func (s *LabelStatement) StatementNode()           {}
func (s *GotoStatement) StatementNode()            {}
func (s *SemicolonStatement) StatementNode()       {}

func (e *Identifier) ExpressionNode()       {}
func (e *NumberLiteral) ExpressionNode()    {}
func (e *StringLiteral) ExpressionNode()    {}
func (e *BooleanLiteral) ExpressionNode()   {}
func (e *NilLiteral) ExpressionNode()       {}
func (e *TableLiteral) ExpressionNode()     {}
func (e *FunctionLiteral) ExpressionNode()  {}
func (e *BinaryExpression) ExpressionNode() {}
func (e *UnaryExpression) ExpressionNode()  {}
func (e *IndexExpression) ExpressionNode()  {}
func (e *MemberExpression) ExpressionNode() {}
func (e *FunctionCall) ExpressionNode()     {}
func (e *TableIndex) ExpressionNode()       {}

type AssignmentStatement struct {
	Names     []*Identifier
	Values    []Expression
	TokenLine int
}

type LocalAssignmentStatement struct {
	Names     []*Identifier
	Values    []Expression
	TokenLine int
}

type FunctionCallStatement struct {
	Function *FunctionCall
}

type FunctionCall struct {
	Function  Expression
	Arguments []Expression
	Method    string
	TokenLine int
}

type IfStatement struct {
	Condition Expression
	Then      []Statement
	ElseIfs   []ElseIfClause
	Else      []Statement
	TokenLine int
}

type ElseIfClause struct {
	Condition Expression
	Then      []Statement
	TokenLine int
}

type WhileStatement struct {
	Condition Expression
	Body      []Statement
	TokenLine int
}

type RepeatStatement struct {
	Body      []Statement
	Condition Expression
	TokenLine int
}

type ForStatement struct {
	Init      *AssignmentStatement
	Condition Expression
	Post      *AssignmentStatement
	Body      []Statement
	TokenLine int
}

type ForInStatement struct {
	Names     []*Identifier
	Values    []Expression
	Body      []Statement
	TokenLine int
}

type FunctionStatement struct {
	Name       *FunctionName
	Parameters []*Identifier
	Body       []Statement
	TokenLine  int
}

type LocalFunctionStatement struct {
	Name       *Identifier
	Parameters []*Identifier
	Body       []Statement
	TokenLine  int
}

type ReturnStatement struct {
	Results   []Expression
	TokenLine int
}

type BreakStatement struct {
	TokenLine int
}

type LabelStatement struct {
	Name      string
	TokenLine int
}

type GotoStatement struct {
	Name      string
	TokenLine int
}

type SemicolonStatement struct {
	TokenLine int
}

type Identifier struct {
	Name      string
	TokenLine int
}

type FunctionName struct {
	Method string
	Table  *IndexExpression
	Name   *Identifier
}

type NumberLiteral struct {
	Value     float64
	IntValue  int64
	IsInt     bool
	TokenLine int
}

type StringLiteral struct {
	Value     string
	TokenLine int
}

type BooleanLiteral struct {
	Value     bool
	TokenLine int
}

type NilLiteral struct {
	TokenLine int
}

type TableLiteral struct {
	Fields    []*TableField
	TokenLine int
}

type TableField struct {
	Key       Expression
	Value     Expression
	TokenLine int
}

type FunctionLiteral struct {
	Parameters []*Identifier
	Body       []Statement
	TokenLine  int
}

type BinaryExpression struct {
	Operator  TokenType
	Left      Expression
	Right     Expression
	TokenLine int
}

type UnaryExpression struct {
	Operator  TokenType
	Right     Expression
	TokenLine int
}

type IndexExpression struct {
	Object    Expression
	Index     Expression
	TokenLine int
}

type MemberExpression struct {
	Object    Expression
	Member    string
	TokenLine int
}

type TableIndex struct {
	Key       Expression
	TokenLine int
}

type ErrorNode struct {
	Message   string
	TokenLine int
}

func (e *ErrorNode) ExpressionNode() {}

package luar

import (
	"fmt"
	"strconv"
	"strings"
)

type Parser struct {
	lexer  *Lexer
	tokens []Token
	pos    int
	errors []string
}

func NewParser(input string) *Parser {
	lexer := NewLexer(input)
	tokens := lexer.Tokens()
	return &Parser{
		lexer:  lexer,
		tokens: tokens,
	}
}

func (p *Parser) currentToken() Token {
	if p.pos >= len(p.tokens) {
		return p.tokens[len(p.tokens)-1]
	}
	return p.tokens[p.pos]
}

func (p *Parser) peekToken(offset int) Token {
	pos := p.pos + offset
	if pos >= len(p.tokens) {
		return p.tokens[len(p.tokens)-1]
	}
	return p.tokens[pos]
}

func (p *Parser) advance() Token {
	if p.pos >= len(p.tokens) {
		return p.tokens[len(p.tokens)-1]
	}
	tok := p.tokens[p.pos]
	p.pos++
	return tok
}

func (p *Parser) expect(t TokenType) Token {
	if p.currentToken().Type == t {
		token := p.currentToken()
		p.advance()
		return token
	}
	p.errors = append(p.errors, fmt.Sprintf("expected %s but got %s at line %d", t, p.currentToken().Type, p.currentToken().Line))
	return Token{Type: t}
}

func (p *Parser) check(t TokenType) bool {
	return p.currentToken().Type == t
}

func (p *Parser) match(t TokenType) bool {
	if p.check(t) {
		p.advance()
		return true
	}
	return false
}

func (p *Parser) errorsAsString() string {
	return strings.Join(p.errors, "\n")
}

func (p *Parser) Parse() (*Program, error) {
	program := &Program{
		Statements: []Statement{},
	}

	for !p.check(EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
	}

	if len(p.errors) > 0 {
		return program, fmt.Errorf("%s", p.errorsAsString())
	}

	return program, nil
}

func (p *Parser) parseStatement() Statement {
	switch p.currentToken().Type {
	case IF:
		return p.parseIfStatement()
	case WHILE:
		return p.parseWhileStatement()
	case REPEAT:
		return p.parseRepeatStatement()
	case FOR:
		return p.parseForStatement()
	case FUNCTION:
		return p.parseFunctionStatement()
	case LOCAL:
		return p.parseLocalStatement()
	case RETURN:
		return p.parseReturnStatement()
	case BREAK:
		p.advance()
		return &BreakStatement{TokenLine: p.currentToken().Line}
	case GOTO:
		return p.parseGotoStatement()
	case LABEL:
		return p.parseLabelStatement()
	case SEMICOLON:
		p.advance()
		return &SemicolonStatement{TokenLine: p.currentToken().Line}
	default:
		return p.parseAssignmentOrExpression()
	}
}

func (p *Parser) parseIfStatement() *IfStatement {
	ifToken := p.expect(IF)
	condition := p.parseExpression()
	p.expect(THEN)

	thenBlock := p.parseBlock()

	elseIfs := []ElseIfClause{}
	elseBlock := []Statement{}

	for p.check(ELSEIF) {
		p.advance()
		elseIfCond := p.parseExpression()
		p.expect(THEN)
		elseIfBlock := p.parseBlock()
		elseIfs = append(elseIfs, ElseIfClause{
			Condition: elseIfCond,
			Then:      elseIfBlock,
			TokenLine: p.currentToken().Line,
		})
	}

	if p.check(ELSE) {
		p.advance()
		elseBlock = p.parseBlock()
	}

	p.expect(END)

	return &IfStatement{
		Condition: condition,
		Then:      thenBlock,
		ElseIfs:   elseIfs,
		Else:      elseBlock,
		TokenLine: ifToken.Line,
	}
}

func (p *Parser) parseWhileStatement() *WhileStatement {
	whileToken := p.expect(WHILE)
	condition := p.parseExpression()
	p.expect(DO)
	body := p.parseBlock()
	p.expect(END)

	return &WhileStatement{
		Condition: condition,
		Body:      body,
		TokenLine: whileToken.Line,
	}
}

func (p *Parser) parseRepeatStatement() *RepeatStatement {
	repeatToken := p.expect(REPEAT)
	body := p.parseBlock()
	p.expect(UNTIL)
	condition := p.parseExpression()

	return &RepeatStatement{
		Body:      body,
		Condition: condition,
		TokenLine: repeatToken.Line,
	}
}

func (p *Parser) parseForStatement() Statement {
	forToken := p.expect(FOR)

	if p.peekToken(1).Type == ASSIGN {
		name := &Identifier{Name: p.expect(IDENT).Literal, TokenLine: forToken.Line}
		p.expect(ASSIGN)
		initVal := p.parseExpression()
		p.expect(COMMA)
		endVal := p.parseExpression()

		var step Expression
		var stepTokenLine int
		if p.check(COMMA) {
			p.advance()
			step = p.parseExpression()
			stepTokenLine = p.currentToken().Line
		}
		p.expect(DO)
		body := p.parseBlock()
		p.expect(END)

		return &ForStatement{
			Init: &AssignmentStatement{
				Names:     []*Identifier{name},
				Values:    []Expression{initVal},
				TokenLine: forToken.Line,
			},
			Condition: endVal,
			Post:      &AssignmentStatement{Names: []*Identifier{name}, Values: []Expression{step}, TokenLine: stepTokenLine},
			Body:      body,
			TokenLine: forToken.Line,
		}
	}

	name := p.expect(IDENT)
	names := []*Identifier{{Name: name.Literal, TokenLine: name.Line}}

	if p.check(COMMA) {
		p.advance()
		names = append(names, &Identifier{Name: p.expect(IDENT).Literal, TokenLine: p.currentToken().Line})
	}

	p.expect(IN)
	values := []Expression{p.parseExpression()}
	for p.check(COMMA) {
		p.advance()
		values = append(values, p.parseExpression())
	}

	p.expect(DO)
	body := p.parseBlock()
	p.expect(END)

	return &ForInStatement{
		Names:     names,
		Values:    values,
		Body:      body,
		TokenLine: forToken.Line,
	}
}

func (p *Parser) parseFunctionStatement() *FunctionStatement {
	funcToken := p.expect(FUNCTION)
	name := p.parseFunctionName()
	p.expect(LPAREN)

	parameters := []*Identifier{}
	if !p.check(RPAREN) {
		for {
			if p.check(IDENT) {
				parameters = append(parameters, &Identifier{Name: p.expect(IDENT).Literal, TokenLine: p.currentToken().Line})
			} else if p.check(ELLIPSIS) {
				parameters = append(parameters, &Identifier{Name: "...", TokenLine: p.currentToken().Line})
				p.advance()
			}
			if p.check(COMMA) {
				p.advance()
				continue
			}
			break
		}
	}
	p.expect(RPAREN)

	body := p.parseBlock()
	p.expect(END)

	return &FunctionStatement{
		Name:       name,
		Parameters: parameters,
		Body:       body,
		TokenLine:  funcToken.Line,
	}
}

func (p *Parser) parseFunctionName() *FunctionName {
	name := &FunctionName{}

	if p.check(IDENT) {
		name.Name = &Identifier{Name: p.expect(IDENT).Literal, TokenLine: p.currentToken().Line}
	}

	if p.check(DOT) {
		p.advance()
		if name.Name == nil {
			name.Name = &Identifier{}
		}
		current := name.Name.Name
		current += "." + p.expect(IDENT).Literal
		name.Name.Name = current
	}

	if p.check(COLON) {
		p.advance()
		name.Method = p.expect(IDENT).Literal
	}

	return name
}

func (p *Parser) parseLocalStatement() Statement {
	localToken := p.expect(LOCAL)

	if p.check(FUNCTION) {
		return p.parseLocalFunction(localToken)
	}

	return p.parseLocalAssignment(localToken)
}

func (p *Parser) parseLocalFunction(localToken Token) *LocalFunctionStatement {
	p.expect(FUNCTION)
	name := &Identifier{Name: p.expect(IDENT).Literal, TokenLine: localToken.Line}
	p.expect(LPAREN)

	parameters := []*Identifier{}
	if !p.check(RPAREN) {
		for {
			if p.check(IDENT) {
				parameters = append(parameters, &Identifier{Name: p.expect(IDENT).Literal, TokenLine: p.currentToken().Line})
			} else if p.check(ELLIPSIS) {
				parameters = append(parameters, &Identifier{Name: "...", TokenLine: p.currentToken().Line})
				p.advance()
			}
			if p.check(COMMA) {
				p.advance()
				continue
			}
			break
		}
	}
	p.expect(RPAREN)

	body := p.parseBlock()
	p.expect(END)

	return &LocalFunctionStatement{
		Name:       name,
		Parameters: parameters,
		Body:       body,
		TokenLine:  localToken.Line,
	}
}

func (p *Parser) parseLocalAssignment(localToken Token) *LocalAssignmentStatement {
	names := []*Identifier{}
	for {
		if p.check(IDENT) {
			names = append(names, &Identifier{Name: p.expect(IDENT).Literal, TokenLine: p.currentToken().Line})
		}
		if p.check(COMMA) {
			p.advance()
			continue
		}
		break
	}

	values := []Expression{}
	if p.check(ASSIGN) {
		p.advance()
		values = p.parseExpressionList()
	}

	return &LocalAssignmentStatement{
		Names:     names,
		Values:    values,
		TokenLine: localToken.Line,
	}
}

func (p *Parser) parseReturnStatement() *ReturnStatement {
	returnToken := p.expect(RETURN)

	results := []Expression{}
	if !p.check(END) && !p.check(ELSE) && !p.check(ELSEIF) && !p.check(UNTIL) && !p.check(EOF) {
		results = p.parseExpressionList()
	}

	if p.check(SEMICOLON) {
		p.advance()
	}

	return &ReturnStatement{
		Results:   results,
		TokenLine: returnToken.Line,
	}
}

func (p *Parser) parseGotoStatement() *GotoStatement {
	gotoToken := p.expect(GOTO)
	name := p.expect(IDENT)

	return &GotoStatement{
		Name:      name.Literal,
		TokenLine: gotoToken.Line,
	}
}

func (p *Parser) parseLabelStatement() *LabelStatement {
	labelToken := p.expect(LABEL)
	name := p.expect(IDENT)
	p.expect(LABEL)

	return &LabelStatement{
		Name:      name.Literal,
		TokenLine: labelToken.Line,
	}
}

func (p *Parser) parseAssignmentOrExpression() Statement {
	expr := p.parseExpression()

	if p.check(ASSIGN) {
		p.advance()

		if ident, ok := expr.(*Identifier); ok {
			names := []*Identifier{ident}
			values := p.parseExpressionList()
			return &AssignmentStatement{
				Names:     names,
				Values:    values,
				TokenLine: p.currentToken().Line,
			}
		}

		if member, ok := expr.(*MemberExpression); ok {
			values := p.parseExpressionList()
			return &AssignmentStatement{
				Names:     []*Identifier{{Name: member.Object.(*Identifier).Name + "." + member.Member}},
				Values:    values,
				TokenLine: p.currentToken().Line,
			}
		}

		if index, ok := expr.(*IndexExpression); ok {
			values := p.parseExpressionList()
			var nameStr string
			if ident, ok := index.Object.(*Identifier); ok {
				nameStr = ident.Name
			}
			return &AssignmentStatement{
				Names:     []*Identifier{{Name: nameStr}},
				Values:    values,
				TokenLine: p.currentToken().Line,
			}
		}
	}

	if fnCall, ok := expr.(*FunctionCall); ok {
		return &FunctionCallStatement{Function: fnCall}
	}

	return &AssignmentStatement{
		Values:    []Expression{expr},
		TokenLine: p.currentToken().Line,
	}
}

func (p *Parser) parseBlock() []Statement {
	statements := []Statement{}

	for {
		if p.check(END) || p.check(ELSE) || p.check(ELSEIF) || p.check(UNTIL) || p.check(EOF) {
			break
		}

		stmt := p.parseStatement()
		if stmt != nil {
			statements = append(statements, stmt)
		}
	}

	return statements
}

func (p *Parser) parseExpressionList() []Expression {
	exprs := []Expression{p.parseExpression()}

	for p.check(COMMA) {
		p.advance()
		exprs = append(exprs, p.parseExpression())
	}

	return exprs
}

func (p *Parser) parseExpression() Expression {
	return p.parseOr()
}

func (p *Parser) parseOr() Expression {
	left := p.parseAnd()

	for p.check(OR) {
		op := p.advance()
		right := p.parseAnd()
		left = &BinaryExpression{Operator: op.Type, Left: left, Right: right, TokenLine: op.Line}
	}

	return left
}

func (p *Parser) parseAnd() Expression {
	left := p.parseComparison()

	for p.check(AND) {
		op := p.advance()
		right := p.parseComparison()
		left = &BinaryExpression{Operator: op.Type, Left: left, Right: right, TokenLine: op.Line}
	}

	return left
}

func (p *Parser) parseComparison() Expression {
	left := p.parseConcat()

	for p.check(EQ) || p.check(NE) || p.check(LT) || p.check(LE) || p.check(GT) || p.check(GE) {
		op := p.advance()
		right := p.parseConcat()
		left = &BinaryExpression{Operator: op.Type, Left: left, Right: right, TokenLine: op.Line}
	}

	return left
}

func (p *Parser) parseConcat() Expression {
	left := p.parseBitwiseOr()

	if p.check(CONCAT) {
		ops := []Token{p.advance()}
		rights := []Expression{p.parseBitwiseOr()}

		for p.check(CONCAT) {
			ops = append(ops, p.advance())
			rights = append(rights, p.parseBitwiseOr())
		}

		result := left
		for i, right := range rights {
			result = &BinaryExpression{Operator: ops[i].Type, Left: result, Right: right, TokenLine: ops[i].Line}
		}
		return result
	}

	return left
}

func (p *Parser) parseBitwiseOr() Expression {
	left := p.parseBitwiseXor()

	for p.check(OR) || p.check(LSHIFT) || p.check(RSHIFT) {
		op := p.advance()
		right := p.parseBitwiseXor()
		left = &BinaryExpression{Operator: op.Type, Left: left, Right: right, TokenLine: op.Line}
	}

	return left
}

func (p *Parser) parseBitwiseXor() Expression {
	left := p.parseBitwiseAnd()

	for p.check(POW) {
		op := p.advance()
		right := p.parseBitwiseAnd()
		left = &BinaryExpression{Operator: op.Type, Left: left, Right: right, TokenLine: op.Line}
	}

	return left
}

func (p *Parser) parseBitwiseAnd() Expression {
	left := p.parseAddSub()

	for p.check(HASH) || p.check(AND) {
		op := p.advance()
		right := p.parseAddSub()
		left = &BinaryExpression{Operator: op.Type, Left: left, Right: right, TokenLine: op.Line}
	}

	return left
}

func (p *Parser) parseAddSub() Expression {
	left := p.parseMulDivMod()

	for p.check(PLUS) || p.check(MINUS) {
		op := p.advance()
		right := p.parseMulDivMod()
		left = &BinaryExpression{Operator: op.Type, Left: left, Right: right, TokenLine: op.Line}
	}

	return left
}

func (p *Parser) parseMulDivMod() Expression {
	left := p.parseUnary()

	for p.check(STAR) || p.check(SLASH) || p.check(MOD) {
		op := p.advance()
		right := p.parseUnary()
		left = &BinaryExpression{Operator: op.Type, Left: left, Right: right, TokenLine: op.Line}
	}

	return left
}

func (p *Parser) parseUnary() Expression {
	if p.check(NOT) || p.check(MINUS) || p.check(HASH) {
		op := p.advance()
		right := p.parseUnary()
		return &UnaryExpression{Operator: op.Type, Right: right, TokenLine: op.Line}
	}

	return p.parsePow()
}

func (p *Parser) parsePow() Expression {
	left := p.parsePostfix()

	for p.check(POW) {
		op := p.advance()
		right := p.parseUnary()
		left = &BinaryExpression{Operator: op.Type, Left: left, Right: right, TokenLine: op.Line}
	}

	return left
}

func (p *Parser) parsePostfix() Expression {
	expr := p.parsePrimary()

	for {
		if p.check(DOT) {
			p.advance()
			member := p.expect(IDENT)
			expr = &MemberExpression{Object: expr, Member: member.Literal, TokenLine: p.currentToken().Line}
		} else if p.check(LBRACKET) {
			p.advance()
			index := p.parseExpression()
			p.expect(RBRACKET)
			expr = &IndexExpression{Object: expr, Index: index, TokenLine: p.currentToken().Line}
		} else if p.check(COLON) {
			p.advance()
			method := p.expect(IDENT).Literal
			p.expect(LPAREN)
			args := []Expression{}
			if !p.check(RPAREN) {
				args = p.parseExpressionList()
			}
			p.expect(RPAREN)
			expr = &FunctionCall{Function: expr, Method: method, Arguments: args, TokenLine: p.currentToken().Line}
		} else if p.check(LPAREN) || p.check(STRING) || p.check(LBRACE) {
			var args []Expression
			if p.check(LPAREN) {
				p.advance()
				if !p.check(RPAREN) {
					args = p.parseExpressionList()
				}
				p.expect(RPAREN)
			} else if p.check(STRING) || p.check(LBRACE) {
				args = p.parseExpressionList()
			}
			expr = &FunctionCall{Function: expr, Arguments: args, TokenLine: p.currentToken().Line}
		} else {
			break
		}
	}

	return expr
}

func (p *Parser) parsePrimary() Expression {
	switch p.currentToken().Type {
	case IDENT:
		ident := p.expect(IDENT)
		return &Identifier{Name: ident.Literal, TokenLine: ident.Line}
	case INT, FLOAT:
		lit := p.advance()
		if lit.Type == INT {
			val, _ := strconv.ParseInt(lit.Literal, 0, 64)
			return &NumberLiteral{IntValue: val, IsInt: true, TokenLine: lit.Line}
		}
		val, _ := strconv.ParseFloat(lit.Literal, 64)
		return &NumberLiteral{Value: val, IsInt: false, TokenLine: lit.Line}
	case STRING:
		str := p.expect(STRING)
		return &StringLiteral{Value: str.Literal, TokenLine: str.Line}
	case TRUE:
		p.advance()
		return &BooleanLiteral{Value: true, TokenLine: p.currentToken().Line}
	case FALSE:
		p.advance()
		return &BooleanLiteral{Value: false, TokenLine: p.currentToken().Line}
	case NIL:
		p.advance()
		return &NilLiteral{TokenLine: p.currentToken().Line}
	case LBRACE:
		return p.parseTableLiteral()
	case FUNCTION:
		return p.parseFunctionLiteral()
	case LPAREN:
		p.advance()
		expr := p.parseExpression()
		p.expect(RPAREN)
		return expr
	default:
		p.errors = append(p.errors, fmt.Sprintf("unexpected token: %s at line %d", p.currentToken().Type, p.currentToken().Line))
		p.advance()
		return &ErrorNode{Message: "unexpected token", TokenLine: p.currentToken().Line}
	}
}

func (p *Parser) parseTableLiteral() *TableLiteral {
	braceToken := p.expect(LBRACE)
	fields := []*TableField{}

	if !p.check(RBRACE) {
		for {
			field := p.parseTableField()
			if field != nil {
				fields = append(fields, field)
			}

			if p.check(RBRACE) {
				break
			}

			p.expect(COMMA)
		}
	}

	p.expect(RBRACE)

	return &TableLiteral{
		Fields:    fields,
		TokenLine: braceToken.Line,
	}
}

func (p *Parser) parseTableField() *TableField {
	key := p.parseExpression()

	if p.check(ASSIGN) {
		p.advance()
		value := p.parseExpression()
		return &TableField{Key: key, Value: value, TokenLine: p.currentToken().Line}
	}

	return &TableField{Value: key, TokenLine: p.currentToken().Line}
}

func (p *Parser) parseFunctionLiteral() *FunctionLiteral {
	funcToken := p.expect(FUNCTION)
	p.expect(LPAREN)

	parameters := []*Identifier{}
	if !p.check(RPAREN) {
		for {
			if p.check(IDENT) {
				parameters = append(parameters, &Identifier{Name: p.expect(IDENT).Literal, TokenLine: p.currentToken().Line})
			} else if p.check(ELLIPSIS) {
				parameters = append(parameters, &Identifier{Name: "...", TokenLine: p.currentToken().Line})
				p.advance()
			}
			if p.check(COMMA) {
				p.advance()
				continue
			}
			break
		}
	}
	p.expect(RPAREN)

	body := p.parseBlock()
	p.expect(END)

	return &FunctionLiteral{
		Parameters: parameters,
		Body:       body,
		TokenLine:  funcToken.Line,
	}
}

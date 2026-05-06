package parser

import (
	"fmt"
	"strconv"

	"pop/ast"
	"pop/lexer"
)

type Parser struct {
	tokens []lexer.Token
	pos    int
}

func New(tokens []lexer.Token) *Parser {
	return &Parser{tokens: tokens, pos: 0}
}

func (p *Parser) peek() lexer.Token {
	for p.pos < len(p.tokens) && p.tokens[p.pos].Type == lexer.TOKEN_NEWLINE {
		p.pos++
	}
	if p.pos >= len(p.tokens) {
		return lexer.Token{Type: lexer.TOKEN_EOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) advance() lexer.Token {
	t := p.peek()
	p.pos++
	return t
}

func (p *Parser) expect(tt lexer.TokenType) (lexer.Token, error) {
	t := p.peek()
	if t.Type != tt {
		return t, fmt.Errorf("line %d: expected %s, got %s (%q)", t.Line, tt, t.Type, t.Value)
	}
	p.pos++
	return t, nil
}

func (p *Parser) skipNewlines() {
	for p.pos < len(p.tokens) && p.tokens[p.pos].Type == lexer.TOKEN_NEWLINE {
		p.pos++
	}
}

func (p *Parser) Parse() (*ast.Program, error) {
	prog := &ast.Program{}
	for p.peek().Type != lexer.TOKEN_EOF {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			prog.Statements = append(prog.Statements, stmt)
		}
	}
	return prog, nil
}

func (p *Parser) parseStatement() (ast.Node, error) {
	t := p.peek()

	switch t.Type {
	case lexer.TOKEN_LIB:
		return p.parseLibImport()
	case lexer.TOKEN_TYPE_INT, lexer.TOKEN_TYPE_FLOAT, lexer.TOKEN_TYPE_STRING, lexer.TOKEN_TYPE_BOOL:
		return p.parseVarDecl()
	case lexer.TOKEN_IF:
		return p.parseIfStmt()
	case lexer.TOKEN_PRINT:
		return p.parsePrintStmt()
	case lexer.TOKEN_STOP:
		return nil, fmt.Errorf("line %d: unexpected 'stop' outside if block", t.Line)
	case lexer.TOKEN_EOF:
		return nil, nil
	default:
		return nil, fmt.Errorf("line %d: unexpected token %s (%q)", t.Line, t.Type, t.Value)
	}
}


func (p *Parser) parseLibImport() (ast.Node, error) {
	line := p.peek().Line
	p.advance() 
	if _, err := p.expect(lexer.TOKEN_IMPORT); err != nil {
		return nil, err
	}
	modTok, err := p.expect(lexer.TOKEN_IDENT)
	if err != nil {
		return nil, err
	}
	node := &ast.LibImport{Module: modTok.Value, Line: line}


	if p.peek().Type == lexer.TOKEN_DOT {
		p.advance()
		symTok, err := p.expect(lexer.TOKEN_IDENT)
		if err != nil {
			return nil, err
		}
		node.Symbol = symTok.Value
	}
	return node, nil
}


func (p *Parser) parseVarDecl() (ast.Node, error) {
	typeTok := p.advance()
	line := typeTok.Line
	nameTok, err := p.expect(lexer.TOKEN_IDENT)
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(lexer.TOKEN_ASSIGN); err != nil {
		return nil, err
	}
	val, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	typeName := ""
	switch typeTok.Type {
	case lexer.TOKEN_TYPE_INT:
		typeName = "int"
	case lexer.TOKEN_TYPE_FLOAT:
		typeName = "float"
	case lexer.TOKEN_TYPE_STRING:
		typeName = "string"
	case lexer.TOKEN_TYPE_BOOL:
		typeName = "bool"
	}
	return &ast.VarDecl{TypeName: typeName, Name: nameTok.Value, Value: val, Line: line}, nil
}


func (p *Parser) parseIfStmt() (ast.Node, error) {
	line := p.peek().Line
	p.advance() 
	cond, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(lexer.TOKEN_SEMICOLON); err != nil {
		return nil, err
	}

	var body []ast.Node
	for {
		p.skipNewlines()
		t := p.peek()
		if t.Type == lexer.TOKEN_STOP {
			p.advance()
			if _, err := p.expect(lexer.TOKEN_SEMICOLON); err != nil {
				return nil, err
			}
			break
		}
		if t.Type == lexer.TOKEN_EOF {
			return nil, fmt.Errorf("line %d: unterminated if block, missing 'stop;'", line)
		}
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			body = append(body, stmt)
		}
	}
	return &ast.IfStmt{Condition: cond, Body: body, Line: line}, nil
}


func (p *Parser) parsePrintStmt() (ast.Node, error) {
	line := p.peek().Line
	p.advance()
	if _, err := p.expect(lexer.TOKEN_LPAREN); err != nil {
		return nil, err
	}
	val, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(lexer.TOKEN_RPAREN); err != nil {
		return nil, err
	}
	return &ast.PrintStmt{Value: val, Line: line}, nil
}


func (p *Parser) parseExpr() (ast.Node, error) {
	return p.parseComparison()
}

func (p *Parser) parseComparison() (ast.Node, error) {
	left, err := p.parseAddSub()
	if err != nil {
		return nil, err
	}
	for {
		t := p.peek()
		if t.Type != lexer.TOKEN_LT && t.Type != lexer.TOKEN_LTE &&
			t.Type != lexer.TOKEN_GT && t.Type != lexer.TOKEN_GTE &&
			t.Type != lexer.TOKEN_EQ && t.Type != lexer.TOKEN_NEQ {
			break
		}
		op := p.advance().Value
		right, err := p.parseAddSub()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpr{Op: op, Left: left, Right: right}
	}
	return left, nil
}

func (p *Parser) parseAddSub() (ast.Node, error) {
	left, err := p.parseMulDiv()
	if err != nil {
		return nil, err
	}
	for {
		t := p.peek()
		if t.Type != lexer.TOKEN_PLUS && t.Type != lexer.TOKEN_MINUS {
			break
		}
		op := p.advance().Value
		right, err := p.parseMulDiv()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpr{Op: op, Left: left, Right: right}
	}
	return left, nil
}

func (p *Parser) parseMulDiv() (ast.Node, error) {
	left, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}
	for {
		t := p.peek()
		if t.Type != lexer.TOKEN_STAR && t.Type != lexer.TOKEN_SLASH {
			break
		}
		op := p.advance().Value
		right, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryExpr{Op: op, Left: left, Right: right}
	}
	return left, nil
}

func (p *Parser) parsePrimary() (ast.Node, error) {
	t := p.peek()

	switch t.Type {
	case lexer.TOKEN_INT_LIT:
		p.advance()
		v, _ := strconv.ParseInt(t.Value, 10, 64)
		return &ast.IntLit{Value: v}, nil

	case lexer.TOKEN_FLOAT_LIT:
		p.advance()
		v, _ := strconv.ParseFloat(t.Value, 64)
		return &ast.FloatLit{Value: v}, nil

	case lexer.TOKEN_STRING_LIT:
		p.advance()
		return &ast.StringLit{Value: t.Value}, nil

	case lexer.TOKEN_BOOL_LIT:
		p.advance()
		return &ast.BoolLit{Value: t.Value == "true"}, nil

	case lexer.TOKEN_LPAREN:
		p.advance()
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(lexer.TOKEN_RPAREN); err != nil {
			return nil, err
		}
		return expr, nil

	case lexer.TOKEN_IDENT:
		return p.parseIdentOrCall()

	default:
		return nil, fmt.Errorf("line %d: unexpected token in expression: %s (%q)", t.Line, t.Type, t.Value)
	}
}


func (p *Parser) parseIdentOrCall() (ast.Node, error) {
	nameTok := p.advance()
	line := nameTok.Line
	if p.peek().Type == lexer.TOKEN_DOT {
		p.advance() 
		funcTok, err := p.expect(lexer.TOKEN_IDENT)
		if err != nil {
			return nil, err
		}
		if p.peek().Type == lexer.TOKEN_LPAREN {
			args, err := p.parseCallArgs()
			if err != nil {
				return nil, err
			}
			return &ast.CallExpr{Module: nameTok.Value, Func: funcTok.Value, Args: args, Line: line}, nil
		}

		return &ast.Identifier{Name: nameTok.Value + "." + funcTok.Value, Line: line}, nil
	}


	if p.peek().Type == lexer.TOKEN_LPAREN {
		args, err := p.parseCallArgs()
		if err != nil {
			return nil, err
		}
		return &ast.CallExpr{Func: nameTok.Value, Args: args, Line: line}, nil
	}


	return &ast.Identifier{Name: nameTok.Value, Line: line}, nil
}


func (p *Parser) parseCallArgs() ([]ast.CallArg, error) {
	if _, err := p.expect(lexer.TOKEN_LPAREN); err != nil {
		return nil, err
	}
	var args []ast.CallArg
	for p.peek().Type != lexer.TOKEN_RPAREN {
		if len(args) > 0 {
			if _, err := p.expect(lexer.TOKEN_COMMA); err != nil {
				return nil, err
			}
		}

		if p.peek().Type == lexer.TOKEN_IDENT {
			saved := p.pos
			nameTok := p.advance()
			if p.peek().Type == lexer.TOKEN_ASSIGN {
				p.advance() 
				val, err := p.parseExpr()
				if err != nil {
					return nil, err
				}
				args = append(args, ast.CallArg{Name: nameTok.Value, Value: val})
				continue
			}

			p.pos = saved
		}

		val, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		args = append(args, ast.CallArg{Value: val})
	}
	if _, err := p.expect(lexer.TOKEN_RPAREN); err != nil {
		return nil, err
	}
	return args, nil
}

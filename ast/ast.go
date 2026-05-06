package ast
import "fmt"

type Node interface {
	nodeType() string
}


type Program struct {
	Statements []Node
}

func (p *Program) nodeType() string { return "Program" }


type VarDecl struct {
	TypeName string
	Name     string
	Value    Node
	Line     int
}

func (v *VarDecl) nodeType() string { return "VarDecl" }


type LibImport struct {
	Module string
	Symbol string
	Line   int
}

func (l *LibImport) nodeType() string { return "LibImport" }


type IfStmt struct {
	Condition Node
	Body      []Node
	Line      int
}

func (i *IfStmt) nodeType() string { return "IfStmt" }


type PrintStmt struct {
	Value Node
	Line  int
}

func (p *PrintStmt) nodeType() string { return "PrintStmt" }


type ExprStmt struct {
	Expr Node
}

func (e *ExprStmt) nodeType() string { return "ExprStmt" }


type IntLit struct {
	Value int64
}

func (i *IntLit) nodeType() string { return "IntLit" }


type FloatLit struct {
	Value float64
}

func (f *FloatLit) nodeType() string { return "FloatLit" }


type StringLit struct {
	Value string
}

func (s *StringLit) nodeType() string { return "StringLit" }


type BoolLit struct {
	Value bool
}

func (b *BoolLit) nodeType() string { return "BoolLit" }


type Identifier struct {
	Name string
	Line int
}

func (i *Identifier) nodeType() string { return "Identifier" }


type BinaryExpr struct {
	Op    string
	Left  Node
	Right Node
}

func (b *BinaryExpr) nodeType() string { return "BinaryExpr" }


type CallExpr struct {
	Module string 
	Func   string
	Args   []CallArg
	Line   int
}

func (c *CallExpr) nodeType() string { return "CallExpr" }

func (c *CallExpr) String() string {
	if c.Module != "" {
		return fmt.Sprintf("%s.%s(...)", c.Module, c.Func)
	}
	return fmt.Sprintf("%s(...)", c.Func)
}


type CallArg struct {
	Name  string 
	Value Node
}

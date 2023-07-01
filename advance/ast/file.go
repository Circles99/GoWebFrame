package ast

import (
	"go/ast"
	"go/token"
)

type printVisitor struct {
	Import []string
	types  []*typeVisitor
}

func (p printVisitor) Visit(node ast.Node) (w ast.Visitor) {
	switch n := node.(type) {
	case *ast.TypeSpec:
		v := &typeVisitor{name: n.Name.String()}
		p.types = append(p.types, v)
		return v
	case *ast.GenDecl:
		if n.Tok == token.IMPORT {
			// 这是import
			for _, spec := range n.Specs {
				p.Import = append(p.Import, spec.(*ast.ImportSpec).Path.Value)
			}
		}
	}
	return p
}

type typeVisitor struct {
	name   string
	fields []Field
}

type Field struct {
	Name string
	Typ  string
}

func (t typeVisitor) Visit(node ast.Node) (w ast.Visitor) {

	n, ok := node.(*ast.Field)
	if !ok {
		return t
	}

	var typ string
	// 获取字段类型
	switch t := n.Type.(type) {
	case *ast.Ident:
		typ = t.String()

	case *ast.StarExpr:
		switch xt := t.X.(type) {
		case *ast.Ident:
			typ = "*" + xt.String()
		case *ast.SelectorExpr:
			typ = "*" + xt.X.(*ast.Ident).String() + "." + xt.Sel.String()
		}

	case *ast.ArrayType:
	//typ =
	default:
		panic("不支持类型")
	}

	for _, name := range n.Names {
		t.fields = append(t.fields, Field{
			Name: name.String(),
			Typ:  typ,
		})
	}
	return t
}

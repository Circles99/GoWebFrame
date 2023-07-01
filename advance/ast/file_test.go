package ast

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestFile(t *testing.T) {
	src := `
package main
import "fmt"

type User struct {
	Name     string
	Age      *int
	NickName *sql.NullString
}



func main() {
	fmt.Println("hello world")
}
`

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		t.Fatal("解析错误")
	}

	ast.Walk(printVisitor{}, f)
	fmt.Println(111)
}

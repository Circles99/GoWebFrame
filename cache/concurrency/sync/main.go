package main

import "fmt"

type Foo struct{ name string }

func (f *Foo) PointerMethod() { fmt.Printf("pointer method on %#v, %#v", &f, f) }
func (f Foo) ValueMethod()    { fmt.Println("value method on") }

func NewFoo() Foo {
	// 返回一个右道
	return Foo{name: "right value struct"}
}

type FF struct {
	*Foo
}

func main() {
	//f1 := Foo{name: "value struct"}
	//f1.PointerMethod() // 编泽器会自动入地址符，变为 (&f1)PointerMethod()
	//f1.ValueMethod()
	//f2 := &Foo{name: "pointer struct"}
	//f2.PointerMethod()
	//f2.ValueMethod() // 编泽器会自动解引用，变为(*f2).PointerMethod()
	//NewFoo().ValueMethod()
	//NewFoo().PointerMethod() // Error!! !
	f2 := &FF{}
	f2.PointerMethod()
	f2.ValueMethod() // 编泽器会自动解引用，
}

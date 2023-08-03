package main

import (
	"fmt"
	"testing"
)

// 因为dd组合了指针base, 去选址时，因为base.BB不是
// dd中base是指针， 编泽器会自动解引用，变为 (*d)BB(), 因为base是指针，并没有实例化，base=nil, 找不到BB的地址 故而报错
// dd中base不是指针 编泽器会自动入地址符， 变为 (&d)bb(), 因base不是指针，D在实例的时候，也会给base分配地址，在解引用的时候找到地址，执行BB方法

// dd中base不是指针， BB是指针方法： dd.base = main.Base{}(00028x), 调用BB方法， 编译器自动入地址符，找到地址直接调用BB
// dd中base不是指针， BB不是指针方法： dd.base = main.Base{}(00028x), 调用BB方法， 直接值调用
// dd中的base是指针， BB是指针方法， dd.base = main.Base{}(nil), 调用BB方法， 直接指针方法调用， 但是内部的值都无法访问
// dd中base是指针，BB不是指针方法，  dd.base = main.Base{}(nil), 调用BB方法，自动解引用，但是base=nil,找不到BB地址，直接报错， 就好比入参是一个结构体，你穿指针肯定进不去

type AA interface {
	BB()
	CC()
}

type base struct {
	Name string
}

func (b base) BB() {
	fmt.Println("bb")
	//BB(b)

}

type DD struct {
	*base
}

func NewDD() AA {
	//d := &DD{}
	//addr := unsafe.Pointer(reflect.ValueOf(d).Pointer())
	return &DD{}
}

func (b DD) CC() {
	fmt.Println("ccc")
}

func TestAA(t *testing.T) {
	d := NewDD()

	d.BB()

}

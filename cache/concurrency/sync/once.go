package main

import "sync"

type MyBiz struct {
	once sync.Once
}

// 必须用指针，once是不能复制的，要不是指针方法，要不就是sync.once是指针
func (b *MyBiz) Init() {
	b.once.Do(func() {

	})
}

type MyBusiness interface {
	DoSomething()
}

type singleton struct {
}

func (s singleton) DoSomething() {
	return
}

var s *singleton
var singletonOnce sync.Once

func GetSingleton() MyBusiness {
	singletonOnce.Do(func() {
		s = &singleton{}
	})

	return s
}

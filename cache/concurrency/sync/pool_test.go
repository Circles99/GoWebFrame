package main

import (
	"sync"
	"testing"
)

func TestPool(t *testing.T) {
	p := sync.Pool{
		New: func() any {
			// 这里的any就是缓存的资源
			// 最好永远不要返回nil
			return "hello"
		},
	}

	str := p.Get().(string)
	defer p.Put(str)
	t.Log(str)
}

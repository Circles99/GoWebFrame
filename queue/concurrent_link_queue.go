package queue

import (
	"context"
	"errors"
	"sync/atomic"
	"unsafe"
)

// 无锁队列的核心：CAS + 自旋
// cas:compare and swap 比较并交换
// 占用CPU，极端情况下CPU占用100%
// 竞争越强，效果越差，CPU消耗越多（典型的随着竞争激烈而性能下降的手段）
// 使用场景：难以预估容量，高并发的地方
type ConcurrentLinkedQueue[T any] struct {
	head unsafe.Pointer
	tail unsafe.Pointer
}

func (c *ConcurrentLinkedQueue[T]) Enqueue(ctx context.Context, data T) error {
	// 入队
	newNode := &node[T]{
		val: data,
	}

	newNodePtr := unsafe.Pointer(newNode)

	// 这里的for循环是为了自旋
	// 比如2个G，1个G是28 一个G是30，指向到atomic的时候只有一个G能改，另一个G就只能自旋把前一个G的数据加载出来。放在另一个G后面
	for {
		tail := atomic.LoadPointer(&c.tail)
		// 切换指针引用
		// 把c.tail 从tail的指针换到newNodePtr上
		if atomic.CompareAndSwapPointer(&c.tail, tail, newNodePtr) {
			// 在这一步。就需要讲tail.next指向c.tail
			// 反向解析出来，才能拿到具体类型
			tailNode := (*node[T])(tail)
			atomic.StorePointer(&tailNode.next, newNodePtr)
			return nil
		}
	}

	// 先修改next，在修改tail

	//newPtr := unsafe.Pointer(newNode)
	//for {
	//	tailPtr := atomic.LoadPointer(&c.tail)
	//	tail := (*node[T])(tailPtr)
	//	tailNext := atomic.LoadPointer(&tail.next)
	//	if tailNext != nil {
	//		// 以及被人修改了，我们不需要修复，因为预期修改的那个人会把c.tail移过去
	//		continue
	//	}
	//
	//	// 把下一个指向的内存地址,指向为新的地址
	//	if atomic.CompareAndSwapPointer(&tail.next, tailNext, newPtr) {
	//
	//		// 把c.tail的指针换到新的指针上
	//		// 失败了也没关系，。说明有人抢先一步
	//		atomic.CompareAndSwapPointer(&c.tail, tailPtr, newPtr)
	//	}
	//
	//}

	// 这种是非线程安全的
	//c.tail.next = &node
	//c.tail = c.tail.next

}

func (c *ConcurrentLinkedQueue[T]) Dequeue(ctx context.Context) (T, error) {
	// 出队

	for {
		if ctx.Err() != nil {
			var t T
			return t, ctx.Err()
		}
		// 获取头部
		headPrt := atomic.LoadPointer(&c.head)
		head := (*node[T])(headPrt)

		tailPrt := atomic.LoadPointer(&c.tail)
		tail := (*node[T])(tailPrt)

		if tail == head {
			var t T
			return t, errors.New("队列为空")
		}

		// 直接head.next 和 atomic.LoadPointer(&head.next) 有什么区别
		nextHeadPtr := atomic.LoadPointer(&head.next)
		// 如果这里为空了， cas操作不会成功
		// 因为原本的数据被人拿走了
		if atomic.CompareAndSwapPointer(&c.head, headPrt, nextHeadPtr) {
			return head.val.(T), nil
		}

		//head := c.head
		//c.head = c.head.next

	}

}

func (c *ConcurrentLinkedQueue[T]) Len() uint64 {
	//TODO implement me
	panic("implement me")
}

func (c *ConcurrentLinkedQueue[T]) IsFull() bool {
	//TODO implement me
	panic("implement me")
}

func (c *ConcurrentLinkedQueue[T]) IsEmpty() bool {
	//TODO implement me
	panic("implement me")
}

type node[T any] struct {
	next unsafe.Pointer
	val  any
}

package queue

import (
	"context"
	"sync"
	"sync/atomic"
	"unsafe"
)

type ConcurrentBlockingQueue[T any] struct {
	mutex    *sync.Mutex
	data     []T
	notFull  chan struct{}
	notEmpty chan struct{}
	maxSize  int

	notFullCond  *Cond
	notEmptyCond *Cond

	count int
	head  int
	tail  int
	zero  T
}

func NewConcurrentBlockingQueue[T any](maxSize int) *ConcurrentBlockingQueue[T] {
	m := &sync.Mutex{}
	return &ConcurrentBlockingQueue[T]{
		data:         make([]T, maxSize),
		mutex:        m,
		notFull:      make(chan struct{}),
		notEmpty:     make(chan struct{}),
		maxSize:      maxSize,
		notFullCond:  NewCond(m),
		notEmptyCond: NewCond(m),
	}
}

func (c *ConcurrentBlockingQueue[T]) Enqueue(ctx context.Context, data any) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	c.mutex.Lock()
	//// 全是bug
	//// 这里需要自己释放锁，如果不是放，另外一边是拿不到
	//select {
	//case <-ctx.Done():
	//	c.mutex.Unlock()
	//	return ctx.Err()
	//default:
	//
	//}

	// 这里不用if 需要用for
	// 如果有多个G，比如G1和G2， 同时唤醒，G2直接入队，G1需要再次阻塞自己 所以使用for
	for c.isFull() {
		// todo chan实现，但有bug
		// 我阻塞住我自己，知道有人唤醒我
		//c.mutex.Unlock()
		//select {
		//case <-c.notFull:
		//	c.mutex.Lock()
		//case <-ctx.Done():
		//	return ctx.Err()
		//}
		err := c.notFullCond.WaitWithTimeout(ctx)
		if err != nil {
			return err
		}
	}

	c.data[c.tail] = data
	c.tail++
	c.count++
	if c.tail == c.maxSize {
		c.tail = 0
	}

	// 可能引起扩容
	//c.data = append(c.data, data)
	// 没有人等 notEmpty的新号，这一句会阻塞住
	//if len(c.data) == 1 {
	//	// 只有从空变不空发信号 chan实现
	//	//c.notEmpty <- struct{}{}
	//
	//
	//}

	// cond实现
	c.notEmptyCond.Broadcast()

	c.mutex.Unlock()
	return nil
}

func (c *ConcurrentBlockingQueue[T]) Dequeue(ctx context.Context) (any, error) {

	if ctx.Err() != nil {
		var t T
		return t, ctx.Err()
	}

	c.mutex.Lock()
	for c.isEmpty() {
		// chan实现，但有bug
		//// 阻塞我自己，等待元素入队
		//c.mutex.Unlock()
		//
		//// 进入select之前一定需要释放锁Unlock 不然导致死锁
		//// 阻塞在notEmpty没有关系，上面的Enqueue还能拿到锁，如果不释放，上面的Enqueue拿不到锁
		//select {
		//case <-c.notEmpty:
		//	c.mutex.Lock()
		//case <-ctx.Done():
		//	var t T
		//	return t, ctx.Err()
		//}
		err := c.notEmptyCond.WaitWithTimeout(ctx)
		if err != nil {
			var t T
			return t, err
		}
	}

	// 队首
	//t := c.data[0]
	//c.data = c.data[1:]
	// 直接使用 c.notFull <- struct{}{}
	// 如果上面根本没有人在select接收，会永远阻塞在这
	// chan实现
	//if len(c.data) == c.maxSize-1 {
	//	// 只有从满变不满发信号 chan实现
	//	//select {
	//	//case c.notFull <- struct{}{}:
	//	//default:
	//	//}
	//}

	// ring buffer原理实现
	t := c.data[c.head]

	// 当前下标数据已经返回出去了，需要填充个0值覆盖掉他
	// 不然外面有用户明明不使用这个A了，但是里面还有个引用指向A，不会被GC回收
	c.data[c.head] = c.zero
	c.head++
	c.count--
	if c.head == c.maxSize {
		c.head = 0
	}

	c.notFullCond.Broadcast()
	c.mutex.Unlock()

	// 没人等notFull 就会一直卡主
	return t, nil

}

func (c *ConcurrentBlockingQueue[T]) Len() uint64 {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return uint64(c.count)
}

func (c *ConcurrentBlockingQueue[T]) IsFull() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.isFull()
}

func (c *ConcurrentBlockingQueue[T]) isFull() bool {

	return c.count == c.maxSize
}

func (c *ConcurrentBlockingQueue[T]) IsEmpty() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.isEmpty()
}

func (c *ConcurrentBlockingQueue[T]) isEmpty() bool {
	return c.count == 0
}

//type cond struct {
//	*sync.Cond
//}
//
//func (c *cond) WaitTimeout(ctx context.Context) error {
//	ch := make(chan struct{})
//
//	go func() {
//		// 等待被唤醒
//		c.Cond.Wait()
//		// 唤醒之后尝试往ch发信号
//		select {
//		case ch <- struct{}{}:
//		default:
//			// 发不进去ch 开始走入default 代表超时返回了
//			// 转发这个信号
//			c.Cond.Signal()
//			// 需要解除锁，因为wait会lock，转发之后需要释放掉
//			c.Cond.L.Unlock()
//		}
//	}()
//
//	select {
//	case <-ctx.Done():
//		return ctx.Err()
//	case <-ch:
//		// 真的被唤醒了
//		return nil
//	}
//
//}

type Cond struct {
	L sync.Locker
	n unsafe.Pointer
}

func NewCond(l sync.Locker) *Cond {
	c := &Cond{
		L: l,
	}
	n := make(chan struct{})
	c.n = unsafe.Pointer(&n)
	return c
}

func (c Cond) Wait() {
	n := c.NotifyChan()
	c.L.Unlock()
	<-n
	c.L.Lock()
}

func (c Cond) WaitWithTimeout(ctx context.Context) error {
	n := c.NotifyChan()
	c.L.Unlock()
	select {
	case <-n:
		c.L.Lock()
		return nil
	case <-ctx.Done():
		c.L.Lock()
		return ctx.Err()
	}
}

func (c Cond) NotifyChan() <-chan struct{} {
	ptr := atomic.LoadPointer(&c.n)
	return *((*chan struct{})(ptr))
}

func (c Cond) Broadcast() {
	n := make(chan struct{})
	ptrOld := atomic.SwapPointer(&c.n, unsafe.Pointer(&n))
	close(*((*chan struct{})(ptrOld)))
}

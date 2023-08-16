package channel

import "context"

type Task func()

type TaskPool struct {
	tasks chan Task
	//close *atomic.Bool
	close chan struct{}
}

// numG goroutine的数量
// caps 缓存的容量
func NewTaskPool(numG int, caps int) *TaskPool {
	res := &TaskPool{
		tasks: make(chan Task, caps),
		//close: atomic.NewBool(false),
		close: make(chan struct{}),
	}
	for i := 0; i < numG; i++ {
		go func() {
			for {
				select {
				case <-res.close:
					return
				case t := <-res.tasks:
					t()
				}
			}

			//if res.close.Load() {
			//	return
			//}
			//
			//for t := range res.tasks {
			//	t()
			//}
		}()
	}
	return res
}

// Do 执行任务
// @author: liujiming
func (p TaskPool) Do(ctx context.Context, t Task) error {

	select {
	case p.tasks <- t:
	case <-ctx.Done():
		// 这里代表超时或者cancel
		return ctx.Err()
	}

	return nil
}

func (p TaskPool) Close() error {
	//p.close.Store(true)

	// 这种实现有缺陷，重复调用close方法 会panic
	close(p.close)
	return nil
}

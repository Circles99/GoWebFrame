package cache

import (
	"context"
	"log"
	"time"
)

//func TestLock(t *testing.T) {
//	var c *Client
//	lock, err := c.TryLock(context.Background(), "key1", time.Minute)
//
//	lock.UnLock()
//}

func ExampleLock_Refresh() {
	// 枷锁成功，拿到了一个lock
	var l *Lock

	stopChan := make(chan struct{})

	errChan := make(chan error)

	timeoutChan := make(chan struct{}, 1)

	go func() {
		// 间隔多久续约一次
		ticker := time.NewTicker(time.Second * 10)
		for {
			select {
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				// 出现了error咋 办
				err := l.Refresh(ctx)
				cancel()

				if err == context.DeadlineExceeded {
					timeoutChan <- struct{}{}
					continue
				}

				if err != nil {
					errChan <- err
					return
				}
			case <-timeoutChan:
				// 超时了重写调动刷新
				// 可以用次数控制住刷新
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				err := l.Refresh(ctx)
				cancel()

				if err == context.DeadlineExceeded {
					timeoutChan <- struct{}{}
					continue
				}

				if err != nil {
					errChan <- err
					return
				}

			case <-stopChan:

			}
		}
	}()

	// 业务执行过程中，检测errChan有无信号

	// 有循环则每次循环都检测一下
	for i := 0; i < 100; i++ {
		select {
		case <-errChan:
			break
		}
	}

	// 没有循环，则每个步骤内都要检测一下
	// 步骤1
	select {
	case err := <-errChan:
		log.Fatal(err)
	default:
	}

	// 步骤2
	select {
	case err := <-errChan:
		log.Fatal(err)
	default:
	}

	// 业务做完 发送信号
	stopChan <- struct{}{}
}

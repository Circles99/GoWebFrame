package net

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"
)

type Pool struct {
	//  空闲连接队列
	idlesConns chan *idleConn
	// 请求队列，也可以使用channel，但需要给个容量。这里使用切片
	reqQueue []connReq

	// 最大连接数
	maxCnt int

	// 当前连接数-已经建好的
	cnt int
	// 最大空闲时间
	maxIdleTime time.Duration
	//
	//// 初始连接数
	//initCnt int

	factory func() (net.Conn, error)

	lock *sync.Mutex
}

func NewPool(initCnt, maxIdleCnt, maxCnt int, maxIdleTime time.Duration, factory func() (net.Conn, error)) (*Pool, error) {

	if initCnt > maxIdleCnt {
		return nil, errors.New("初始连接数量不能大于最大空闲连接数")
	}

	idleConns := make(chan *idleConn, maxIdleCnt)
	// 初始化连接，放入到空闲队列中
	for i := 0; i < initCnt; i++ {
		conn, err := factory()
		if err != nil {
			return nil, err
		}

		idleConns <- &idleConn{c: conn, lastActiveTime: time.Now()}
	}

	res := &Pool{
		idlesConns:  make(chan *idleConn, maxIdleCnt),
		reqQueue:    nil,
		maxCnt:      maxCnt,
		cnt:         0,
		maxIdleTime: maxIdleTime,
		//initCnt:     initCnt,
		factory: factory,
	}

	return res, nil
}

func (p Pool) Get(ctx context.Context) (net.Conn, error) {
	select {
	case <-ctx.Done():
		// 超时了
		return nil, ctx.Err()
	default:

	}

	// 这里for循环是为了如果拿到一个连接发现已经需要关闭了，再去拿下一个
	for {
		select {
		case ic := <-p.idlesConns:
			// 拿到了空闲连接

			// 最大空闲时间小于你当前的时间，关闭他
			if ic.lastActiveTime.Add(p.maxIdleTime).Before(time.Now()) {
				_ = ic.c.Close()
				continue
			}
			return ic.c, nil
		default:
			// 没有空闲连接
			p.lock.Lock()
			if p.cnt >= p.maxCnt {
				// 超过上限了, 加入到请求队列中去等待
				req := connReq{connChan: make(chan net.Conn, 1)}
				p.reqQueue = append(p.reqQueue, req)
				p.lock.Unlock()

				select {
				case <-ctx.Done():
					// 这里也需要考虑超时
					// 选择1：删除请求队列中的这个值
					// 选择2：在这选择转发
					go func() {
						c := <-req.connChan
						// 重新开了个G，不能沿用已有的context
						// 失败了没事，就浪费一个连接
						_ = p.Put(context.Background(), c)
					}()
				case c := <-req.connChan:
					// 这个分支是等别人归还连接,因为是别人刚归还的连接，这里不检测，直接当他可使用

					return c, nil
				}
			}

			// 没超过上限
			c, err := p.factory()
			if err != nil {
				return nil, err
			}
			p.cnt++
			p.lock.Unlock()
			return c, nil

		}
	}

}

func (p Pool) Put(ctx context.Context, c net.Conn) error {
	p.lock.Lock()
	if len(p.reqQueue) > 0 {
		// 有阻塞的请求, 可从队首和队尾拿，这里使用队首
		req := p.reqQueue[0]
		// 取走了一个，需要重新给reqQueue赋值
		p.reqQueue = p.reqQueue[1:]
		p.lock.Unlock()
		req.connChan <- c
		return nil
	}
	p.lock.Unlock()
	// 没有阻塞的请求

	ic := &idleConn{c: c, lastActiveTime: time.Now()}
	select {

	case p.idlesConns <- ic:
	default:
		// 空闲队列满了
		_ = c.Close()

		p.lock.Lock()
		p.cnt--
		p.lock.Unlock()
	}
	return nil
}

type idleConn struct {
	c net.Conn
	// 上次使用时间
	lastActiveTime time.Time
}

type connReq struct {
	connChan chan net.Conn
}

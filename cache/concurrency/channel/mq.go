package channel

import (
	"errors"
	"sync"
)

type Broker struct {
	mutex sync.RWMutex
	chans []chan Msg
}

type Msg struct {
	Content string
}

func (b *Broker) Send(m Msg) error {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	for _, ch := range b.chans {
		// 这么写会直接阻塞住，当缓存用完以后也会直接阻塞住
		//ch <- m
		select {
		case ch <- m:
		default:
			return errors.New("消息队列已满")
		}
	}

	return nil
}

// <- 只读的chan
func (b *Broker) Subscribe(cap int) (<-chan Msg, error) {
	res := make(chan Msg, cap)
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.chans = append(b.chans, res)
	return res, nil
}

package memory

import (
	"GoWebFrame/session"
	"context"
	"errors"
	"github.com/patrickmn/go-cache"
	"github.com/spf13/cast"
	"sync"
	"time"
)

type Store struct {
	// 加锁保证同一个id被多个goroutine操作
	mutex *sync.RWMutex
	// 利用缓存来
	cache *cache.Cache
	// 过期时间
	expireTime time.Duration
}

func NewStore(expireTime time.Duration) *Store {
	return &Store{
		mutex:      &sync.RWMutex{},
		cache:      cache.New(expireTime, time.Second),
		expireTime: expireTime,
	}
}

func (s *Store) Generate(ctx context.Context, id string) (session.Session, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	sess := &Session{
		id: id,
		sm: &sync.Map{},
	}

	s.cache.Set(sess.ID(), sess, s.expireTime)
	return sess, nil
}

func (s *Store) Get(ctx context.Context, id string) (session.Session, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	sess, ok := s.cache.Get(id)
	if !ok {
		return nil, errors.New("sess 未找到")
	}
	return sess.(*Session), nil

}

func (s *Store) Remove(ctx context.Context, id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.cache.Delete(id)
	return nil
}

func (s *Store) Refresh(ctx context.Context, id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	// 刷新=重新set进去

	sess, ok := s.cache.Get(id)
	if !ok {
		return errors.New("sess 未找到")
	}

	s.cache.Set(id, sess, s.expireTime)
	return nil

}

type Session struct {
	sm *sync.Map
	id string
}

func (s Session) Get(ctx context.Context, key string) (string, error) {
	value, ok := s.sm.Load(key)
	if !ok {
		return "", errors.New("找不到这个key")
	}
	return cast.ToString(value), nil

}

func (s Session) Set(ctx context.Context, key, val string) error {
	s.sm.Store(key, val)
	return nil
}

func (s Session) ID() string {
	return s.id
}

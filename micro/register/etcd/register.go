package etcd

import (
	"GoWebFrame/micro/register"
	"context"
	"encoding/json"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/concurrency"
	"sync"
)

type Register struct {
	c       *clientv3.Client
	sess    *concurrency.Session
	cancels []func()
	mutex   sync.Mutex
}

func NewRegister(cc *clientv3.Client) (*Register, error) {
	sess, err := concurrency.NewSession(cc)
	if err != nil {
		return nil, err
	}
	return &Register{
		c:    cc,
		sess: sess,
	}, nil
}

func (r *Register) Register(ctx context.Context, si register.ServiceInstance) error {
	val, err := json.Marshal(si)
	if err != nil {
		return err
	}

	_, err = r.c.Put(ctx, r.instanceKey(si), string(val), clientv3.WithLease(r.sess.Lease()))
	return err
}

func (r *Register) UnRegister(ctx context.Context, si register.ServiceInstance) error {
	_, err := r.c.Delete(ctx, r.instanceKey(si))
	return err
}

func (r *Register) ListenServices(ctx context.Context, serviceName string) ([]register.ServiceInstance, error) {

	// 按前缀匹配
	resp, err := r.c.Get(ctx, r.serviceKey(serviceName), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	res := make([]register.ServiceInstance, 0, len(resp.Kvs))

	for _, kv := range resp.Kvs {
		var si register.ServiceInstance
		err = json.Unmarshal(kv.Value, &si)
		if err != nil {
			return nil, err
		}
		res = append(res, si)

	}
	return res, nil

}

func (r *Register) Subscribe(serviceName string) (<-chan register.Event, error) {
	// 监听以他为前缀的所有的key‘
	ctx, cancel := context.WithCancel(context.Background())

	r.mutex.Lock()
	r.cancels = append(r.cancels, cancel)
	r.mutex.Unlock()
	ctx = clientv3.WithRequireLeader(ctx)
	watchResp := r.c.Watch(ctx, r.serviceKey(serviceName), clientv3.WithPrefix())

	res := make(chan register.Event)

	go func() {
		for {
			select {
			case resp := <-watchResp:

				if resp.Err() != nil {
					return
				}

				//下面掉了cancel，因为传了ctx，这边会有cancel事件
				if resp.Canceled {
					return
				}

				// 返回的可能有很多个event事件
				for range resp.Events {
					res <- register.Event{}

				}

			case <-ctx.Done():
				// 借助ctx.done关闭这个goroutine
				// 调用cancel就会进入到这
				return
			}
		}
	}()

	return res, nil

}

func (r *Register) Close() error {
	r.mutex.Lock()
	cancels := r.cancels
	//重置掉cancels
	r.cancels = nil
	r.mutex.Unlock()
	// 逐步关闭上述所有G
	for _, cancel := range cancels {
		cancel()
	}

	// sess已关闭，租约结束
	err := r.sess.Close()
	if err != nil {
		return err
	}
	return nil
}

func (r *Register) instanceKey(si register.ServiceInstance) string {
	// 可以考虑说直接用Address
	// 也可以在si里引用一个instanceName字段
	return fmt.Sprintf("/micro/%s/%s", si.Name, si.Address)
}

func (r *Register) serviceKey(sname string) string {
	return fmt.Sprintf("/micro/%s", sname)
}

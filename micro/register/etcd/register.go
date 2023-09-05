package etcd

import (
	"GoWebFrame/micro/register"
	"context"
	"encoding/json"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/concurrency"
)

type Register struct {
	c    *clientv3.Client
	sess *concurrency.Session
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
	//TODO implement me
	panic("implement me")
}

func (r *Register) Subscribe(serviceName string) (<-chan register.Event, error) {
	//TODO implement me
	panic("implement me")
}

func (r *Register) Close() error {
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

func (r *Register) serviceKey(si register.ServiceInstance) string {
	return fmt.Sprintf("/micro/%s", si.Name)
}

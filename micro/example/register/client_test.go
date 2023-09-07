package register

import (
	"GoWebFrame/micro"
	"GoWebFrame/micro/example/gen"
	"GoWebFrame/micro/register/etcd"
	"context"
	"github.com/coreos/etcd/clientv3"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	c, err := clientv3.New(clientv3.Config{
		Endpoints: []string{},
	})
	require.NoError(t, err)
	r, err := etcd.NewRegister(c)

	require.NoError(t, err)

	rb, err := micro.NewRegisterBuilder(r, time.Second*3)

	require.NoError(t, err)
	client, err := micro.NewClient(micro.ClientRb(rb))
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	cc, err := client.Dail(ctx, "user-service")
	require.NoError(t, err)
	uc := gen.NewUserServiceClient(cc)
	resp, err := uc.GetById(ctx, &gen.GetByIdReq{Id: 1})
	require.NoError(t, err)
	t.Log(resp)

}

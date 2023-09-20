package route

import (
	"GoWebFrame/micro"
	"GoWebFrame/micro/example/gen"
	"GoWebFrame/micro/register/etcd"
	"GoWebFrame/micro/route"
	"GoWebFrame/micro/route/round_robin"
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

	client, err := micro.NewClient(micro.ClientWithRegistry(r, time.Second*3),
		micro.ClientWithPickerBuilder("GROUP_ROUND_ROBIN", &round_robin.Builder{
			Filter: route.GroupFilter{}.Build(),
		}))

	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	cc, err := client.Dail(context.WithValue(ctx, "group", "A"), "user-service")
	require.NoError(t, err)
	uc := gen.NewUserServiceClient(cc)

	for i := 0; i < 10; i++ {
		resp, err := uc.GetById(ctx, &gen.GetByIdReq{Id: 1})
		require.NoError(t, err)
		t.Log(resp)
	}
}

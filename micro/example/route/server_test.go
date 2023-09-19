package route

import (
	"GoWebFrame/micro"
	"GoWebFrame/micro/example/gen"
	"GoWebFrame/micro/register/etcd"
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
	"testing"
)

func TestServer(t *testing.T) {
	c, err := clientv3.New(clientv3.Config{
		Endpoints: []string{},
	})
	require.NoError(t, err)
	r, err := etcd.NewRegister(c)

	require.NoError(t, err)

	us := &UserServiceServer{}

	var eg errgroup.Group
	for i := 0; i < 3; i++ {
		var group = "A"
		if i%2 == 0 {
			group = "B"
		}

		server, err := micro.NewServer("user-service", micro.ServerWithRegister(r), micro.ServerWithGroup(group))
		require.NoError(t, err)

		gen.RegisterUserServiceServer(server, us)

		eg.Go(func() error {
			// 这里调用start方法，代表us完全准备好了
			return server.Start(fmt.Sprintf(":808%d", i+1))
		})
	}
	err = eg.Wait()
	require.NoError(t, err)
}

type UserServiceServer struct {
	gen.UnimplementedUserServiceServer
}

func (s UserServiceServer) GetById(ctx context.Context, req *gen.GetByIdReq) (*gen.GetByIdResp, error) {
	fmt.Println(req)
	return &gen.GetByIdResp{User: &gen.User{
		Name: "hello word",
	}}, nil
}

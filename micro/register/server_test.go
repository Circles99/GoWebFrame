package register

import (
	"GoWebFrame/micro"
	"GoWebFrame/micro/example/gen"
	"GoWebFrame/micro/register/etcd"
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/stretchr/testify/require"
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

	server, err := micro.NewServer("user-service", micro.ServerWithRegister(r))
	require.NoError(t, err)

	gen.RegisterUserServiceServer(server, us)

	// 这里调用start方法，代表us完全准备好了
	err = server.Start(":8081")
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

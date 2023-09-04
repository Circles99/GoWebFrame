package grpc_resolver

import (
	"GoWebFrame/micro/example/gen"
	"context"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	cc, err := grpc.Dial("register:///localhost:8081", grpc.WithInsecure(), grpc.WithResolvers(&Builder{}))
	require.NoError(t, err)
	client := gen.NewUserServiceClient(cc)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	resp, err := client.GetById(ctx, &gen.GetByIdReq{Id: 1})
	require.NoError(t, err)
	t.Log(resp)

}

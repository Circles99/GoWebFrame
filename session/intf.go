package session

import (
	"context"
	"net/http"
)

// Session session本体， 真实存储用户数据
type Session interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, val string) error
	ID() string
}

// Store 用来管来Session, 生成， 刷新， 删除，查找
type Store interface {
	Generate(ctx context.Context, id string) (Session, error)
	Get(ctx context.Context, id string) (Session, error)
	Remove(ctx context.Context, id string) error
	Refresh(ctx context.Context, id string) error
}

// Propagator SessionId的存储和提取的抽象，返回在http中存在于不同的地方
type Propagator interface {

	// Inject 讲sessionId注入到response中， 必须是幂等
	Inject(id string, writer http.ResponseWriter) error
	// Extract 从request中获取sessionId
	Extract(req *http.Request) (string, error)
	// Remove 删除 SessionId 在response中
	Remove(writer http.ResponseWriter) error
}

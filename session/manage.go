package session

import (
	"GoWebFrame"
	"GoWebFrame/session/intf"
)

type Manager struct {
	intf.Store
	intf.Propagator
	SessCtxKey string
}

// GetSession 将会尝试从 ctx 中拿到 Session，
// 如果成功了，那么它会将 Session 实例缓存到 ctx 的 UserValues 里面
func (m *Manager) GetSession(ctx *GoWebFrame.Context) (intf.Session, error) {
	if ctx.UserValues == nil {
		ctx.UserValues = make(map[string]any, 1)
	}
	// 缓存
	val, ok := ctx.UserValues[m.SessCtxKey]
	if ok {
		return val.(intf.Session), nil
	}
	id, err := m.Extract(ctx.Req)
	if err != nil {
		return nil, err
	}
	sess, err := m.Get(ctx.Req.Context(), id)
	if err != nil {
		return nil, err
	}
	ctx.UserValues[m.SessCtxKey] = sess
	return sess, nil
}

// InitSession 初始化一个 session，并且注入到 http response 里面
func (m *Manager) InitSession(ctx *GoWebFrame.Context, id string) (intf.Session, error) {
	sess, err := m.Generate(ctx.Req.Context(), id)
	if err != nil {
		return nil, err
	}
	if err = m.Inject(id, ctx.Resp); err != nil {
		return nil, err
	}
	return sess, nil
}

// RefreshSession 刷新 Session
func (m *Manager) RefreshSession(ctx *GoWebFrame.Context) (intf.Session, error) {
	sess, err := m.GetSession(ctx)
	if err != nil {
		return nil, err
	}
	// 刷新存储的过期时间
	err = m.Refresh(ctx.Req.Context(), sess.ID())
	if err != nil {
		return nil, err
	}
	// 重新注入 HTTP 里面
	if err = m.Inject(sess.ID(), ctx.Resp); err != nil {
		return nil, err
	}
	return sess, nil
}

// RemoveSession 删除 Session
func (m *Manager) RemoveSession(ctx *GoWebFrame.Context) error {
	sess, err := m.GetSession(ctx)
	if err != nil {
		return err
	}
	err = m.Store.Remove(ctx.Req.Context(), sess.ID())
	if err != nil {
		return err
	}
	return m.Propagator.Remove(ctx.Resp)
}

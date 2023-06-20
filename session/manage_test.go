package session

import (
	"GoWebFrame"
	"GoWebFrame/session/cookie"
	lRedis "GoWebFrame/session/redis"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"net/http"
	"testing"
)

func TestManager(t *testing.T) {
	s := GoWebFrame.NewHttpServer()

	p := cookie.NewPropagator("sessid")

	p.CustomCookieOption(func(c *http.Cookie) {
		c.HttpOnly = true
	})

	m := Manager{
		SessCtxKey: "_sess",
		Store: lRedis.NewStore(redis.NewClient(&redis.Options{
			Addr:     "127.0.0.1:6379",
			DB:       1,
			Password: "",
		})),
		Propagator: p,
	}

	s.Get("/login", func(ctx *GoWebFrame.Context) {
		// 前面就是你登录的时候一大堆的登录校验
		id := uuid.New()
		sess, err := m.InitSession(ctx, id.String())
		if err != nil {
			ctx.RespStatusCode = http.StatusInternalServerError
			return
		}
		// 然后根据自己的需要设置
		err = sess.Set(ctx.Req.Context(), "mykey", "some value")
		if err != nil {
			ctx.RespStatusCode = http.StatusInternalServerError
			return
		}
	})
	s.Get("/resource", func(ctx *GoWebFrame.Context) {
		sess, err := m.GetSession(ctx)
		if err != nil {
			ctx.RespStatusCode = http.StatusInternalServerError
			return
		}
		val, err := sess.Get(ctx.Req.Context(), "mykey")
		ctx.RespData = []byte(val)
	})

	s.Get("/logout", func(ctx *GoWebFrame.Context) {
		_ = m.RemoveSession(ctx)
	})

	s.Use("POST", "/logout", func(next GoWebFrame.HandleFunc) GoWebFrame.HandleFunc {
		return func(ctx *GoWebFrame.Context) {
			// 执行校验
			if ctx.Req.URL.Path != "/login" {
				sess, err := m.GetSession(ctx)
				// 不管发生了什么错误，对于用户我们都是返回未授权
				if err != nil {
					ctx.RespStatusCode = http.StatusUnauthorized
					return
				}
				ctx.UserValues["sess"] = sess
				_ = m.Refresh(ctx.Req.Context(), sess.ID())
			}
			next(ctx)
		}
	})

	s.Start(":8081")
}

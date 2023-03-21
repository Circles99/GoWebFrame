package cookie

import "net/http"

type Propagator struct {
	CookieName string
	cookieOpt  func(c *http.Cookie)
}

type PropagatorOptions func(propagator *Propagator)

func WithCookieOption(opt func(c *http.Cookie)) PropagatorOptions {
	return func(propagator *Propagator) {
		propagator.cookieOpt = opt
	}
}
func NewPropagator(CookieName string, options ...PropagatorOptions) *Propagator {
	return &Propagator{
		CookieName: CookieName,
		cookieOpt:  func(c *http.Cookie) {},
	}
}

func (p *Propagator) CustomCookieOption(opt func(c *http.Cookie)) {
	p.cookieOpt = opt
}

// Inject 回写时
func (p *Propagator) Inject(id string, writer http.ResponseWriter) error {
	cookie := &http.Cookie{Name: p.CookieName, Value: id}
	p.cookieOpt(cookie)
	http.SetCookie(writer, cookie)
	return nil
}

// Extract 请求进来时
func (p *Propagator) Extract(req *http.Request) (string, error) {
	c, err := req.Cookie(p.CookieName)
	if err != nil {
		return "", err
	}
	return c.Value, nil
}

func (p Propagator) Remove(writer http.ResponseWriter) error {
	// -1 代表马上过期
	cookie := &http.Cookie{Name: p.CookieName, MaxAge: -1}
	p.cookieOpt(cookie)
	http.SetCookie(writer, cookie)
	return nil
}

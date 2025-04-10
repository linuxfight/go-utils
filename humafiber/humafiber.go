package humafiber

import (
	"bytes"
	"context"
	"crypto/tls"
	"github.com/gofiber/fiber/v3"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
)

// Custom context wrapper to merge Fiber user values with standard context
type fiberContextWrapper struct {
	context.Context
	fiberCtx fiber.Ctx
}

func (w *fiberContextWrapper) Value(key interface{}) interface{} {
	// First, check standard context
	if val := w.Context.Value(key); val != nil {
		return val
	}

	if w.fiberCtx.RequestCtx() != nil {
		// Second, check Fiber's local storage
		if val := w.fiberCtx.Locals(key); val != nil {
			return val
		}
	}

	// return nil
	return nil
}

// Unwrap extracts the underlying Fiber context from a Huma context. If passed a
// context from a different adapter it will panic. Keep in mind the limitations
// of the underlying Fiber/fasthttp libraries and how that impacts
// memory-safety: https://docs.gofiber.io/#zero-allocation. Do not keep
// references to the underlying context or its values!
func Unwrap(ctx huma.Context) fiber.Ctx {
	if c, ok := ctx.(*fiberWrapper); ok {
		return c.Unwrap()
	}
	panic("not a humafiber context")
}

type fiberAdapter struct {
	tester requestTester
	router router
}

type fiberWrapper struct {
	op     *huma.Operation
	status int
	orig   fiber.Ctx
	ctx    context.Context
}

// check that fiberCtx implements huma.Context
var _ huma.Context = &fiberWrapper{}

func (c *fiberWrapper) Unwrap() fiber.Ctx {
	return c.orig
}

func (c *fiberWrapper) Operation() *huma.Operation {
	return c.op
}

func (c *fiberWrapper) Matched() string {
	return c.orig.Route().Path
}

func (c *fiberWrapper) Context() context.Context {
	// Return our merged context wrapper
	return &fiberContextWrapper{
		Context:  c.orig.Context(),
		fiberCtx: c.orig,
	}
}

func (c *fiberWrapper) Method() string {
	return c.orig.Method()
}

func (c *fiberWrapper) Host() string {
	return c.orig.Hostname()
}

func (c *fiberWrapper) RemoteAddr() string {
	return c.orig.RequestCtx().RemoteAddr().String()
}

func (c *fiberWrapper) URL() url.URL {
	u, _ := url.Parse(string(c.orig.Request().RequestURI()))
	return *u
}

func (c *fiberWrapper) Param(name string) string {
	return c.orig.Params(name)
}

func (c *fiberWrapper) Query(name string) string {
	return c.orig.Query(name)
}

func (c *fiberWrapper) Header(name string) string {
	return c.orig.Get(name)
}

func (c *fiberWrapper) EachHeader(cb func(name, value string)) {
	c.orig.Request().Header.VisitAll(func(k, v []byte) {
		cb(string(k), string(v))
	})
}

func (c *fiberWrapper) BodyReader() io.Reader {
	var orig = c.orig
	if orig.App().Server().StreamRequestBody {
		// Streaming is enabled, so send the reader.
		return orig.Request().BodyStream()
	}
	return bytes.NewReader(orig.BodyRaw())
}

func (c *fiberWrapper) GetMultipartForm() (*multipart.Form, error) {
	return c.orig.MultipartForm()
}

func (c *fiberWrapper) SetReadDeadline(deadline time.Time) error {
	// Note: for this to work properly you need to do two things:
	// 1. Set the Fiber app's `StreamRequestBody` to `true`
	// 2. Set the Fiber app's `BodyLimit` to some small value like `1`
	// Fiber will only call the request handler for streaming once the limit is
	// reached. This is annoying but currently how things work.
	return c.orig.RequestCtx().Conn().SetReadDeadline(deadline)
}

func (c *fiberWrapper) SetStatus(code int) {
	var orig = c.orig
	c.status = code
	orig.Status(code)
}

func (c *fiberWrapper) Status() int {
	return c.status
}
func (c *fiberWrapper) AppendHeader(name string, value string) {
	c.orig.Append(name, value)
}

func (c *fiberWrapper) SetHeader(name string, value string) {
	c.orig.Set(name, value)
}

func (c *fiberWrapper) BodyWriter() io.Writer {
	return c.orig.Response().BodyWriter()
}

func (c *fiberWrapper) TLS() *tls.ConnectionState {
	return c.orig.RequestCtx().TLSConnectionState()
}

func (c *fiberWrapper) Version() huma.ProtoVersion {
	return huma.ProtoVersion{
		Proto: c.orig.Protocol(),
	}
}

type router interface {
	Add(method, path string, handlers ...fiber.Handler) fiber.Router
}

type requestTester interface {
	Test(*http.Request, ...int) (*http.Response, error)
}

type contextWrapperValue struct {
	Key   any
	Value any
}

type contextWrapper struct {
	values []*contextWrapperValue
	context.Context
}

var (
	_ context.Context = &contextWrapper{}
)

func (c *contextWrapper) Value(key any) any {
	var raw = c.Context.Value(key)
	if raw != nil {
		return raw
	}
	for _, pair := range c.values {
		if pair.Key == key {
			return pair.Value
		}
	}
	return nil
}

func (a *fiberAdapter) Handle(op *huma.Operation, handler func(huma.Context)) {
	// Convert {param} to :param
	path := op.Path
	path = strings.ReplaceAll(path, "{", ":")
	path = strings.ReplaceAll(path, "}", "")
	a.router.Add(op.Method, path, func(c fiber.Ctx) error {
		var values []*contextWrapperValue
		c.RequestCtx().VisitUserValuesAll(func(key, value any) {
			values = append(values, &contextWrapperValue{
				Key:   key,
				Value: value,
			})
		})
		handler(&fiberWrapper{
			op:   op,
			orig: c,
			ctx: &contextWrapper{
				values:  values,
				Context: c.Context(),
			},
		})
		return nil
	})
}

func (a *fiberAdapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// b, _ := httputil.DumpRequest(r, true)
	// fmt.Println(string(b))
	resp, err := a.tester.Test(r)
	if resp != nil && resp.Body != nil {
		defer func() {
			_ = resp.Body.Close()
		}()
	}
	if err != nil {
		panic(err)
	}
	h := w.Header()
	for k, v := range resp.Header {
		for item := range v {
			h.Add(k, v[item])
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func New(r *fiber.App, config huma.Config) huma.API {
	return huma.NewAPI(config, &fiberAdapter{
		tester: &fiberRequestTester{app: r},
		router: &fiberRouterAdapter{router: r},
	})
}

func NewWithGroup(r *fiber.App, g fiber.Router, config huma.Config) huma.API {
	return huma.NewAPI(config, &fiberAdapter{
		tester: &fiberRequestTester{app: r},
		router: &fiberRouterAdapter{router: g},
	})
}

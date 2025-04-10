package humafiber

import (
	"github.com/gofiber/fiber/v3"
	"net/http"
	"time"
)

// Create adapter for Fiber's Test method
type fiberRequestTester struct {
	app *fiber.App
}

func (frt *fiberRequestTester) Test(req *http.Request, msTimeout ...int) (*http.Response, error) {
	// Convert timeout to Fiber's TestConfig format
	cfg := fiber.TestConfig{}
	if len(msTimeout) > 0 {
		cfg.Timeout = time.Duration(msTimeout[0]) * time.Millisecond
	}
	return frt.app.Test(req, cfg)
}

// Create adapter for Fiber's router interface
type fiberRouterAdapter struct {
	router fiber.Router
}

func (fra *fiberRouterAdapter) Add(method, path string, handlers ...fiber.Handler) fiber.Router {
	// Convert Huma's Add to Fiber's routing methods
	switch method {
	case http.MethodGet:
		router := fra.router
		for _, h := range handlers {
			router = router.Get(path, h)
		}
		return router
	case http.MethodPost:
		router := fra.router
		for _, h := range handlers {
			router = router.Post(path, h)
		}
		return router
	case http.MethodPut:
		router := fra.router
		for _, h := range handlers {
			router = router.Put(path, h)
		}
		return router
	case http.MethodDelete:
		router := fra.router
		for _, h := range handlers {
			router = router.Delete(path, h)
		}
		return router
	case http.MethodPatch:
		router := fra.router
		for _, h := range handlers {
			router = router.Patch(path, h)
		}
		return router
	case http.MethodHead:
		router := fra.router
		for _, h := range handlers {
			router = router.Head(path, h)
		}
		return router
	case http.MethodOptions:
		router := fra.router
		for _, h := range handlers {
			router = router.Options(path, h)
		}
		return router
	default:
		methods := append([]string{}, method)
		router := fra.router
		for _, h := range handlers {
			router = router.Add(methods, path, h)
		}
		return router
	}
}

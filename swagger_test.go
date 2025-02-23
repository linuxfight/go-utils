package middlewares

import (
	_ "embed"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
)

//go:embed swagger.json
var swaggerJSON []byte

//go:embed swagger.yaml
var swaggerYAML []byte

func performRequest(method, target string, app *fiber.App) *http.Response {
	r := httptest.NewRequest(method, target, nil)
	resp, _ := app.Test(r)
	return resp
}

func TestNew(t *testing.T) {
	t.Run("Endpoint check with only custom path", func(t *testing.T) {
		app := fiber.New()

		cfg := SwaggerConfig{
			Path: "custompath",
		}
		app.Use(NewSwagger(cfg))

		w1 := performRequest("GET", "/custompath", app)
		require.Equal(t, 200, w1.StatusCode)

		w2 := performRequest("GET", "/swagger.json", app)
		require.Equal(t, 200, w2.StatusCode)

		w3 := performRequest("GET", "/notfound", app)
		require.Equal(t, 404, w3.StatusCode)
	})

	t.Run("Endpoint check with only custom basepath", func(t *testing.T) {
		app := fiber.New()

		cfg := SwaggerConfig{
			BasePath: "/api/v1",
		}
		app.Use(NewSwagger(cfg))

		w1 := performRequest("GET", "/api/v1/docs", app)
		require.Equal(t, 200, w1.StatusCode)

		w2 := performRequest("GET", "/api/v1/swagger.json", app)
		require.Equal(t, 200, w2.StatusCode)

		w3 := performRequest("GET", "/notfound", app)
		require.Equal(t, 404, w3.StatusCode)
	})

	t.Run("Endpoint check with custom config", func(t *testing.T) {
		app := fiber.New()

		cfg := SwaggerConfig{
			BasePath: "/",
			FilePath: "swagger.json",
		}
		app.Use(NewSwagger(cfg))

		w1 := performRequest("GET", "/docs", app)
		require.Equal(t, 200, w1.StatusCode)

		w2 := performRequest("GET", "/swagger.json", app)
		require.Equal(t, 200, w2.StatusCode)

		w3 := performRequest("GET", "/notfound", app)
		require.Equal(t, 404, w3.StatusCode)
	})

	t.Run("Endpoint check with custom path", func(t *testing.T) {
		app := fiber.New()

		cfg := SwaggerConfig{
			BasePath: "/",
			FilePath: "swagger.json",
			Path:     "swagger",
		}
		app.Use(NewSwagger(cfg))

		w1 := performRequest("GET", "/swagger", app)
		require.Equal(t, 200, w1.StatusCode)

		w2 := performRequest("GET", "/swagger.json", app)
		require.Equal(t, 200, w2.StatusCode)

		w3 := performRequest("GET", "/notfound", app)
		require.Equal(t, 404, w3.StatusCode)
	})

	t.Run("Endpoint check with custom config and yaml spec", func(t *testing.T) {
		app := fiber.New()

		cfg := SwaggerConfig{
			BasePath: "/",
			FilePath: "./swagger.yaml",
		}
		app.Use(NewSwagger(cfg))

		w1 := performRequest("GET", "/docs", app)
		require.Equal(t, 200, w1.StatusCode)

		w2 := performRequest("GET", "/swagger.yaml", app)
		require.Equal(t, 200, w2.StatusCode)

		w3 := performRequest("GET", "/notfound", app)
		require.Equal(t, 404, w3.StatusCode)
	})

	t.Run("Endpoint check with custom path and yaml spec", func(t *testing.T) {
		app := fiber.New()

		cfg := SwaggerConfig{
			BasePath: "/",
			FilePath: "swagger.yaml",
			Path:     "swagger",
		}
		app.Use(NewSwagger(cfg))

		w1 := performRequest("GET", "/swagger", app)
		require.Equal(t, 200, w1.StatusCode)

		w2 := performRequest("GET", "/swagger.yaml", app)
		require.Equal(t, 200, w2.StatusCode)

		w3 := performRequest("GET", "/notfound", app)
		require.Equal(t, 404, w3.StatusCode)
	})

	t.Run("Endpoint check with empty custom config", func(t *testing.T) {
		app := fiber.New()

		cfg := SwaggerConfig{}

		app.Use(NewSwagger(cfg))

		w1 := performRequest("GET", "/docs", app)
		require.Equal(t, 200, w1.StatusCode)

		w2 := performRequest("GET", "/swagger.json", app)
		require.Equal(t, 200, w2.StatusCode)

		w3 := performRequest("GET", "/notfound", app)
		require.Equal(t, 404, w3.StatusCode)
	})

	t.Run("Endpoint check with default config", func(t *testing.T) {
		app := fiber.New()

		app.Use(NewSwagger())

		w1 := performRequest("GET", "/docs", app)
		require.Equal(t, 200, w1.StatusCode)

		w2 := performRequest("GET", "/swagger.json", app)
		require.Equal(t, 200, w2.StatusCode)

		w3 := performRequest("GET", "/notfound", app)
		require.Equal(t, 404, w3.StatusCode)
	})

	t.Run("Swagger.json file is not exist", func(t *testing.T) {
		app := fiber.New()

		cfg := SwaggerConfig{
			FilePath: "./docs/swagger.json",
		}

		require.Panics(t, func() {
			app.Use(NewSwagger(cfg))
		}, "/swagger.json file is not exist")
	})

	t.Run("Swagger.json missing file", func(t *testing.T) {
		app := fiber.New()

		cfg := SwaggerConfig{
			FilePath: "./docs/swagger_missing.json",
		}

		require.Panics(t, func() {
			app.Use(NewSwagger(cfg))
		}, "invalid character ':' after object key:value pair")
	})

	t.Run("Endpoint check with multiple Swagger instances", func(t *testing.T) {
		app := fiber.New()

		app.Use(NewSwagger(SwaggerConfig{
			BasePath: "/api/v1",
		}))

		app.Use(NewSwagger(SwaggerConfig{
			BasePath: "/api/v2",
		}))

		w1 := performRequest("GET", "/api/v1/docs", app)
		require.Equal(t, 200, w1.StatusCode)

		w2 := performRequest("GET", "/api/v1/swagger.json", app)
		require.Equal(t, 200, w2.StatusCode)

		w3 := performRequest("GET", "/api/v2/docs", app)
		require.Equal(t, 200, w3.StatusCode)

		w4 := performRequest("GET", "/api/v2/swagger.json", app)
		require.Equal(t, 200, w4.StatusCode)

		w5 := performRequest("GET", "/notfound", app)
		require.Equal(t, 404, w5.StatusCode)
	})

	t.Run("Endpoint check with custom routes", func(t *testing.T) {
		app := fiber.New()

		app.Use(NewSwagger(SwaggerConfig{
			BasePath: "/api/v1",
		}))

		app.Get("/api/v1/tasks", func(c fiber.Ctx) error {
			return c.SendString("success")
		})

		app.Get("/api/v1", func(c fiber.Ctx) error {
			return c.SendString("success")
		})

		w1 := performRequest("GET", "/api/v1/docs", app)
		require.Equal(t, 200, w1.StatusCode)

		w2 := performRequest("GET", "/api/v1/swagger.json", app)
		require.Equal(t, 200, w2.StatusCode)

		w3 := performRequest("GET", "/notfound", app)
		require.Equal(t, 404, w3.StatusCode)

		// Verify we can send request to handler with the same BasePath as the middleware
		w4 := performRequest("GET", "/api/v1/tasks", app)
		bodyBytes, err := io.ReadAll(w4.Body)
		require.NoError(t, err)
		require.Equal(t, 200, w4.StatusCode)
		require.Equal(t, "success", string(bodyBytes))

		// Verify handler in BasePath still works
		w5 := performRequest("GET", "/api/v1", app)
		bodyBytes, err = io.ReadAll(w5.Body)
		require.NoError(t, err)
		require.Equal(t, 200, w5.StatusCode)
		require.Equal(t, "success", string(bodyBytes))

		w6 := performRequest("GET", "/api/v1/", app)
		bodyBytes, err = io.ReadAll(w6.Body)
		require.NoError(t, err)
		require.Equal(t, 200, w6.StatusCode)
		require.Equal(t, "success", string(bodyBytes))
	})
}

func TestNewWithFileContent(t *testing.T) {
	t.Run("Endpoint check with only custom path", func(t *testing.T) {
		app := fiber.New()

		cfg := SwaggerConfig{
			Path:        "custompath",
			FileContent: swaggerJSON,
			FilePath:    "doesnotexist-swagger.json",
		}
		app.Use(NewSwagger(cfg))

		w1 := performRequest("GET", "/custompath", app)
		require.Equal(t, 200, w1.StatusCode)

		w2 := performRequest("GET", "/doesnotexist-swagger.json", app)
		require.Equal(t, 200, w2.StatusCode)

		w3 := performRequest("GET", "/notfound", app)
		require.Equal(t, 404, w3.StatusCode)
	})

	t.Run("Endpoint check with only custom basepath", func(t *testing.T) {
		app := fiber.New()

		cfg := SwaggerConfig{
			BasePath:    "/api/v1",
			FileContent: swaggerJSON,
			FilePath:    "doesnotexist-swagger.json",
		}
		app.Use(NewSwagger(cfg))

		w1 := performRequest("GET", "/api/v1/docs", app)
		require.Equal(t, 200, w1.StatusCode)

		w2 := performRequest("GET", "/api/v1/doesnotexist-swagger.json", app)
		require.Equal(t, 200, w2.StatusCode)

		w3 := performRequest("GET", "/notfound", app)
		require.Equal(t, 404, w3.StatusCode)
	})

	t.Run("Endpoint check with custom config", func(t *testing.T) {
		app := fiber.New()

		cfg := SwaggerConfig{
			BasePath:    "/",
			FilePath:    "doesnotexist-swagger.json",
			FileContent: swaggerJSON,
		}
		app.Use(NewSwagger(cfg))

		w1 := performRequest("GET", "/docs", app)
		require.Equal(t, 200, w1.StatusCode)

		w2 := performRequest("GET", "/doesnotexist-swagger.json", app)
		require.Equal(t, 200, w2.StatusCode)

		w3 := performRequest("GET", "/notfound", app)
		require.Equal(t, 404, w3.StatusCode)
	})

	t.Run("Endpoint check with custom path", func(t *testing.T) {
		app := fiber.New()

		cfg := SwaggerConfig{
			BasePath:    "/",
			FilePath:    "doesnotexist-swagger.json",
			Path:        "swagger",
			FileContent: swaggerJSON,
		}
		app.Use(NewSwagger(cfg))

		w1 := performRequest("GET", "/swagger", app)
		require.Equal(t, 200, w1.StatusCode)

		w2 := performRequest("GET", "/doesnotexist-swagger.json", app)
		require.Equal(t, 200, w2.StatusCode)

		w3 := performRequest("GET", "/notfound", app)
		require.Equal(t, 404, w3.StatusCode)
	})

	t.Run("Endpoint check with custom config and yaml spec", func(t *testing.T) {
		app := fiber.New()

		cfg := SwaggerConfig{
			BasePath:    "/",
			FilePath:    "./doesnotexist-swagger.yaml",
			FileContent: swaggerYAML,
		}
		app.Use(NewSwagger(cfg))

		w1 := performRequest("GET", "/docs", app)
		require.Equal(t, 200, w1.StatusCode)

		w2 := performRequest("GET", "/doesnotexist-swagger.yaml", app)
		require.Equal(t, 200, w2.StatusCode)

		w3 := performRequest("GET", "/notfound", app)
		require.Equal(t, 404, w3.StatusCode)
	})

	t.Run("Endpoint check with custom path and yaml spec", func(t *testing.T) {
		app := fiber.New()

		cfg := SwaggerConfig{
			BasePath:    "/",
			FilePath:    "doesnotexist-swagger.yaml",
			Path:        "swagger",
			FileContent: swaggerYAML,
		}
		app.Use(NewSwagger(cfg))

		w1 := performRequest("GET", "/swagger", app)
		require.Equal(t, 200, w1.StatusCode)

		w2 := performRequest("GET", "/doesnotexist-swagger.yaml", app)
		require.Equal(t, 200, w2.StatusCode)

		w3 := performRequest("GET", "/notfound", app)
		require.Equal(t, 404, w3.StatusCode)
	})

	t.Run("Endpoint check with empty custom config", func(t *testing.T) {
		app := fiber.New()

		cfg := SwaggerConfig{
			FileContent: swaggerJSON,
			FilePath:    "doesnotexist-swagger.json",
		}

		app.Use(NewSwagger(cfg))

		w1 := performRequest("GET", "/docs", app)
		require.Equal(t, 200, w1.StatusCode)

		w2 := performRequest("GET", "/doesnotexist-swagger.json", app)
		require.Equal(t, 200, w2.StatusCode)

		w3 := performRequest("GET", "/notfound", app)
		require.Equal(t, 404, w3.StatusCode)
	})

	t.Run("Swagger file content not specified", func(t *testing.T) {
		app := fiber.New()

		cfg := SwaggerConfig{
			FilePath: "./docs/swagger.json",
		}

		require.Panics(t, func() {
			app.Use(NewSwagger(cfg))
		}, "content not specified")
	})

	t.Run("Endpoint check with multiple Swagger instances", func(t *testing.T) {
		app := fiber.New()

		app.Use(NewSwagger(SwaggerConfig{
			BasePath:    "/api/v1",
			FileContent: swaggerJSON,
			FilePath:    "doesnotexist-swagger.json",
		}))

		app.Use(NewSwagger(SwaggerConfig{
			BasePath:    "/api/v2",
			FileContent: swaggerJSON,
			FilePath:    "doesnotexist-swagger.json",
		}))

		w1 := performRequest("GET", "/api/v1/docs", app)
		require.Equal(t, 200, w1.StatusCode)

		w2 := performRequest("GET", "/api/v1/doesnotexist-swagger.json", app)
		require.Equal(t, 200, w2.StatusCode)

		w3 := performRequest("GET", "/api/v2/docs", app)
		require.Equal(t, 200, w3.StatusCode)

		w4 := performRequest("GET", "/api/v2/doesnotexist-swagger.json", app)
		require.Equal(t, 200, w4.StatusCode)

		w5 := performRequest("GET", "/notfound", app)
		require.Equal(t, 404, w5.StatusCode)
	})

	t.Run("Endpoint check with custom routes", func(t *testing.T) {
		app := fiber.New()

		app.Use(NewSwagger(SwaggerConfig{
			BasePath:    "/api/v1",
			FileContent: swaggerJSON,
			FilePath:    "doesnotexist-swagger.json",
		}))

		app.Get("/api/v1/tasks", func(c fiber.Ctx) error {
			return c.SendString("success")
		})

		app.Get("/api/v1", func(c fiber.Ctx) error {
			return c.SendString("success")
		})

		w1 := performRequest("GET", "/api/v1/docs", app)
		require.Equal(t, 200, w1.StatusCode)

		w2 := performRequest("GET", "/api/v1/doesnotexist-swagger.json", app)
		require.Equal(t, 200, w2.StatusCode)

		w3 := performRequest("GET", "/notfound", app)
		require.Equal(t, 404, w3.StatusCode)

		// Verify we can send request to handler with the same BasePath as the middleware
		w4 := performRequest("GET", "/api/v1/tasks", app)
		bodyBytes, err := io.ReadAll(w4.Body)
		require.NoError(t, err)
		require.Equal(t, 200, w4.StatusCode)
		require.Equal(t, "success", string(bodyBytes))

		// Verify handler in BasePath still works
		w5 := performRequest("GET", "/api/v1", app)
		bodyBytes, err = io.ReadAll(w5.Body)
		require.NoError(t, err)
		require.Equal(t, 200, w5.StatusCode)
		require.Equal(t, "success", string(bodyBytes))

		w6 := performRequest("GET", "/api/v1/", app)
		bodyBytes, err = io.ReadAll(w6.Body)
		require.NoError(t, err)
		require.Equal(t, 200, w6.StatusCode)
		require.Equal(t, "success", string(bodyBytes))
	})
}

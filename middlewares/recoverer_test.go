package middlewares

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/require"
)

func Test_Recovery(t *testing.T) {
	t.Parallel()
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			require.Equal(t, "Hi, I'm an error!", err.Error())
			return c.SendStatus(fiber.StatusInternalServerError)
		},
	})

	app.Use(NewRecovery())

	app.Get("/panic", func(_ fiber.Ctx) error {
		panic("Hi, I'm an error!")
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/panic", nil))
	require.NoError(t, err)
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

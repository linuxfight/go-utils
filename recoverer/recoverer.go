package recoverer

import "github.com/gofiber/fiber/v3"

func Recovery() fiber.Handler {
	return func(c fiber.Ctx) error {
		defer func() {
			if err := recover(); err != nil {
				_ = c.Status(fiber.StatusInternalServerError).JSON(
					&fiber.Error{
						Message: "internal server error",
						Code:    fiber.StatusInternalServerError},
				)
			}
		}()

		return c.Next()
	}
}

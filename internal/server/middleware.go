package server

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/session"
)

func RequireAuth(c fiber.Ctx) error {
	sess := session.FromContext(c)
	if sess == nil {
		return c.Redirect().To("login")
	}

	if sess.Get("authenticated") != true {
		return c.Redirect().To("login")
	}

	return c.Next()
}

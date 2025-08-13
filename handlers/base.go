package handlers

import (
	views "github.com/AlphaByte02/SplitFlow/templates"
	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v3"
)

func Render(c fiber.Ctx, component templ.Component) error {
	c.Set("Content-Type", "text/html")
	return component.Render(c, c.Response().BodyWriter())
}

func HandleIndex(c fiber.Ctx) error {
	return Render(c, views.Layout("SplitFlow"))
}

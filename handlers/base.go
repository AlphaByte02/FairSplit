package handlers

import (
	"github.com/AlphaByte02/FairSplit/internal/db"
	views "github.com/AlphaByte02/FairSplit/web/templates"
	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v3"
)

func Render(c fiber.Ctx, component templ.Component) error {
	c.Set("Content-Type", "text/html")
	return component.Render(c, c.Response().BodyWriter())
}

func HandleIndex(c fiber.Ctx) error {
	user := fiber.Locals[db.User](c, "user")

	Q, _ := fiber.GetState[*db.Queries](c.App().State(), "queries")

	sessions, err := Q.ListSessionsForUser(c, user.ID)
	if err != nil {
		return err
	}

	return Render(c, views.Dashboard(sessions))
}

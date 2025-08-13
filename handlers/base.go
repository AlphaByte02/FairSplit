package handlers

import (
	"github.com/AlphaByte02/FairSplit/internal/db"
	views "github.com/AlphaByte02/FairSplit/web/templates"
	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/session"
)

func GetCurrentUser(c fiber.Ctx) (db.User, bool) {
	sess := session.FromContext(c)
	if sess == nil {
		return db.User{}, false
	}

	if sess.Get("authenticated") != true {
		return db.User{}, false
	}

	return sess.Get("user").(db.User), true
}

func Render(c fiber.Ctx, component templ.Component) error {
	c.Set("Content-Type", "text/html")
	return component.Render(c, c.Response().BodyWriter())
}

func HandleIndex(c fiber.Ctx) error {
	user, _ := GetCurrentUser(c)

	Q, _ := fiber.GetState[*db.Queries](c.App().State(), "queries")

	sessions, err := Q.ListSessionsForUser(c, user.ID)
	if err != nil {
		return err
	}

	c.Locals("user", user)
	return Render(c, views.Dashboard(sessions))
}

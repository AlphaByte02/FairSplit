package handlers

import (
	"github.com/AlphaByte02/FairSplit/internal/db"
	views "github.com/AlphaByte02/FairSplit/web/templates"
	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v3"
)

type ToastEvent struct {
	Level   string `json:"level"`
	Title   string `json:"title"`
	Message string `json:"message"`
}

func SendError(c fiber.Ctx, status int, level, title, message string) error {
	toast := ToastEvent{Level: level, Title: title, Message: message}
	return c.Status(status).JSON(toast)
}

func Render(c fiber.Ctx, component templ.Component) error {
	c.Set("Content-Type", "text/html")
	return component.Render(c, c.Response().BodyWriter())
}

func HandleIndex(c fiber.Ctx) error {
	user := fiber.Locals[db.User](c, "user")

	Q, _ := fiber.GetState[*db.Queries](c.App().State(), "queries")

	sessions, err := Q.ListSessionsForUser(c, user.ID)
	if err != nil {
		return SendError(c, fiber.StatusInternalServerError, "danger", "Errore", err.Error())
	}

	return Render(c, views.Dashboard(sessions))
}

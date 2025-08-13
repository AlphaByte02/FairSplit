package handlers

import (
	"github.com/AlphaByte02/FairSplit/internal/db"
	views "github.com/AlphaByte02/FairSplit/web/templates"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func HandleSession(c fiber.Ctx) error {
	user, _ := GetCurrentUser(c)

	name := c.FormValue("name")
	if name == "" {
		return fiber.ErrBadRequest
	}

	newSessionID, _ := uuid.NewV7()

	Q, _ := fiber.GetState[*db.Queries](c.App().State(), "queries")
	session, err := Q.CreateSession(c, db.CreateSessionParams{ID: newSessionID, CreatedByID: user.ID, Name: name})
	if err != nil {
		if err, ok := err.(*pgconn.PgError); ok && err.Code == "23505" {
			if c.Get("HX-Request") == "true" {
				c.Set(
					"HX-Trigger",
					`{"showToast": {"level" : "danger", "title" : "Errore", "message" : "Nome duplicato"}}`,
				)
				return c.SendStatus(fiber.StatusBadRequest)
			}
		} else {
			return err
		}
	}

	return Render(c, views.SessionItem(session))
}

package server

import (
	"github.com/AlphaByte02/FairSplit/internal/db"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/session"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func RequireAuth(c fiber.Ctx) error {
	sess := session.FromContext(c)
	if sess == nil {
		return c.Redirect().To("/login")
	}

	user, ok := sess.Get("user").(db.User)
	if !ok || user.Email == "" {
		sess.Reset()
		return c.Redirect().To("/login")
	}
	fiber.Locals(c, "user", user)

	return c.Next()
}

func HaveSessionAccess(c fiber.Ctx) error {
	user := fiber.Locals[db.User](c, "user")

	if err := uuid.Validate(c.Params("id")); err != nil {
		return fiber.ErrBadRequest
	}

	Q, _ := fiber.GetState[*db.Queries](c.App().State(), "queries")

	sessionID, _ := fiber.Convert(c.Params("id"), uuid.Parse)
	session, err := Q.GetSession(c, sessionID)
	if err != nil {
		if err == pgx.ErrNoRows {
			if c.Get("HX-Request") == "true" {
				c.Set(
					"HX-Trigger",
					`{"showToast": {"level" : "danger", "title" : "Errore", "message" : "Sessione non trovata"}}`,
				)
			}
			return c.SendStatus(fiber.StatusNotFound)
		} else {
			return err
		}
	}

	haveAccess, _ := Q.CheckSessionAccess(
		c,
		db.CheckSessionAccessParams{SessionID: session.Session.ID, UserID: user.ID},
	)
	if !haveAccess {
		if c.Get("HX-Request") == "true" {
			c.Set(
				"HX-Trigger",
				`{"showToast": {"level" : "danger", "title" : "Errore", "message" : "Non hai accesso a questa sessione"}}`,
			)
		}
		return c.SendStatus(fiber.StatusForbidden)
	}

	fiber.Locals(c, "session", session.Session)

	return c.Next()
}

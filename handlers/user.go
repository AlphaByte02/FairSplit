package handlers

import (
	"strings"

	"github.com/AlphaByte02/FairSplit/internal/db"
	views "github.com/AlphaByte02/FairSplit/web/templates"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/session"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func HandleLogin(c fiber.Ctx) error {
	sess := session.FromContext(c)

	username := c.FormValue("username")

	if username != "" {
		Q, _ := fiber.GetState[*db.Queries](c.App().State(), "queries")

		var user db.User

		user, err := Q.GetUserByUsername(c, strings.ToLower(username))
		if err != nil {
			if err == pgx.ErrNoRows {
				newUserID, _ := uuid.NewV7()
				user, err = Q.CreateUser(c, db.CreateUserParams{ID: newUserID, Username: username})
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}

		if err := sess.Session.Regenerate(); err != nil {
			return err
		}

		sess.Set("user", user)
		sess.Set("authenticated", true)

		if c.Get("HX-Request") == "true" {
			c.Set("HX-Redirect", "/")
			return c.SendStatus(fiber.StatusNoContent)
		}
		return c.Redirect().To("/")
	}

	return c.Status(401).SendString("Invalid credentials")
}

func HandleLogout(c fiber.Ctx) error {
	sess := session.FromContext(c)

	if err := sess.Reset(); err != nil {
		return c.Status(500).SendString("Session error")
	}

	return c.Redirect().To("/login")
}

func Login(c fiber.Ctx) error {
	return Render(c, views.Login())
}

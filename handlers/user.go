package handlers

import (
	"regexp"
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
		if !regexp.MustCompile(`^[a-zA-Z0-9_-]{3,20}$`).MatchString(username) {
			return SendError(c, fiber.StatusBadRequest, "danger", "Errore", "Username non valido")
		}

		Q, _ := fiber.GetState[*db.Queries](c.App().State(), "queries")

		var user db.User
		user, err := Q.GetUserByUsername(c, strings.ToLower(username))
		if err != nil {
			if err == pgx.ErrNoRows {
				newUserID, _ := uuid.NewV7()
				user, err = Q.CreateUser(c, db.CreateUserParams{ID: newUserID, Username: strings.ToLower(username)})
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

	return c.Status(fiber.StatusUnauthorized).SendString("Invalid credentials")
}

func HandleLogout(c fiber.Ctx) error {
	sess := session.FromContext(c)

	if err := sess.Reset(); err != nil {
		return SendError(c, fiber.StatusInternalServerError, "danger", "Errore", "Session error")
	}

	return c.Redirect().To("/login")
}

func Login(c fiber.Ctx) error {
	return Render(c, views.Login())
}

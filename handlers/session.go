package handlers

import (
	"github.com/AlphaByte02/FairSplit/internal/db"
	views "github.com/AlphaByte02/FairSplit/web/templates"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func HandleSession(c fiber.Ctx) error {
	user := fiber.Locals[db.User](c, "user")

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
			}
			return c.SendStatus(fiber.StatusBadRequest)
		}

		return err
	}

	err = Q.AddSessionParticipant(c, db.AddSessionParticipantParams{SessionID: session.ID, UserID: user.ID})
	if err != nil {
		return err
	}

	return Render(c, views.SessionItem(session))
}

func Session(c fiber.Ctx) error {
	session := fiber.Locals[db.Session](c, "session")

	Q, _ := fiber.GetState[*db.Queries](c.App().State(), "queries")

	participants, _ := Q.ListSessionParticipants(c, session.ID)
	transactions, _ := Q.ListTransactionsBySession(c, session.ID)

	return Render(c, views.SessionPage(session, participants, transactions))

}

func SessionInvite(c fiber.Ctx) error {
	session := fiber.Locals[db.Session](c, "session")

	Q, _ := fiber.GetState[*db.Queries](c.App().State(), "queries")

	participant, err := Q.GetUserByUsername(c, c.FormValue("username"))
	if err != nil {
		if err == pgx.ErrNoRows {
			if c.Get("HX-Request") == "true" {
				c.Set(
					"HX-Trigger",
					`{"showToast": {"level" : "danger", "title" : "Errore", "message" : "Utente non trovato"}}`,
				)
			}
			return c.SendStatus(fiber.StatusNotFound)
		}

		return err
	}

	err = Q.AddSessionParticipant(c, db.AddSessionParticipantParams{SessionID: session.ID, UserID: participant.ID})
	if err != nil {
		if c.Get("HX-Request") == "true" {
			c.Set(
				"HX-Trigger",
				`{"showToast": {"level" : "danger", "title" : "Errore", "message" : "Can not add"}}`,
			)
		}
		return c.SendStatus(fiber.StatusBadRequest)
	}

	participants, _ := Q.ListSessionParticipants(c, session.ID)

	return Render(c, views.ParticipantList(participants))
}

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

	onlyBody := fiber.Query(c, "onlyBody", false)

	if onlyBody {
		return Render(c, views.SessionBody(session, transactions))
	}

	return Render(c, views.SessionPage(session, participants, transactions))

}

func SessionInvite(c fiber.Ctx) error {
	session := fiber.Locals[db.Session](c, "session")

	Q, _ := fiber.GetState[*db.Queries](c.App().State(), "queries")

	username := c.FormValue("username")
	if username == "" {
		if c.Get("HX-Request") == "true" {
			c.Set(
				"HX-Trigger",
				`{"showToast": {"level" : "danger", "title" : "Errore", "message" : ""}}`,
			)
		}
		return c.SendStatus(fiber.StatusBadRequest)
	}

	participant, err := Q.GetUserByUsername(c, c.FormValue("username"))
	if err != nil {
		if err == pgx.ErrNoRows {
			newUserID, _ := uuid.NewV7()
			participant, _ = Q.CreateUser(c, db.CreateUserParams{ID: newUserID, Username: username})
		} else {
			return err
		}

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

	Render(c, views.PartecipantsModalList(session, participants))
	Render(c, views.PartecipantsCount(len(participants)))
	return Render(c, views.NewTransactionModalContent(session, participants, false))
}

func SessionKick(c fiber.Ctx) error {
	user := fiber.Locals[db.User](c, "user")
	session := fiber.Locals[db.Session](c, "session")

	if session.CreatedByID != user.ID {
		if c.Get("HX-Request") == "true" {
			c.Set(
				"HX-Trigger",
				`{"showToast": {"level" : "danger", "title" : "Errore", "message" : "Non hai i permessi per rimuovere utenti"}}`,
			)
		}
		return c.SendStatus(fiber.StatusForbidden)
	}

	Q, _ := fiber.GetState[*db.Queries](c.App().State(), "queries")

	toKickID, _ := fiber.Convert(c.Params("partecipant"), uuid.Parse)

	count, _ := Q.CountTransactionByUser(c, toKickID)
	if count > 0 {
		if c.Get("HX-Request") == "true" {
			c.Set(
				"HX-Trigger",
				`{"showToast": {"level" : "danger", "title" : "Errore", "message" : "Non puoi rimuovere questo utente"}}`,
			)
		}
		return c.SendStatus(fiber.StatusBadRequest)
	}

	err := Q.DeleteSessionParticipant(c, toKickID)
	if err != nil {
		if c.Get("HX-Request") == "true" {
			c.Set(
				"HX-Trigger",
				`{"showToast": {"level" : "danger", "title" : "Errore", "message" : "Non puoi rimuovere questo utente"}}`,
			)
		}
		return c.SendStatus(fiber.StatusBadRequest)
	}

	participants, _ := Q.ListSessionParticipants(c, session.ID)

	Render(c, views.PartecipantsCount(len(participants)))
	return Render(c, views.NewTransactionModalContent(session, participants, true))
}

func SessionClose(c fiber.Ctx) error {
	user := fiber.Locals[db.User](c, "user")
	session := fiber.Locals[db.Session](c, "session")

	if session.CreatedByID != user.ID {
		if c.Get("HX-Request") == "true" {
			c.Set(
				"HX-Trigger",
				`{"showToast": {"level" : "danger", "title" : "Errore", "message" : "Can not close"}}`,
			)
		}
		return c.SendStatus(fiber.StatusForbidden)
	}

	Q, _ := fiber.GetState[*db.Queries](c.App().State(), "queries")
	err := Q.CloseSession(c, session.ID)
	if err != nil {
		if c.Get("HX-Request") == "true" {
			c.Set(
				"HX-Trigger",
				`{"showToast": {"level" : "danger", "title" : "Errore", "message" : ""}}`,
			)
		}
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	session.IsClosed = true

	participants, _ := Q.ListSessionParticipants(c, session.ID)
	return Render(c, views.SessionHeader(session, participants))
}

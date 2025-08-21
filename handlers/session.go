package handlers

import (
	"regexp"
	"strings"

	"github.com/AlphaByte02/FairSplit/internal/db"
	views "github.com/AlphaByte02/FairSplit/web/templates"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var sessionRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-\s]{3,20}$`)

func HandleSession(c fiber.Ctx) error {
	user := fiber.Locals[db.User](c, "user")

	name := strings.Trim(c.FormValue("name"), " ")
	if !sessionRegex.MatchString(name) {
		return SendError(c, fiber.StatusBadRequest, "danger", "Errore", "Nome sessione non valido")
	}

	newSessionID, _ := uuid.NewV7()

	Q, _ := fiber.GetState[*db.Queries](c.App().State(), "queries")
	session, err := Q.CreateSession(c, db.CreateSessionParams{ID: newSessionID, CreatedByID: user.ID, Name: name})
	if err != nil {
		if err, ok := err.(*pgconn.PgError); ok && err.Code == "23505" {
			return SendError(c, fiber.StatusBadRequest, "danger", "Errore", "Questa sessione esiste già")
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

	username := strings.ToLower(c.FormValue("username"))
	if username == "" {
		return SendError(c, fiber.StatusBadRequest, "danger", "Errore", "L'username non può essere vuoto")
	}

	participant, err := Q.GetUserByEmailOrUsername(c, username)
	if err != nil {
		if err == pgx.ErrNoRows {
			return SendError(c, fiber.StatusBadRequest, "danger", "Errore", "Questo utente non esiste")
		}

		return err
	}

	err = Q.AddSessionParticipant(c, db.AddSessionParticipantParams{SessionID: session.ID, UserID: participant.ID})
	if err != nil {
		return SendError(c, fiber.StatusInternalServerError, "danger", "Errore Server", "Errore in aggiunta")
	}

	participants, _ := Q.ListSessionParticipants(c, session.ID)

	Render(c, views.PartecipantsModalList(session, participants))
	Render(c, views.PartecipantsCount(len(participants)))
	return Render(
		c,
		views.TransactionModalContent(views.TransactionModalProps{Session: session, AllParticipants: participants}),
	)
}

func SessionKick(c fiber.Ctx) error {
	user := fiber.Locals[db.User](c, "user")
	session := fiber.Locals[db.Session](c, "session")

	if session.CreatedByID != user.ID {
		return SendError(c, fiber.StatusForbidden, "danger", "Errore", "Non hai i permessi per rimuovere utenti")
	}

	Q, _ := fiber.GetState[*db.Queries](c.App().State(), "queries")

	toKickID, _ := fiber.Convert(c.Params("partecipant"), uuid.Parse)

	count, _ := Q.CountTransactionByUser(c, toKickID)
	if count > 0 {
		return SendError(c, fiber.StatusBadRequest, "danger", "Errore", "Non puoi rimuovere questo utente")
	}

	err := Q.DeleteSessionParticipant(c, toKickID)
	if err != nil {
		return SendError(c, fiber.StatusInternalServerError, "danger", "Errore Server", "Errore in rimozione")
	}

	participants, _ := Q.ListSessionParticipants(c, session.ID)

	Render(c, views.PartecipantsCount(len(participants)))
	return Render(
		c,
		views.TransactionModalContent(views.TransactionModalProps{Session: session, AllParticipants: participants}),
	)
}

func SessionClose(c fiber.Ctx) error {
	user := fiber.Locals[db.User](c, "user")
	session := fiber.Locals[db.Session](c, "session")

	if session.CreatedByID != user.ID {
		return SendError(
			c,
			fiber.StatusForbidden,
			"danger",
			"Errore",
			"Non hai i permessi per chiudere questa sessione",
		)
	}

	Q, _ := fiber.GetState[*db.Queries](c.App().State(), "queries")
	err := Q.CloseSession(c, session.ID)
	if err != nil {
		return SendError(
			c,
			fiber.StatusInternalServerError,
			"danger",
			"Errore Server",
			"Impossibile chiudere la sessione",
		)
	}
	session.IsClosed = true

	participants, _ := Q.ListSessionParticipants(c, session.ID)
	return Render(c, views.SessionHeader(session, participants))
}

func SessionRename(c fiber.Ctx) error {
	user := fiber.Locals[db.User](c, "user")
	session := fiber.Locals[db.Session](c, "session")

	if session.CreatedByID != user.ID {
		return SendError(
			c,
			fiber.StatusForbidden,
			"danger",
			"Errore",
			"Non hai i permessi per rinominare questa sessione",
		)
	}

	newName := strings.Trim(c.FormValue("name"), " ")
	if !sessionRegex.MatchString(newName) {
		return SendError(c, fiber.StatusBadRequest, "danger", "Errore", "Nome sessione non valido")
	}

	Q, _ := fiber.GetState[*db.Queries](c.App().State(), "queries")
	err := Q.RenameSession(c, db.RenameSessionParams{ID: session.ID, Name: newName})
	if err != nil {
		return SendError(
			c,
			fiber.StatusInternalServerError,
			"danger",
			"Errore Server",
			"Impossibile chiudere la sessione",
		)
	}
	session.Name = newName

	participants, _ := Q.ListSessionParticipants(c, session.ID)
	return Render(c, views.SessionHeader(session, participants))
}

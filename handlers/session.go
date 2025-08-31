package handlers

import (
	"regexp"
	"slices"
	"strings"

	"github.com/AlphaByte02/FairSplit/internal/db"
	views "github.com/AlphaByte02/FairSplit/web/templates"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
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
	pool, _ := fiber.GetState[*pgxpool.Pool](c.App().State(), "pool")

	tx, _ := pool.Begin(c)
	defer tx.Rollback(c)
	qtx := Q.WithTx(tx)

	session, err := qtx.CreateSession(c, db.CreateSessionParams{ID: newSessionID, CreatedByID: user.ID, Name: name})
	if err != nil {
		if err, ok := err.(*pgconn.PgError); ok && err.Code == "23505" {
			return SendError(c, fiber.StatusBadRequest, "danger", "Errore", "Questa sessione esiste già")
		}

		return err
	}

	err = qtx.AddSessionParticipant(c, db.AddSessionParticipantParams{SessionID: session.ID, UserID: user.ID})
	if err != nil {
		return err
	}

	err = tx.Commit(c)
	if err != nil {
		return SendError(c, fiber.StatusBadRequest, "danger", "Errore", "Qualcosa è andato storto")
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

	count, _ := Q.CountTransactionByUserAndSession(
		c,
		db.CountTransactionByUserAndSessionParams{UserID: toKickID, SessionID: session.ID},
	)
	if count > 0 {
		return SendError(c, fiber.StatusBadRequest, "danger", "Errore", "Non puoi rimuovere questo utente")
	}

	err := Q.DeleteSessionParticipant(c, db.DeleteSessionParticipantParams{SessionID: session.ID, UserID: toKickID})
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

	balances, err := Q.GetSessionBalances(c, session.ID)
	if err != nil {
		return err
	}

	minValue := decimal.NewFromFloat(0.01)

	type balanceItem struct {
		User   db.User
		Amount decimal.Decimal
	}

	var debtors, creditors []balanceItem
	for _, b := range balances {
		if b.Balance.GreaterThan(minValue) {
			creditors = append(creditors, balanceItem{
				User:   b.User,
				Amount: b.Balance,
			})
		} else if b.Balance.LessThan(minValue.Neg()) {
			debtors = append(debtors, balanceItem{
				User:   b.User,
				Amount: b.Balance.Neg(),
			})
		}
	}

	slices.SortFunc(debtors, func(a, b balanceItem) int {
		return b.Amount.Cmp(a.Amount)
	})
	slices.SortFunc(creditors, func(a, b balanceItem) int {
		return b.Amount.Cmp(a.Amount)
	})

	var minimizedBalances []db.SaveFinalBalanceParams
	di, ci := 0, 0
	for di < len(debtors) && ci < len(creditors) {
		payAmt := decimal.Min(debtors[di].Amount, creditors[ci].Amount)
		if payAmt.GreaterThan(minValue) {
			newID, _ := uuid.NewV7()
			minimizedBalances = append(minimizedBalances, db.SaveFinalBalanceParams{
				ID:         newID,
				SessionID:  session.ID,
				CreditorID: debtors[di].User.ID,
				DebtorID:   creditors[ci].User.ID,
				Amount:     payAmt,
			})
		}
		debtors[di].Amount = debtors[di].Amount.Sub(payAmt)
		creditors[ci].Amount = creditors[ci].Amount.Sub(payAmt)
		if debtors[di].Amount.LessThan(minValue) {
			di++
		}
		if creditors[ci].Amount.LessThan(minValue) {
			ci++
		}
	}

	Q.SaveFinalBalance(c, minimizedBalances)

	err = Q.CloseSession(c, session.ID)
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

	transfers, _ := Q.GetFinalBalancesBySession(c, session.ID)

	participants, _ := Q.ListSessionParticipants(c, session.ID)

	Render(c, views.SessionHeader(session, participants))
	return Render(c, views.FinalBalance(session, transfers))
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

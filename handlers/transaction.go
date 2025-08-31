package handlers

import (
	"github.com/AlphaByte02/FairSplit/internal/db"
	"github.com/AlphaByte02/FairSplit/internal/types"
	views "github.com/AlphaByte02/FairSplit/web/templates"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

func DeleteTransaction(c fiber.Ctx) error {
	user := fiber.Locals[db.User](c, "user")
	Q, _ := fiber.GetState[*db.Queries](c.App().State(), "queries")

	toDeleteID, _ := fiber.Convert(c.Params("transaction"), uuid.Parse)

	transaction, err := Q.GetTransaction(c, toDeleteID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return SendError(c, fiber.StatusNotFound, "danger", "Errore", "Transazione non trovata")
		}

		return SendError(c, fiber.StatusInternalServerError, "danger", "Errore", "C'è stato un errore")
	}

	if transaction.CreatedByID != user.ID {
		return SendError(c, fiber.StatusForbidden, "danger", "Errore", "Non hai i permessi per questa azione")
	}

	err = Q.DeleteTransaction(c, transaction.ID)
	if err != nil {
		return SendError(c, fiber.StatusInternalServerError, "danger", "Errore Server", "Errore in eliminazione")
	}

	session := fiber.Locals[db.Session](c, "session")
	transactions, _ := Q.ListTransactionsBySession(c, session.ID)

	return Render(c, views.SessionBody(session, transactions))
}

func Transaction(c fiber.Ctx) error {
	session := fiber.Locals[db.Session](c, "session")
	Q, _ := fiber.GetState[*db.Queries](c.App().State(), "queries")

	allParticipants, _ := Q.ListSessionParticipants(c, session.ID)

	return Render(
		c,
		views.TransactionModalContent(
			views.TransactionModalProps{
				Session:         session,
				AllParticipants: allParticipants,
				IsEdit:          false,
			},
		),
	)
}

func EditTransaction(c fiber.Ctx) error {
	user := fiber.Locals[db.User](c, "user")
	session := fiber.Locals[db.Session](c, "session")
	Q, _ := fiber.GetState[*db.Queries](c.App().State(), "queries")

	transactionID, _ := fiber.Convert(c.Params("transaction"), uuid.Parse)

	transaction, err := Q.GetTransaction(c, transactionID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return SendError(c, fiber.StatusNotFound, "danger", "Errore", "Transazione non trovata")
		}

		return SendError(c, fiber.StatusInternalServerError, "danger", "Errore", "C'è stato un errore")
	}

	if transaction.CreatedByID != user.ID {
		return SendError(c, fiber.StatusForbidden, "danger", "Errore", "Non hai i permessi per questa azione")
	}

	allParticipants, _ := Q.ListSessionParticipants(c, session.ID)
	participants, _ := Q.ListTransactionParticipants(c, transaction.ID)

	return Render(
		c,
		views.TransactionModalContent(
			views.TransactionModalProps{
				Session:         session,
				AllParticipants: allParticipants,
				Participants:    participants,
				Transaction:     transaction,
				IsEdit:          true,
			},
		),
	)
}

func HandleEditTransaction(c fiber.Ctx) error {
	user := fiber.Locals[db.User](c, "user")
	session := fiber.Locals[db.Session](c, "session")
	Q, _ := fiber.GetState[*db.Queries](c.App().State(), "queries")
	pool, _ := fiber.GetState[*pgxpool.Pool](c.App().State(), "pool")

	transactionID, _ := fiber.Convert(c.Params("transaction"), uuid.Parse)

	transaction, err := Q.GetTransaction(c, transactionID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return SendError(c, fiber.StatusNotFound, "danger", "Errore", "Transazione non trovata")
		}

		return SendError(c, fiber.StatusInternalServerError, "danger", "Errore", "C'è stato un errore")
	}

	if transaction.CreatedByID != user.ID {
		return SendError(c, fiber.StatusForbidden, "danger", "Errore", "Non hai i permessi per questa azione")
	}

	var transactionParams struct {
		Payer       uuid.UUID       `json:"payer" form:"payer"`
		Amount      decimal.Decimal `json:"amount" form:"amount"`
		Description types.Text      `json:"description" form:"description"`
		PaidFor     []string        `json:"paid_for" form:"paid_for"`
	}
	err = c.Bind().Form(&transactionParams)
	if err != nil {
		return SendError(c, fiber.StatusBadRequest, "danger", "Errore", "Alcuni campi sono formattati male")
	}

	if len(transactionParams.PaidFor) == 0 {
		return SendError(c, fiber.StatusBadRequest, "danger", "Errore", "Nessun partecipante aggiunto")
	}

	tx, _ := pool.Begin(c)
	defer tx.Rollback(c)
	qtx := Q.WithTx(tx)

	err = qtx.UpdateTransactions(
		c,
		db.UpdateTransactionsParams{
			ID:          transaction.ID,
			SessionID:   session.ID,
			PayerID:     transactionParams.Payer,
			Amount:      transactionParams.Amount,
			Description: transactionParams.Description,
		},
	)

	// TODO: Manage difference on PaidFor

	err = tx.Commit(c)
	if err != nil {
		return SendError(c, fiber.StatusBadRequest, "danger", "Errore", "Qualcosa è andato storto")
	}

	transactions, _ := Q.ListTransactionsBySession(c, session.ID)

	return Render(c, views.SessionBody(session, transactions))
}

func HandleTransaction(c fiber.Ctx) error {
	user := fiber.Locals[db.User](c, "user")
	session := fiber.Locals[db.Session](c, "session")
	Q, _ := fiber.GetState[*db.Queries](c.App().State(), "queries")
	pool, _ := fiber.GetState[*pgxpool.Pool](c.App().State(), "pool")

	var transactionParams struct {
		Payer       uuid.UUID       `json:"payer" form:"payer"`
		Amount      decimal.Decimal `json:"amount" form:"amount"`
		Description types.Text      `json:"description" form:"description"`
		PaidFor     []string        `json:"paid_for" form:"paid_for"`
	}
	err := c.Bind().Form(&transactionParams)
	if err != nil {
		return SendError(c, fiber.StatusBadRequest, "danger", "Errore", "Alcuni campi sono formattati male")
	}

	if len(transactionParams.PaidFor) == 0 {
		return SendError(c, fiber.StatusBadRequest, "danger", "Errore", "Nessun partecipante aggiunto")
	}

	tx, _ := pool.Begin(c)
	defer tx.Rollback(c)
	qtx := Q.WithTx(tx)

	newTransactionID, _ := uuid.NewV7()
	_, err = qtx.CreateTransaction(c, db.CreateTransactionParams{
		ID:          newTransactionID,
		SessionID:   session.ID,
		PayerID:     transactionParams.Payer,
		Amount:      transactionParams.Amount,
		Description: transactionParams.Description,
		CreatedByID: user.ID,
	})
	if err != nil {
		return SendError(c, fiber.StatusInternalServerError, "danger", "Errore Server", "Errore in creazione")
	}

	for _, partecipantUUID := range transactionParams.PaidFor {
		pUUID, err := uuid.Parse(partecipantUUID)
		if err != nil {
			return SendError(c, fiber.StatusBadRequest, "danger", "Errore", "L'id di un partecipante non è valido")
		}

		err = qtx.AddTransactionParticipant(
			c,
			db.AddTransactionParticipantParams{TransactionID: newTransactionID, UserID: pUUID},
		)
		if err != nil {
			return SendError(c, fiber.StatusBadRequest, "danger", "Errore", "L'id di un partecipante non è valido")
		}
	}

	err = tx.Commit(c)
	if err != nil {
		return SendError(c, fiber.StatusBadRequest, "danger", "Errore", "Qualcosa è andato storto")
	}

	transactions, _ := Q.ListTransactionsBySession(c, session.ID)

	return Render(c, views.SessionBody(session, transactions))
}

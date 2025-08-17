package handlers

import (
	"github.com/AlphaByte02/FairSplit/internal/db"
	"github.com/AlphaByte02/FairSplit/internal/types"
	views "github.com/AlphaByte02/FairSplit/web/templates"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

func DeleteTransaction(c fiber.Ctx) error {
	Q, _ := fiber.GetState[*db.Queries](c.App().State(), "queries")

	toDeleteID, _ := fiber.Convert(c.Params("transaction"), uuid.Parse)

	err := Q.DeleteTransaction(c, toDeleteID)
	if err != nil {
		return SendError(c, fiber.StatusInternalServerError, "danger", "Errore Server", "Errore in eliminazione")
	}

	// session := fiber.Locals[db.Session](c, "session")
	// transactions, _ := Q.ListTransactionsBySession(c, session.ID)
	// return Render(c, views.TransactionList(transactions))

	return c.SendStatus(fiber.StatusNoContent)
}

func HandleTransaction(c fiber.Ctx) error {
	session := fiber.Locals[db.Session](c, "session")
	Q, _ := fiber.GetState[*db.Queries](c.App().State(), "queries")

	var transactionParams struct {
		Payer       uuid.UUID     `json:"payer" form:"payer"`
		Amount      types.Numeric `json:"amount" form:"amount"`
		Description types.Text    `json:"description" form:"description"`
		PaidFor     []string      `json:"paid_for" form:"paid_for"`
	}
	err := c.Bind().Form(&transactionParams)
	if err != nil {
		return SendError(c, fiber.StatusBadRequest, "danger", "Errore", "Alcuni campi sono formattati male")
	}

	if len(transactionParams.PaidFor) == 0 {
		return SendError(c, fiber.StatusBadRequest, "danger", "Errore", "Nessun partecipante aggiunto")
	}

	newTransactionID, _ := uuid.NewV7()
	_, err = Q.CreateTransaction(c, db.CreateTransactionParams{
		ID:          newTransactionID,
		SessionID:   session.ID,
		PayerID:     transactionParams.Payer,
		Amount:      transactionParams.Amount,
		Description: transactionParams.Description,
	})
	if err != nil {
		return SendError(c, fiber.StatusInternalServerError, "danger", "Errore Server", "Errore in creazione")
	}

	for _, partecipantUUID := range transactionParams.PaidFor {
		pUUID, err := uuid.Parse(partecipantUUID)
		if err != nil {
			return SendError(c, fiber.StatusBadRequest, "danger", "Errore", "L'id di un partecipante non è valido")
		}

		err = Q.AddTransactionParticipant(
			c,
			db.AddTransactionParticipantParams{TransactionID: newTransactionID, UserID: pUUID},
		)
		if err != nil {
			return SendError(c, fiber.StatusBadRequest, "danger", "Errore", "L'id di un partecipante non è valido")
		}
	}

	transactions, _ := Q.ListTransactionsBySession(c, session.ID)

	return Render(c, views.SessionBody(session, transactions))
}

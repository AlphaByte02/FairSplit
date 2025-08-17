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
		if c.Get("HX-Request") == "true" {
			c.Set(
				"HX-Trigger",
				`{"showToast": {"level" : "danger", "title" : "Errore", "message" : "Can not remove"}}`,
			)
		}
		return c.SendStatus(fiber.StatusBadRequest)
	}

	// session := fiber.Locals[db.Session](c, "session")
	// transactions, _ := Q.ListTransactionsBySession(c, session.ID)
	// return Render(c, views.TransactionList(transactions))

	return c.SendStatus(fiber.StatusNoContent)
}

func HandleTransaction(c fiber.Ctx) error {
	user := fiber.Locals[db.User](c, "user")
	session := fiber.Locals[db.Session](c, "session")
	Q, _ := fiber.GetState[*db.Queries](c.App().State(), "queries")

	var transactionParams struct {
		Amount      types.Numeric `json:"amount" form:"amount"`
		Description types.Text    `json:"description" form:"description"`
		PaidFor     []string      `json:"paid_for" form:"paid_for"`
	}
	err := c.Bind().Form(&transactionParams)
	if err != nil {
		if c.Get("HX-Request") == "true" {
			c.Set(
				"HX-Trigger",
				`{"showToast": {"level" : "danger", "title" : "Errore", "message" : ""}}`,
			)
		}

		return c.SendStatus(fiber.StatusBadRequest)
	}

	newTransactionID, _ := uuid.NewV7()
	_, err = Q.CreateTransaction(c, db.CreateTransactionParams{
		ID:          newTransactionID,
		SessionID:   session.ID,
		PayerID:     user.ID,
		Amount:      transactionParams.Amount,
		Description: transactionParams.Description,
	})
	if err != nil {
		if c.Get("HX-Request") == "true" {
			c.Set(
				"HX-Trigger",
				`{"showToast": {"level" : "danger", "title" : "Errore", "message" : ""}}`,
			)
		}

		return c.SendStatus(fiber.StatusInternalServerError)
	}

	for _, partecipantUUID := range transactionParams.PaidFor {
		pUUID, err := uuid.Parse(partecipantUUID)
		if err != nil {
			return err
		}

		err = Q.AddTransactionParticipant(
			c,
			db.AddTransactionParticipantParams{TransactionID: newTransactionID, UserID: pUUID},
		)
		if err != nil {
			return err
		}
	}

	transactions, _ := Q.ListTransactionsBySession(c, session.ID)

	return Render(c, views.TransactionList(transactions))
}

package handlers

import (
	"maps"
	"slices"

	"github.com/AlphaByte02/FairSplit/internal/db"
	"github.com/AlphaByte02/FairSplit/internal/types"
	views "github.com/AlphaByte02/FairSplit/web/templates"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

func BalancesIntermediate(c fiber.Ctx) error {
	session := fiber.Locals[db.Session](c, "session")

	Q, _ := fiber.GetState[*db.Queries](c.App().State(), "queries")

	rows, err := Q.GetIntermediateBalances(c, session.ID)
	if err != nil {
		return err
	}

	m := make(map[uuid.UUID]views.IntermediateBalanceParticipant)
	for _, rows := range rows {
		if rows.User.ID == rows.User_2.ID {
			continue
		}

		if p, ok := m[rows.User_2.ID]; ok {
			p.Transactions = append(p.Transactions, struct {
				Transaction db.Transaction
				Payer       db.User
				Amount      types.Numeric
			}{Transaction: rows.Transaction, Payer: rows.User, Amount: rows.AmountPerUser})

			p.Sum.Add(rows.AmountPerUser)
		} else {
			p := views.IntermediateBalanceParticipant{
				User: rows.User_2,
				Sum:  rows.AmountPerUser,
				Transactions: make([]struct {
					Transaction db.Transaction
					Payer       db.User
					Amount      types.Numeric
				}, 0),
			}

			p.Transactions = append(p.Transactions, struct {
				Transaction db.Transaction
				Payer       db.User
				Amount      types.Numeric
			}{Transaction: rows.Transaction, Payer: rows.User, Amount: rows.AmountPerUser})

			m[rows.User_2.ID] = p
		}

	}

	participants := slices.Collect(maps.Values(m))

	return Render(c, views.IntermediateBalance(session, participants))
}

func BalancesFinal(c fiber.Ctx) error {
	session := fiber.Locals[db.Session](c, "session")

	return Render(c, views.FinalBalance(session))
}

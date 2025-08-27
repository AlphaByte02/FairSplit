package handlers

import (
	"maps"
	"slices"
	"strings"

	"github.com/AlphaByte02/FairSplit/internal/db"
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

	balances := make(map[uuid.UUID]views.IntermediateBalanceParticipant)
	for _, row := range rows {
		if row.User.ID == row.User_2.ID {
			continue
		}

		balance, exists := balances[row.User_2.ID]
		if !exists {
			balance = views.IntermediateBalanceParticipant{
				Debtor:       row.User_2,
				Transactions: make([]views.IntermediateBalanceTransaction, 0),
				Sum:          row.AmountPerUser,
			}
		} else {
			balance.Sum = balance.Sum.Add(row.AmountPerUser)
		}

		balance.Transactions = append(
			balance.Transactions,
			views.IntermediateBalanceTransaction{
				Transaction: row.Transaction,
				Payer:       row.User,
				Amount:      row.AmountPerUser,
			},
		)

		balances[row.User_2.ID] = balance
	}

	participants := slices.Collect(maps.Values(balances))

	slices.SortFunc(participants, func(a, b views.IntermediateBalanceParticipant) int {
		return strings.Compare(a.Debtor.Username.String, b.Debtor.Username.String)
	})

	return Render(c, views.IntermediateBalance(session, participants))
}

func BalancesFinal(c fiber.Ctx) error {
	session := fiber.Locals[db.Session](c, "session")

	var transfers []db.GetFinalBalancesBySessionRow

	return Render(c, views.FinalBalance(session, transfers))
}

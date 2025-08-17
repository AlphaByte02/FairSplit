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
			balance.Sum, _ = balance.Sum.Add(row.AmountPerUser)
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
		return strings.Compare(a.Debtor.Username, b.Debtor.Username)
	})

	return Render(c, views.IntermediateBalance(session, participants))
}

type balanceItem struct {
	User   db.User
	Amount float64
}

func BalancesFinal(c fiber.Ctx) error {
	session := fiber.Locals[db.Session](c, "session")
	Q, _ := fiber.GetState[*db.Queries](c.App().State(), "queries")

	balances, err := Q.GetSessionBalances(c, session.ID)
	if err != nil {
		return err
	}

	var debtors, creditors []balanceItem
	for _, b := range balances {
		bal, _ := b.Balance.Float64()
		if bal > 0.01 {
			creditors = append(creditors, balanceItem{
				User:   b.User,
				Amount: bal,
			})
		} else if bal < -0.01 {
			debtors = append(debtors, balanceItem{
				User:   b.User,
				Amount: -bal,
			})
		}
	}

	slices.SortFunc(debtors, func(a, b balanceItem) int {
		return int(b.Amount - a.Amount)
	})
	slices.SortFunc(creditors, func(a, b balanceItem) int {
		return int(b.Amount - a.Amount)
	})

	var transfers []views.BalanceTransferItem
	di, ci := 0, 0
	for di < len(debtors) && ci < len(creditors) {
		payAmt := min(debtors[di].Amount, creditors[ci].Amount)
		if payAmt > 0.01 {
			transfers = append(transfers, views.BalanceTransferItem{
				From:   debtors[di].User,
				To:     creditors[ci].User,
				Amount: payAmt,
			})
		}
		debtors[di].Amount -= payAmt
		creditors[ci].Amount -= payAmt
		if debtors[di].Amount < 0.01 {
			di++
		}
		if creditors[ci].Amount < 0.01 {
			ci++
		}
	}

	return Render(c, views.FinalBalance(session, transfers))
}

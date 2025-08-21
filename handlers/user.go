package handlers

import (
	"regexp"

	"github.com/AlphaByte02/FairSplit/internal/db"
	"github.com/AlphaByte02/FairSplit/internal/types"
	views "github.com/AlphaByte02/FairSplit/web/templates"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/session"
)

func User(c fiber.Ctx) error {
	return Render(c, views.User())
}

func HandleUpdateUser(c fiber.Ctx) error {
	user := fiber.Locals[db.User](c, "user")

	var userInfo struct {
		Username       types.Text            `json:"username" form:"username"`
		PaypalUsername types.Text            `json:"paypal_username" form:"paypal_username"`
		Iban           types.EncryptedString `json:"iban" form:"iban"`
	}
	err := c.Bind().Form(&userInfo)
	if err != nil {
		return SendError(c, fiber.StatusBadRequest, "danger", "Errore", "Alcuni campi sono formattati male")
	}

	newUsername := userInfo.Username.String
	if !regexp.MustCompile(`^[a-zA-Z0-9_-]{3,20}$`).MatchString(newUsername) {
		return SendError(c, fiber.StatusBadRequest, "danger", "Errore", "Username non valido")
	}

	Q, _ := fiber.GetState[*db.Queries](c.App().State(), "queries")

	exists, err := Q.CheckUserExists(c, newUsername)
	if (err != nil || exists) && newUsername != user.Username.String {
		return SendError(c, fiber.StatusBadRequest, "danger", "Errore", "Questo username esiste già")
	}

	err = Q.UpdateUser(
		c,
		db.UpdateUserParams{
			ID:             user.ID,
			Username:       types.NewText(newUsername),
			PaypalUsername: userInfo.PaypalUsername,
			Iban:           userInfo.Iban,
		},
	)
	if err != nil {
		return SendError(c, fiber.StatusInternalServerError, "danger", "Errore", "Impossibile aggiornare l'utente")
	}

	sess := session.FromContext(c)

	user.Username = userInfo.Username
	user.PaypalUsername = userInfo.PaypalUsername
	user.Iban = userInfo.Iban

	sess.Set("user", user)

	Render(c, views.UserNav(user))
	return Render(c, views.UserInfo(user))
}

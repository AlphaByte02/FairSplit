package handlers

import (
	"context"
	"encoding/json"

	"github.com/AlphaByte02/FairSplit/internal/db"
	"github.com/AlphaByte02/FairSplit/internal/service/google"
	"github.com/AlphaByte02/FairSplit/internal/types"
	views "github.com/AlphaByte02/FairSplit/web/templates"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/session"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/oauth2"
)

func HandleGoogleLogin(c fiber.Ctx) error {
	sess := session.FromContext(c)
	gconfig := google.GetConfig()

	state := uuid.New().String()
	sess.Set("oauth_state", state)
	url := gconfig.AuthCodeURL(state, oauth2.AccessTypeOffline)

	if c.Get("HX-Request") == "true" {
		c.Set("HX-Redirect", url)
		return c.SendStatus(fiber.StatusNoContent)
	}
	return c.Redirect().To(url)
}

func HandleGoogleLoginCallback(c fiber.Ctx) error {
	sess := session.FromContext(c)

	storedState := sess.Get("oauth_state")
	if storedState == nil || c.Query("state") != storedState.(string) {
		return c.Status(fiber.StatusBadRequest).SendString("State non valido (CSRF)")
	}
	sess.Delete("oauth_state")
	code := c.Query("code")
	if code == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Code mancante")
	}

	gconfig := google.GetConfig()

	// Scambia code per token
	token, err := gconfig.Exchange(context.Background(), code)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).SendString("Errore scambio token: " + err.Error())
	}

	client := gconfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Errore fetch userinfo: " + err.Error())
	}
	defer resp.Body.Close()

	var userInfo struct {
		Email      string `json:"email"`
		Name       string `json:"name"`
		GivenName  string `json:"given_name"`
		FamilyName string `json:"family_name"`
		Picture    string `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Errore decodifica userinfo: " + err.Error())
	}

	if userInfo.Email == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Email non disponibile")
	}

	redirectUrl := "/"

	Q, _ := fiber.GetState[*db.Queries](c.App().State(), "queries")

	var user db.User
	user, err = Q.GetUserByEmail(c, userInfo.Email)
	if err != nil {
		if err == pgx.ErrNoRows {
			newUserID, _ := uuid.NewV7()
			user, err = Q.CreateUser(
				c,
				db.CreateUserParams{ID: newUserID, Email: userInfo.Email, Picture: types.NewText(userInfo.Picture)},
			)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).SendString("Errore creazione user")
			}

			redirectUrl = "/user"
		} else {
			return c.Status(fiber.StatusInternalServerError).SendString("Errore creazione user")
		}
	} else {
		if user.Picture.String != userInfo.Picture {
			Q.UpdateUserPicture(c, db.UpdateUserPictureParams{ID: user.ID, Picture: types.NewText(userInfo.Picture)})
			user.Picture = types.NewText(userInfo.Picture)
		}
	}

	if err := sess.Session.Regenerate(); err != nil {
		return err
	}

	sess.Set("user", user)

	if c.Get("HX-Request") == "true" {
		c.Set("HX-Redirect", redirectUrl)
		return c.SendStatus(fiber.StatusNoContent)
	}
	return c.Redirect().To(redirectUrl)
}

func HandleLogout(c fiber.Ctx) error {
	sess := session.FromContext(c)

	if err := sess.Reset(); err != nil {
		return SendError(c, fiber.StatusInternalServerError, "danger", "Errore", "Session error")
	}

	return c.Redirect().To("/login")
}

func Login(c fiber.Ctx) error {
	return Render(c, views.Login())
}

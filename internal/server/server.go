package server

import (
	"fmt"
	"strings"

	"github.com/AlphaByte02/FairSplit/handlers"
	"github.com/gofiber/fiber/v3"
)

type Server struct {
	*fiber.App
}

func New(app *fiber.App) *Server {
	return &Server{app}
}

func (s *Server) RegisterRoutes() {
	s.Get("/", RequireAuth, handlers.HandleIndex)

	s.Get("/login", handlers.Login)
	s.Get("/auth/google", handlers.HandleGoogleLogin)
	s.Get("/auth/google/callback", handlers.HandleGoogleLoginCallback)
	s.Get("/logout", RequireAuth, handlers.HandleLogout)

	s.Get("/user", RequireAuth, handlers.User)
	s.Patch("/user", RequireAuth, handlers.HandleUpdateUser)

	s.Post("/sessions/new", RequireAuth, handlers.HandleSession)
	s.Get("/sessions/:id", RequireAuth, HaveSessionAccess, handlers.Session)
	s.Post("/sessions/:id/rename", RequireAuth, HaveSessionAccess, handlers.SessionRename)
	s.Post("/sessions/:id/close", RequireAuth, HaveSessionAccess, handlers.SessionClose)
	s.Post("/sessions/:id/invite", RequireAuth, HaveSessionAccess, handlers.SessionInvite)
	s.Delete("/sessions/:id/kick/:partecipant", RequireAuth, HaveSessionAccess, handlers.SessionKick)

	s.Get("/sessions/:id/transactions", RequireAuth, HaveSessionAccess, handlers.Transaction)
	s.Post("/sessions/:id/transactions", RequireAuth, HaveSessionAccess, handlers.HandleTransaction)
	s.Get("/sessions/:id/transactions/:transaction", RequireAuth, HaveSessionAccess, handlers.EditTransaction)
	s.Patch("/sessions/:id/transactions/:transaction", RequireAuth, HaveSessionAccess, handlers.HandleEditTransaction)
	s.Delete("/sessions/:id/transactions/:transaction", RequireAuth, HaveSessionAccess, handlers.DeleteTransaction)

	s.Get("/sessions/:id/balances/intermediate", RequireAuth, HaveSessionAccess, handlers.BalancesIntermediate)
	s.Get("/sessions/:id/balances/final", RequireAuth, HaveSessionAccess, handlers.BalancesFinal)
	s.Post("/sessions/:id/balances/:balance/toggle-paid", RequireAuth, HaveSessionAccess, handlers.BalanceTogglePaid)
}

func (s *Server) Start(port string) {
	err := s.Listen(fmt.Sprintf(":%s", strings.Trim(port, " :")))
	if err != nil {
		panic(fmt.Sprintf("cannot start server: %s", err))
	}
}

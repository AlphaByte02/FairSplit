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
	s.Post("/login", handlers.HandleLogin)
	s.Get("/logout", RequireAuth, handlers.HandleLogout)

	s.Post("/sessions/new", RequireAuth, handlers.HandleSession)
}

func (s *Server) Start(port string) {
	err := s.Listen(fmt.Sprintf(":%s", strings.Trim(port, " :")))
	if err != nil {
		panic(fmt.Sprintf("cannot start server: %s", err))
	}
}

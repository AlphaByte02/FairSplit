package main

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	db "github.com/AlphaByte02/FairSplit/internal/db"
	"github.com/AlphaByte02/FairSplit/internal/server"

	_ "github.com/joho/godotenv/autoload"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/favicon"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/session"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	appEnv := os.Getenv("ENV")
	isProd := appEnv == "production"

	DB_URI := os.Getenv("DATABASE_URL")
	if DB_URI == "" {
		panic(errors.New("'DATABASE_URL' may not be empty"))
	}
	config, err := pgxpool.ParseConfig(DB_URI)
	if err != nil {
		log.Fatalf("Could not read DSN: %v", err)
	}
	// config.MaxConns = int32(max(10, runtime.NumCPU() * 2))
	log.Printf("Set pgxpool MaxConns a: %d", config.MaxConns)

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatalf("Could not create pgx pool: %v", err)
	}
	defer pool.Close()

	err = pool.Ping(context.Background())
	if err != nil {
		log.Fatalf("Could not ping db: %v", err)
	}
	log.Println("Database connection (via pool) successful!")

	queries := db.New(pool)

	fiberConfig := fiber.Config{}
	reverseProxy := os.Getenv("REVERSE_PROXY")
	if reverseProxy == "true" {
		log.Println("Reverse proxy enabled")
		fiberConfig.ProxyHeader = fiber.HeaderXForwardedFor
	}

	app := fiber.New(fiberConfig)
	app.State().Set("queries", queries)
	app.State().Set("pool", pool)

	app.Use(logger.New())
	app.Use(favicon.New(favicon.Config{
		File: "./web/assets/favicon.png",
		URL:  "/favicon.png",
	}))

	cacheDuration := 1 * time.Hour
	if !isProd {
		cacheDuration = 10 * time.Second
	}
	app.Get("/static/*", static.New("./web/assets", static.Config{Compress: true, CacheDuration: cacheDuration}))

	if isProd {
		app.Use(compress.New())
	}

	app.Use(func(c fiber.Ctx) error {
		c.Locals("isProd", isProd)
		return c.Next()
	})

	sessionMiddleware, store := session.NewWithStore(session.Config{
		CookieSecure:    isProd,           // HTTPS only
		CookieHTTPOnly:  true,             // HTTP Only
		CookieSameSite:  "Lax",            // Same Site
		IdleTimeout:     30 * time.Minute, // Session timeout
		AbsoluteTimeout: 24 * time.Hour,   // Maximum session life
	})
	store.RegisterType(db.User{})
	app.Use(sessionMiddleware)

	srv := server.New(app)
	srv.RegisterRoutes()

	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = ":8080"
	}
	srv.Start(PORT)
}

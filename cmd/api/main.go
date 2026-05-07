package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/furkan/yol-asistani-api/config"
	"github.com/furkan/yol-asistani-api/internal/auth"
	"github.com/furkan/yol-asistani-api/internal/db"
	"github.com/furkan/yol-asistani-api/internal/route"
	"github.com/furkan/yol-asistani-api/internal/trip"
	"github.com/furkan/yol-asistani-api/internal/waypoint"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// DB bağlantısı — postgres hazır olana kadar retry
	var pool *pgxpool.Pool
	for i := range 10 {
		pool, err = db.NewPool(ctx, cfg.DatabaseURL)
		if err == nil {
			break
		}
		log.Printf("db not ready (attempt %d/10): %v", i+1, err)
		time.Sleep(3 * time.Second)
	}
	if pool == nil {
		log.Fatalf("db connect: gave up after 10 attempts")
	}
	defer pool.Close()

	// Migration'ları çalıştır
	if err := db.Migrate(context.Background(), pool, "./migrations"); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	// Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "yol-asistani-api",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"data":  nil,
				"error": err.Error(),
			})
		},
	})

	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${method} ${path} ${status} ${latency}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
	}))

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "time": time.Now()})
	})

	// --- Auth ---
	authSvc := auth.NewService(pool, cfg)
	authHandler := auth.NewHandler(authSvc)

	authGroup := app.Group("/auth")
	authGroup.Post("/register", authHandler.Register)
	authGroup.Post("/login", authHandler.Login)
	authGroup.Post("/refresh", authHandler.Refresh)
	authGroup.Post("/logout", authHandler.Logout)
	authGroup.Get("/me", auth.JWTProtected(cfg), authHandler.Me)

	// --- API v1 (JWT korumalı) ---
	jwtMw := auth.JWTProtected(cfg)
	v1 := app.Group("/api/v1", jwtMw)

	// Trips
	tripRepo := trip.NewRepository(pool)
	tripSvc := trip.NewService(tripRepo)
	tripHandler := trip.NewHandler(tripSvc)

	trips := v1.Group("/trips")
	trips.Get("/", tripHandler.List)
	trips.Post("/", tripHandler.Create)
	trips.Get("/:id", tripHandler.Get)
	trips.Put("/:id", tripHandler.Update)
	trips.Delete("/:id", tripHandler.Delete)

	// Waypoints
	wpRepo := waypoint.NewRepository(pool)
	wpSvc := waypoint.NewService(wpRepo)
	wpHandler := waypoint.NewHandler(wpSvc)

	trips.Get("/:id/waypoints", wpHandler.List)
	trips.Post("/:id/waypoints", wpHandler.Create)
	trips.Put("/:id/waypoints/:wid", wpHandler.Update)
	trips.Delete("/:id/waypoints/:wid", wpHandler.Delete)
	trips.Post("/:id/waypoints/reorder", wpHandler.Reorder)

	// Routes
	osrmClient := route.NewOSRMClient(cfg.OSRMBaseURL)
	routeCache, err := route.NewRouteCache(cfg.RedisURL)
	if err != nil {
		log.Fatalf("route cache: %v", err)
	}
	routeSvc := route.NewService(osrmClient, routeCache)
	routeHandler := route.NewHandler(routeSvc)

	routes := v1.Group("/routes")
	routes.Get("/alternatives", routeHandler.Alternatives)
	routes.Get("/segment", routeHandler.Segment)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		addr := fmt.Sprintf(":%s", cfg.Port)
		log.Printf("starting on %s (env=%s)", addr, cfg.Env)
		if err := app.Listen(addr); err != nil {
			log.Printf("server stopped: %v", err)
		}
	}()

	<-quit
	log.Println("shutting down...")
	if err := app.ShutdownWithTimeout(5 * time.Second); err != nil {
		log.Printf("shutdown error: %v", err)
	}
	log.Println("bye")
}

package main

import (
	"log"
	"net/http"

	gingerapp "github.com/ginger-framework/ginger/pkg/app"
	"github.com/ginger-framework/ginger/pkg/config"
	"github.com/ginger-framework/ginger/pkg/middleware"
	"github.com/ginger-framework/ginger/pkg/router"

	"github.com/ginger-framework/ginger/example/internal/api/handlers"
	"github.com/ginger-framework/ginger/example/internal/api/repositories"
	"github.com/ginger-framework/ginger/example/internal/api/services"
)

func main() {
	cfg, err := config.Load("configs/app.yaml")
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	app := gingerapp.New(cfg)

	// CORS for all routes
	app.Router.Use(middleware.CORS())

	// Wire up dependencies (pass a real *sql.DB in production)
	userRepo := repositories.NewUserRepository(nil)
	userSvc := services.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userSvc)

	// Mount routes under /api/v1
	v1 := app.Router.Group("/api/v1")
	userHandler.Register(v1)

	v1.GET("/ping", func(w http.ResponseWriter, r *http.Request) {
		router.JSON(w, http.StatusOK, map[string]string{"message": "pong"})
	})

	if err := app.Run(); err != nil {
		log.Fatalf("app: %v", err)
	}
}

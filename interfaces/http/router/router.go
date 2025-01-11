package router

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/goodfoodcesi/auth-api/infrastructure/jwt"
	"github.com/goodfoodcesi/auth-api/interfaces/http/handler"
	customMiddleware "github.com/goodfoodcesi/auth-api/interfaces/http/middleware"
	"go.uber.org/zap"
)

func NewRouter(
	userHandler *handler.UserHandler,
	logger *zap.Logger,
	tokenManager *jwt.TokenManager,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(customMiddleware.LoggerMiddleware(logger))
	r.Use(middleware.Timeout(60 * time.Second))

	rateLimiter := customMiddleware.NewRateLimiter(100, 100)
	r.Use(rateLimiter.Middleware)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Route("/auth", func(r chi.Router) {
		// Routes publiques
		r.Group(func(r chi.Router) {
			r.Post("/register", userHandler.Register)
			r.Post("/login", userHandler.Login)
		})

		// Routes protégées
		r.Group(func(r chi.Router) {
			r.Use(customMiddleware.AuthMiddleware(logger, tokenManager))

			// Routes utilisateur
			r.Get("/me", userHandler.GetProfile)
			r.Put("/me", userHandler.Update)
		})
	})

	return r
}

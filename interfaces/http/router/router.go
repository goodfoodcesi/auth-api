package router

import (
	"github.com/goodfoodcesi/auth-api/interfaces/http/response"
	"net/http"
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

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusNotFound, "Not found", nil)
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		response.Error(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
	})

	r.Route("/auth", func(r chi.Router) {
		// Routes publiques
		r.Group(func(r chi.Router) {
			r.Post("/register", userHandler.Register)
			r.Post("/login", userHandler.Login)
			r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
				response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
			})
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

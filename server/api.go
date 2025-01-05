package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/goodfoodcesi/auth-api/config"
	"github.com/goodfoodcesi/auth-api/controllers"
	db "github.com/goodfoodcesi/auth-api/db/sqlc"
	"github.com/goodfoodcesi/auth-api/logging"
	"github.com/goodfoodcesi/auth-api/routes"
	"github.com/goodfoodcesi/auth-api/utils"
	"github.com/jackc/pgx/v5"
)

func SetupApi(cfg config.Config, logManager *logging.LogManager) *gin.Engine {
	router := gin.New()

	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"message": "Route not found"})
	})

	router.Use(gin.Recovery())
	logManager.SetupHttp(router)

	// Configuration de la connexion à la base de données
	dbSource := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	// Connexion à la base de données avec pgx
	conn, err := pgx.Connect(context.Background(), dbSource)
	if err != nil {
		panic(fmt.Sprintf("Impossible de se connecter à la base de données: %v", err))
	}

	// Test de la connexion
	err = conn.Ping(context.Background())
	if err != nil {
		panic(fmt.Sprintf("Impossible de ping la base de données: %v", err))
	}

	// Création du contexte et des queries
	ctx := context.Background()
	queries := db.New(conn)

	// Initialisation des contrôleurs
	userController := controllers.NewUserController(queries, ctx)
	authController := controllers.NewAuthController(queries, ctx, cfg.APISecret, cfg.APIRefreshSecret)

	// Configuration des routes
	api := router.Group("/auth")

	// Initialisation des routes
	userRoutes := routes.NewRouteUser(*userController)
	userRoutes.UserRoute(api)

	authRoutes := routes.NewRouteAuth(*authController)

	authRoutes.AuthRoute(api)

	return router
}

func JWTInterceptor(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := utils.ExtractTokenFromHeader(c.Request.Header.Get("Authorization"))
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		claims, err := utils.ValidateToken(tokenString, cfg.APISecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.UserRole)
		c.Next()
	}
}

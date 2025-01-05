package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/goodfoodcesi/auth-api/controllers"
)

type AuthRoutes struct {
	authController controllers.AuthController
}

func NewRouteAuth(authController controllers.AuthController) AuthRoutes {
	return AuthRoutes{authController}
}

func (ar *AuthRoutes) AuthRoute(rg *gin.RouterGroup) {
	rg.POST("/login", ar.authController.Login)
}

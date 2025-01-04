package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/goodfoodcesi/auth-api/controllers"
)

type UserRoutes struct {
	userController controllers.UserController
}

func NewRouteUser(userController controllers.UserController) UserRoutes {
	return UserRoutes{userController}
}

func (ur *UserRoutes) UserRoute(rg *gin.RouterGroup) {
	router := rg.Group("users")
	router.POST("/", ur.userController.CreateUser)
	router.GET("/", ur.userController.GetAllUsers)
	router.GET("/:userId", ur.userController.GetUserById)
	router.PATCH("/:userId", ur.userController.UpdateUser)
	router.DELETE("/:userId", ur.userController.DeleteUserById)
}

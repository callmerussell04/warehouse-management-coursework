package handler

import (
	"log/slog"

	"warehouse-management-system/internal/middleware"
	"warehouse-management-system/internal/model"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func InitRoutes(r *gin.Engine, logger *slog.Logger, jwtSecret []byte, productH *ProductHandler, counterpartyH *CounterpartyHandler, orderH *OrderHandler, userH *UserHandler) {
	r.StaticFile("/openapi.yml", "./api/openapi.yml")

	docsUrl := ginSwagger.URL("/openapi.yml")
	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, docsUrl))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	authGroup := r.Group("/auth")
	{
		authGroup.POST("/login", userH.Login)
		authGroup.POST("/logout", userH.Logout)
		authGroup.POST("/request-otp", userH.RequestOTP)
		authGroup.POST("/reset-password", userH.ResetPassword)
		authGroup.POST("/forgot-username", userH.ForgotUsername)
		authGroup.POST("/refresh", userH.RefreshToken)
	}

	authMiddleware := middleware.AuthMiddleware(logger, jwtSecret)

	api := r.Group("/api")
	api.Use(authMiddleware)
	{
		userGroup := api.Group("/users")
		userGroup.Use(middleware.RequireRole(model.RoleAdmin))
		{
			userGroup.POST("", userH.Create)
			userGroup.GET("", userH.GetList)
			userGroup.DELETE("/:id", userH.Delete)
		}

		commonAccess := middleware.RequireRole(model.RoleAdmin, model.RoleWorker)

		productsGroup := api.Group("/products")
		productsGroup.Use(commonAccess)
		{
			productsGroup.POST("", productH.Create)
			productsGroup.GET("", productH.GetList)
			productsGroup.GET("/:id", productH.Get)
			productsGroup.PUT("/:id", productH.Update)
			productsGroup.DELETE("/:id", productH.Delete)
			productsGroup.POST("/:id/stock", productH.UpdateStock)
		}

		counterpartyGroup := api.Group("/counterparties")
		counterpartyGroup.Use(commonAccess)
		{
			counterpartyGroup.POST("", counterpartyH.Create)
			counterpartyGroup.GET("", counterpartyH.GetList)
			counterpartyGroup.GET("/:id", counterpartyH.Get)
			counterpartyGroup.PUT("/:id", counterpartyH.Update)
			counterpartyGroup.DELETE("/:id", counterpartyH.Delete)
		}

		orderGroup := api.Group("/orders")
		orderGroup.Use(commonAccess)
		{
			orderGroup.POST("", orderH.Create)
			orderGroup.GET("", orderH.GetList)
			orderGroup.GET("/:id", orderH.Get)
			orderGroup.PUT("/:id", orderH.Update)
			orderGroup.DELETE("/:id", orderH.Delete)
		}
	}
}

package handler

import (
	"github.com/gin-gonic/gin"
)

func InitRoutes(r *gin.Engine, productHandler *ProductHandler, counterpartyHandler *CounterpartyHandler, orderHandler *OrderHandler) {
	productsGroup := r.Group("/products")
	{
		productsGroup.POST("/", productHandler.Create)
		productsGroup.GET("/", productHandler.GetList)
		productsGroup.GET("/:id", productHandler.Get)
		productsGroup.PUT("/:id", productHandler.Update)
		productsGroup.DELETE("/:id", productHandler.Delete)
		productsGroup.POST("/:id/stock", productHandler.UpdateStock)
	}

	counterpartyGroup := r.Group("/counterparties")
	{
		counterpartyGroup.POST("/", counterpartyHandler.Create)
		counterpartyGroup.GET("/", counterpartyHandler.GetList)
		counterpartyGroup.GET("/:id", counterpartyHandler.Get)
		counterpartyGroup.PUT("/:id", counterpartyHandler.Update)
		counterpartyGroup.DELETE("/:id", counterpartyHandler.Delete)
	}

	orderGroup := r.Group("/orders")
	{
		orderGroup.POST("/", orderHandler.Create)
		orderGroup.GET("/", orderHandler.GetList)
		orderGroup.GET("/:id", orderHandler.Get)
		orderGroup.PUT("/:id", orderHandler.Update)
		orderGroup.DELETE("/:id", orderHandler.Delete)
	}
}

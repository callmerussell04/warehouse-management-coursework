package handler

import (
	"github.com/gin-gonic/gin"
)

func InitRoutes(r *gin.Engine, productHandler *ProductHandler) {
	productsGroup := r.Group("/products")
	{
		productsGroup.POST("/", productHandler.Create)
		productsGroup.GET("/", productHandler.GetList)
		productsGroup.GET("/:id", productHandler.Get)
		productsGroup.PUT("/:id", productHandler.Update)
		productsGroup.DELETE("/:id", productHandler.Delete)
		productsGroup.POST("/:id/stock", productHandler.UpdateStock)
	}
}

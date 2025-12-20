package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const maxRequestBodyBytes = 1 << 20

func LimitRequestBody() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRequestBodyBytes)
		}
		c.Next()
	}
}

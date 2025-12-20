package handler

import (
	"strconv"

	customErrors "warehouse-management-system/internal/errors"

	"github.com/gin-gonic/gin"
)

func parsePaging(c *gin.Context) (int, int, error) {
	page, pageErr := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, pageSizeErr := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if pageErr != nil || pageSizeErr != nil || page < 1 || pageSize < 1 || pageSize > 100 {
		return 0, 0, customErrors.NewAppError(customErrors.ErrInvalidInput, "'page' must be positive and 'pageSize' must be between 1 and 100")
	}
	return page, pageSize, nil
}

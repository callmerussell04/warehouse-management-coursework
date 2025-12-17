package handler

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ReportService interface {
	GenerateTurnoverPDF(ctx context.Context, fromStr, toStr string) ([]byte, error)
}

type ReportHandler struct {
	service ReportService
	logger  *slog.Logger
}

func NewReportHandler(service ReportService, logger *slog.Logger) *ReportHandler {
	return &ReportHandler{service: service, logger: logger}
}

func (h *ReportHandler) DownloadTurnoverReport(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")

	if from == "" || to == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "parameters 'from' and 'to' are required"})
		return
	}

	pdfBytes, err := h.service.GenerateTurnoverPDF(c.Request.Context(), from, to)
	if err != nil {
		RespondWithError(c, h.logger, err)
		return
	}

	filename := "turnover_report_" + time.Now().Format("20060102") + ".pdf"

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Length", strconv.Itoa(len(pdfBytes)))

	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

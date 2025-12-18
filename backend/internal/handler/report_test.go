package handler_test

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"warehouse-management-system/internal/handler"
	"warehouse-management-system/mocks"
)

func setupReportRouter(svc handler.ReportService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := handler.NewReportHandler(svc, logger)

	r.GET("/reports/turnover", h.DownloadTurnoverReport)

	return r
}

func TestReportHandler_DownloadTurnoverReport(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		prepare        func(m *mocks.ReportService)
		expectedStatus int
		checkBody      func(*testing.T, []byte)
		checkHeaders   func(*testing.T, http.Header)
	}{
		{
			name:  "Success",
			query: "?from=2025-01-01&to=2025-01-31",
			prepare: func(m *mocks.ReportService) {
				pdfContent := []byte("%PDF-1.4 mock content")
				m.EXPECT().GenerateTurnoverPDF(mock.Anything, "2025-01-01", "2025-01-31").Return(pdfContent, nil)
			},
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, body []byte) {
				assert.Equal(t, []byte("%PDF-1.4 mock content"), body)
			},
			checkHeaders: func(t *testing.T, h http.Header) {
				assert.Equal(t, "application/pdf", h.Get("Content-Type"))
				assert.Equal(t, "21", h.Get("Content-Length"))
				assert.Contains(t, h.Get("Content-Disposition"), "attachment; filename=turnover_report_")
			},
		},
		{
			name:  "Missing From",
			query: "?to=2025-01-31",
			prepare: func(m *mocks.ReportService) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:  "Missing To",
			query: "?from=2025-01-01",
			prepare: func(m *mocks.ReportService) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:  "Service Error",
			query: "?from=2025-01-01&to=2025-01-31",
			prepare: func(m *mocks.ReportService) {
				m.EXPECT().GenerateTurnoverPDF(mock.Anything, "2025-01-01", "2025-01-31").Return(nil, errors.New("generation failed"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := mocks.NewReportService(t)
			if tc.prepare != nil {
				tc.prepare(m)
			}

			r := setupReportRouter(m)
			req := httptest.NewRequest(http.MethodGet, "/reports/turnover"+tc.query, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			if tc.checkBody != nil {
				tc.checkBody(t, w.Body.Bytes())
			}
			if tc.checkHeaders != nil {
				tc.checkHeaders(t, w.Result().Header)
			}
		})
	}
}

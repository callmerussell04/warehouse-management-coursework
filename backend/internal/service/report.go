package service

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"warehouse-management-system/internal/dto"
	customErrors "warehouse-management-system/internal/errors"

	"github.com/go-pdf/fpdf"
)

type ReportRepository interface {
	GetTurnoverData(ctx context.Context, from, to time.Time) ([]dto.TurnoverReportItem, error)
}

type ReportService struct {
	repo   ReportRepository
	logger *slog.Logger
}

func NewReportService(repo ReportRepository, logger *slog.Logger) *ReportService {
	return &ReportService{repo: repo, logger: logger}
}

func (s *ReportService) GenerateTurnoverPDF(ctx context.Context, fromStr, toStr string) ([]byte, error) {
	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		return nil, customErrors.NewAppError(customErrors.ErrInvalidInput, "Invalid 'from' date (YYYY-MM-DD)")
	}

	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		return nil, customErrors.NewAppError(customErrors.ErrInvalidInput, "Invalid 'to' date (YYYY-MM-DD)")
	}
	to = to.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	data, err := s.repo.GetTurnoverData(ctx, from, to)
	if err != nil {
		return nil, err
	}

	pdf := fpdf.New("P", "mm", "A4", "")

	pdf.AddUTF8Font("CustomFont", "", "assets/arial.ttf")
	pdf.SetFont("CustomFont", "", 12)

	pdf.AddPage()

	pdf.SetFont("CustomFont", "", 16)
	title := fmt.Sprintf("Обороты (%s - %s)", from.Format("2006-01-02"), to.Format("2006-01-02"))
	pdf.Cell(40, 10, title)
	pdf.Ln(12)

	pdf.SetFont("CustomFont", "", 10)
	pdf.SetFillColor(240, 240, 240)

	headers := []string{"Товар", "Артикул", "Нач.ост", "Приход", "Расход", "Кон.ост"}
	widths := []float64{60, 30, 25, 25, 25, 25}

	for i, h := range headers {
		pdf.CellFormat(widths[i], 10, h, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetFont("CustomFont", "", 10)
	pdf.SetFillColor(255, 255, 255)

	for _, item := range data {
		pdf.CellFormat(widths[0], 8, item.ProductName, "1", 0, "L", false, 0, "")
		pdf.CellFormat(widths[1], 8, item.SKU, "1", 0, "L", false, 0, "")
		pdf.CellFormat(widths[2], 8, strconv.Itoa(item.StartBalance), "1", 0, "R", false, 0, "")
		pdf.CellFormat(widths[3], 8, strconv.Itoa(item.Income), "1", 0, "R", false, 0, "")
		pdf.CellFormat(widths[4], 8, strconv.Itoa(item.Expense), "1", 0, "R", false, 0, "")
		pdf.CellFormat(widths[5], 8, strconv.Itoa(item.EndBalance), "1", 0, "R", false, 0, "")
		pdf.Ln(-1)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		s.logger.Error("failed to generate PDF", "error", err)
		return nil, customErrors.ErrInternal
	}

	return buf.Bytes(), nil
}

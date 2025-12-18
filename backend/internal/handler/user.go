package handler

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"warehouse-management-system/internal/dto"
	customErrors "warehouse-management-system/internal/errors"
	"warehouse-management-system/internal/model"
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=UserService --output=../../mocks --outpkg=mocks --with-expecter=true
type UserService interface {
	CreateUser(ctx context.Context, username, email, fullName, role string) (*model.User, error)
	Login(ctx context.Context, username, password string) (accessToken, refreshToken string, user *model.User, err error)
	RecoverUsername(ctx context.Context, email string) error
	GenerateAndSendOTP(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, email, code, newPassword string) error
	RefreshToken(ctx context.Context, tokenString string) (string, error)
	GetList(ctx context.Context, page, pageSize int) ([]*model.User, int, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type UserHandler struct {
	service UserService
	logger  *slog.Logger
}

func NewUserHandler(service UserService, logger *slog.Logger) *UserHandler {
	return &UserHandler{service: service, logger: logger}
}

func (h *UserHandler) mapUserToResponse(u *model.User) dto.UserResponse {
	return dto.UserResponse{
		ID:       u.ID.String(),
		Username: u.Username,
		Email:    u.Email,
		FullName: u.FullName,
		Role:     string(u.Role),
		IsActive: u.IsActive,
	}
}

func (h *UserHandler) Create(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	user, err := h.service.CreateUser(c.Request.Context(), req.Username, req.Email, req.FullName, req.Role)
	if err != nil {
		RespondWithError(c, h.logger, err)
		return
	}

	c.JSON(http.StatusCreated, h.mapUserToResponse(user))
}

func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accessToken, refreshToken, user, err := h.service.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		RespondWithError(c, h.logger, err)
		return
	}

	const refreshCookieDuration = 7 * 24 * time.Hour
	c.SetCookie(
		"refresh_token",
		refreshToken,
		int(refreshCookieDuration.Seconds()),
		"/",
		c.Request.Host,
		true,
		true,
	)

	c.JSON(http.StatusOK, dto.LoginResponse{
		AccessToken: accessToken,
		User:        h.mapUserToResponse(user),
	})
}

func (h *UserHandler) Logout(c *gin.Context) {
	c.SetCookie(
		"refresh_token",
		"",
		-1,
		"/",
		c.Request.Host,
		true,
		true,
	)

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func (h *UserHandler) ForgotUsername(c *gin.Context) {
	var req dto.ForgotUsernameRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	err := h.service.RecoverUsername(c.Request.Context(), req.Email)
	if err != nil {
		RespondWithError(c, h.logger, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "If the email is registered, the username has been sent to it."})
}

func (h *UserHandler) RequestOTP(c *gin.Context) {
	var req dto.SendOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.service.GenerateAndSendOTP(c.Request.Context(), req.Email)

	if err != nil {
		h.logger.Error("OTP generation failed", "error", err, "email", req.Email)
	}

	c.JSON(http.StatusOK, gin.H{"message": "If the user exists, an OTP has been sent to the email."})
}

func (h *UserHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.service.ResetPassword(c.Request.Context(), req.Email, req.OTP, req.NewPassword)
	if err != nil {
		RespondWithError(c, h.logger, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully. Account is now active."})
}

func (h *UserHandler) RefreshToken(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token not provided"})
		return
	}

	newAccessToken, err := h.service.RefreshToken(c.Request.Context(), refreshToken)
	if err != nil {
		c.SetCookie("refresh_token", "", -1, "/", c.Request.Host, true, true)
		RespondWithError(c, h.logger, customErrors.ErrUnauthorized)
		return
	}

	c.JSON(http.StatusOK, dto.RefreshTokenResponse{
		AccessToken: newAccessToken,
	})
}

func (h *UserHandler) GetList(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	users, totalCount, err := h.service.GetList(c.Request.Context(), page, pageSize)
	if err != nil {
		RespondWithError(c, h.logger, err)
		return
	}

	userDtos := make([]dto.UserResponse, len(users))
	for i, user := range users {
		userDtos[i] = h.mapUserToResponse(user)
	}

	response := dto.PagedUsers{
		Paging: dto.Paging{
			Page:  page,
			Size:  pageSize,
			Total: totalCount,
		},
		Items: userDtos,
	}

	c.JSON(http.StatusOK, response)
}

func (h *UserHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID format"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		RespondWithError(c, h.logger, err)
		return
	}

	c.Status(http.StatusNoContent)
}

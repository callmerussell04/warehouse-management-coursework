package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	customErrors "warehouse-management-system/internal/errors"
	"warehouse-management-system/internal/model"
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=UserRepository --output=../../mocks --outpkg=mocks --with-expecter=true
type UserRepository interface {
	CreateUser(ctx context.Context, u *model.User) error
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	UpdatePasswordAndActivate(ctx context.Context, email, passwordHash string) error
	Update(ctx context.Context, u *model.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetList(ctx context.Context, limit, offset int) ([]*model.User, int, error)
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=OTPRepository --output=../../mocks --outpkg=mocks --with-expecter=true
type OTPRepository interface {
	Save(ctx context.Context, email, code string, duration time.Duration) error
	Verify(ctx context.Context, email, code string, maxAttempts int) (bool, error)
}

type RefreshSessionRepository interface {
	Save(ctx context.Context, token string, userID uuid.UUID, ttl time.Duration) error
	Rotate(ctx context.Context, oldToken, newToken string, ttl time.Duration) (uuid.UUID, error)
	Revoke(ctx context.Context, token string) error
	RevokeAll(ctx context.Context, userID uuid.UUID) error
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=RefreshSessionRepository --output=../../mocks --outpkg=mocks --with-expecter=true

//go:generate go run github.com/vektra/mockery/v2@latest --name=NotificationService --output=../../mocks --outpkg=mocks --with-expecter=true
type NotificationService interface {
	SendEmail(to, subject, body string) error
}

type UserService struct {
	userRepository  UserRepository
	otpRepository   OTPRepository
	notifier        NotificationService
	refreshSessions RefreshSessionRepository
	logger          *slog.Logger
	jwtSecret       []byte
}

func NewUserService(userRepo UserRepository, otpRepo OTPRepository, notifier NotificationService, refreshSessions RefreshSessionRepository, logger *slog.Logger, jwtSecret []byte) *UserService {
	return &UserService{
		userRepository:  userRepo,
		otpRepository:   otpRepo,
		notifier:        notifier,
		refreshSessions: refreshSessions,
		logger:          logger,
		jwtSecret:       jwtSecret,
	}
}

func (s *UserService) EnsureAdminExists(ctx context.Context, username, email, password string) error {
	_, err := s.userRepository.GetByUsername(ctx, username)
	if errors.Is(err, customErrors.ErrNotFound) {
		hashedPassword, hashErr := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if hashErr != nil {
			return fmt.Errorf("failed to hash admin password: %w", hashErr)
		}
		admin := &model.User{
			ID:           uuid.New(),
			Username:     username,
			Email:        email,
			PasswordHash: string(hashedPassword),
			FullName:     "System Administrator",
			Role:         model.RoleAdmin,
			IsActive:     true,
		}
		if err := s.userRepository.CreateUser(ctx, admin); err != nil {
			return fmt.Errorf("failed to create admin user: %w", err)
		}
		s.logger.Info("Administrator account created", "username", username)
	} else if err != nil {
		return fmt.Errorf("failed to check administrator account: %w", err)
	}

	legacyAdmin, err := s.userRepository.GetByUsername(ctx, "admin")
	if errors.Is(err, customErrors.ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to check legacy administrator account: %w", err)
	}
	if bcrypt.CompareHashAndPassword([]byte(legacyAdmin.PasswordHash), []byte("admin")) == nil {
		hashedPassword, hashErr := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if hashErr != nil {
			return fmt.Errorf("failed to hash replacement admin password: %w", hashErr)
		}
		if err := s.userRepository.UpdatePasswordAndActivate(ctx, legacyAdmin.Email, string(hashedPassword)); err != nil {
			return fmt.Errorf("failed to rotate legacy administrator password: %w", err)
		}
		s.logger.Info("Legacy administrator password rotated")
	}
	return nil
}

func (s *UserService) CreateUser(ctx context.Context, username, email, fullName, roleStr string) (*model.User, error) {
	username = strings.TrimSpace(username)
	email = normalizeEmail(email)
	fullName = strings.TrimSpace(fullName)
	if username == "" || utf8.RuneCountInString(username) > 255 ||
		email == "" || utf8.RuneCountInString(email) > 255 || !validEmail(email) ||
		fullName == "" || utf8.RuneCountInString(fullName) > 255 {
		return nil, customErrors.ErrInvalidInput
	}
	role := model.Role(roleStr)
	if role != model.RoleAdmin && role != model.RoleWorker {
		return nil, customErrors.NewAppError(customErrors.ErrInvalidInput, "Invalid role. Allowed: admin, worker")
	}

	if _, err := s.userRepository.GetByUsername(ctx, username); err == nil {
		return nil, customErrors.ErrAlreadyExists
	} else if !errors.Is(err, customErrors.ErrNotFound) {
		return nil, err
	}
	if _, err := s.userRepository.GetByEmail(ctx, email); err == nil {
		return nil, customErrors.ErrAlreadyExists
	} else if !errors.Is(err, customErrors.ErrNotFound) {
		return nil, err
	}

	u := &model.User{
		ID:       uuid.New(),
		Username: username,
		Email:    email,
		FullName: fullName,
		Role:     role,
		IsActive: false,
	}

	if err := s.userRepository.CreateUser(ctx, u); err != nil {
		return nil, err
	}

	s.logger.Info("User created successfully", "id", u.ID, "role", u.Role)
	return u, nil
}

func (s *UserService) Login(ctx context.Context, username, password string) (accessToken, refreshToken string, user *model.User, err error) {
	u, err := s.userRepository.GetByUsername(ctx, username)
	if errors.Is(err, customErrors.ErrNotFound) {
		const dummyHash = "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
		_ = bcrypt.CompareHashAndPassword([]byte(dummyHash), []byte(password))
		return "", "", nil, invalidCredentialsError()
	}
	if err != nil {
		return "", "", nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil || !u.IsActive {
		return "", "", nil, invalidCredentialsError()
	}

	accessToken, err = s.generateAccessToken(u)
	if err != nil {
		s.logger.Error("failed to generate access token", "error", err)
		return "", "", nil, customErrors.ErrInternal
	}

	refreshToken, err = generateRefreshToken()
	if err != nil {
		s.logger.Error("failed to generate refresh token", "error", err)
		return "", "", nil, customErrors.ErrInternal
	}

	if err := s.refreshSessions.Save(ctx, refreshToken, u.ID, 7*24*time.Hour); err != nil {
		return "", "", nil, err
	}
	return accessToken, refreshToken, u, nil
}

func invalidCredentialsError() error {
	return customErrors.NewAppError(customErrors.ErrUnauthorized, "Invalid credentials")
}

func (s *UserService) RecoverUsername(ctx context.Context, email string) error {
	email = normalizeEmail(email)
	user, err := s.userRepository.GetByEmail(ctx, email)
	if errors.Is(err, customErrors.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	subject := "Warehouse System - Username Recovery"
	body := fmt.Sprintf("Hello %s,\n\nYou requested to recover your username.\nYour username is: %s\n\nIf you did not request this, please ignore this email.", user.FullName, user.Username)

	if err := s.notifier.SendEmail(email, subject, body); err != nil {
		s.logger.Error("Failed to send username recovery email", "error", err)
		return customErrors.ErrInternal
	}

	return nil
}

func (s *UserService) RefreshToken(ctx context.Context, tokenString string) (string, string, error) {
	newRefreshToken, err := generateRefreshToken()
	if err != nil {
		return "", "", customErrors.ErrInternal
	}
	userID, err := s.refreshSessions.Rotate(ctx, tokenString, newRefreshToken, 7*24*time.Hour)
	if errors.Is(err, customErrors.ErrUnauthorized) {
		return "", "", customErrors.NewAppError(customErrors.ErrUnauthorized, "Invalid or expired refresh token")
	}
	if err != nil {
		return "", "", err
	}
	user, dbErr := s.userRepository.GetByID(ctx, userID)
	if dbErr != nil {
		_ = s.refreshSessions.Revoke(ctx, newRefreshToken)
		return "", "", customErrors.NewAppError(customErrors.ErrUnauthorized, "User not found")
	}
	if !user.IsActive {
		_ = s.refreshSessions.Revoke(ctx, newRefreshToken)
		return "", "", customErrors.NewAppError(customErrors.ErrUnauthorized, "User account is inactive")
	}

	newAccessToken, genErr := s.generateAccessToken(user)
	if genErr != nil {
		s.logger.Error("failed to generate new access token", "error", genErr)
		return "", "", customErrors.ErrInternal
	}
	return newAccessToken, newRefreshToken, nil
}

func (s *UserService) Logout(ctx context.Context, refreshToken string) error {
	return s.refreshSessions.Revoke(ctx, refreshToken)
}

func (s *UserService) GenerateAndSendOTP(ctx context.Context, email string) error {
	email = normalizeEmail(email)
	if _, err := s.userRepository.GetByEmail(ctx, email); errors.Is(err, customErrors.ErrNotFound) {
		return nil
	} else if err != nil {
		return err
	}

	code, err := generateRandomCode(6)
	if err != nil {
		return customErrors.ErrInternal
	}

	if err := s.otpRepository.Save(ctx, email, code, 15*time.Minute); err != nil {
		return err
	}

	subject := "Warehouse System - Security Code"
	body := fmt.Sprintf("Your security code is: %s. It expires in 15 minutes.", code)

	return s.notifier.SendEmail(email, subject, body)
}

func (s *UserService) ResetPassword(ctx context.Context, email, code, newPassword string) error {
	if len(newPassword) < 12 || len(newPassword) > 72 {
		return customErrors.NewAppError(customErrors.ErrInvalidInput, "Password must contain between 12 and 72 bytes")
	}
	email = normalizeEmail(email)
	valid, err := s.otpRepository.Verify(ctx, email, code, 5)
	if errors.Is(err, customErrors.ErrNotFound) {
		return customErrors.NewAppError(customErrors.ErrInvalidInput, "Invalid or expired OTP code")
	}
	if err != nil {
		return err
	}
	if !valid {
		return customErrors.NewAppError(customErrors.ErrInvalidInput, "Invalid or expired OTP code")
	}
	user, err := s.userRepository.GetByEmail(ctx, email)
	if err != nil {
		return err
	}
	if err := s.refreshSessions.RevokeAll(ctx, user.ID); err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("failed to hash password", "error", err)
		return customErrors.ErrInternal
	}

	if err := s.userRepository.UpdatePasswordAndActivate(ctx, email, string(hashedPassword)); err != nil {
		return err
	}

	s.logger.Info("Password updated successfully", "user_id", user.ID)
	return nil
}

func (s *UserService) GetList(ctx context.Context, page, pageSize int) ([]*model.User, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	} else if pageSize > 100 {
		pageSize = 100
	}
	limit := pageSize
	offset := (page - 1) * pageSize
	return s.userRepository.GetList(ctx, limit, offset)
}

func (s *UserService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.userRepository.Delete(ctx, id); err != nil {
		return err
	}
	s.logger.Info("User deleted", "id", id)
	return nil
}

func (s *UserService) generateAccessToken(user *model.User) (string, error) {
	now := time.Now()
	claims := model.UserClaims{
		UserID:   user.ID.String(),
		Username: user.Username,
		Role:     string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "warehouse-system",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func generateRefreshToken() (string, error) {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func validEmail(email string) bool {
	address, err := mail.ParseAddress(email)
	return err == nil && address.Address == email
}

func generateRandomCode(length int) (string, error) {
	const digits = "0123456789"
	ret := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		ret[i] = digits[num.Int64()]
	}
	return string(ret), nil
}

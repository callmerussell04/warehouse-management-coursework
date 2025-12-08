package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	customErrors "warehouse-management-system/internal/errors"
	"warehouse-management-system/internal/model"
)

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

type OTPRepository interface {
	Save(ctx context.Context, email, code string, duration time.Duration) error
	Get(ctx context.Context, email string) (string, error)
	Delete(ctx context.Context, email string) error
}

type NotificationService interface {
	SendEmail(to, subject, body string) error
}

type UserService struct {
	userRepository UserRepository
	otpRepository  OTPRepository
	notifier       NotificationService
	logger         *slog.Logger
	jwtSecret      []byte
}

func NewUserService(userRepo UserRepository, otpRepo OTPRepository, notifier NotificationService, logger *slog.Logger, jwtSecret []byte) *UserService {
	return &UserService{
		userRepository: userRepo,
		otpRepository:  otpRepo,
		notifier:       notifier,
		logger:         logger,
		jwtSecret:      jwtSecret,
	}
}

func (s *UserService) EnsureAdminExists(ctx context.Context) error {
	const (
		adminUsername = "admin"
		adminPassword = "admin"
		adminEmail    = "admin@system.local"
	)

	_, err := s.userRepository.GetByUsername(ctx, adminUsername)
	if err == nil {
		s.logger.Info("Admin user already exists, skipping creation")
		return nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash admin password: %w", err)
	}

	admin := &model.User{
		ID:           uuid.New(),
		Username:     adminUsername,
		Email:        adminEmail,
		PasswordHash: string(hashedPassword),
		FullName:     "System Administrator",
		Role:         model.RoleAdmin,
		IsActive:     true,
	}

	if err := s.userRepository.CreateUser(ctx, admin); err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	s.logger.Info("Default admin user created successfully", "username", adminUsername, "password", adminPassword)
	return nil
}

func (s *UserService) CreateUser(ctx context.Context, username, email, fullName, roleStr string) (*model.User, error) {
	role := model.Role(roleStr)
	if role != model.RoleAdmin && role != model.RoleWorker {
		return nil, customErrors.NewAppError(customErrors.ErrInvalidInput, "Invalid role. Allowed: admin, worker")
	}

	if _, err := s.userRepository.GetByUsername(ctx, username); err == nil {
		return nil, customErrors.ErrAlreadyExists
	}
	if _, err := s.userRepository.GetByEmail(ctx, email); err == nil {
		return nil, customErrors.ErrAlreadyExists
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
	if err != nil {
		return "", "", nil, customErrors.NewAppError(customErrors.ErrInvalidInput, "Invalid credentials")
	}

	if !u.IsActive {
		return "", "", nil, customErrors.NewAppError(customErrors.ErrUnauthorized, "Account not activated")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return "", "", nil, customErrors.NewAppError(customErrors.ErrInvalidInput, "Invalid credentials")
	}

	accessToken, err = s.generateToken(u, 15*time.Minute)
	if err != nil {
		s.logger.Error("failed to generate access token", "error", err)
		return "", "", nil, customErrors.ErrInternal
	}

	refreshToken, err = s.generateToken(u, 7*24*time.Hour)
	if err != nil {
		s.logger.Error("failed to generate refresh token", "error", err)
		return "", "", nil, customErrors.ErrInternal
	}

	return accessToken, refreshToken, u, nil
}

func (s *UserService) RefreshToken(ctx context.Context, tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &model.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return "", customErrors.NewAppError(customErrors.ErrUnauthorized, "Invalid or expired refresh token")
	}

	claims, ok := token.Claims.(*model.UserClaims)
	if !ok || !token.Valid {
		return "", customErrors.NewAppError(customErrors.ErrUnauthorized, "Invalid token claims")
	}

	userID, parseErr := uuid.Parse(claims.UserID)
	if parseErr != nil {
		return "", customErrors.NewAppError(customErrors.ErrUnauthorized, "Invalid User ID in token")
	}

	user, dbErr := s.userRepository.GetByID(ctx, userID)
	if dbErr != nil {
		return "", customErrors.NewAppError(customErrors.ErrUnauthorized, "User not found")
	}
	if !user.IsActive {
		return "", customErrors.NewAppError(customErrors.ErrUnauthorized, "User account is inactive")
	}

	newAccessToken, genErr := s.generateToken(user, 15*time.Minute)
	if genErr != nil {
		s.logger.Error("failed to generate new access token", "error", genErr)
		return "", customErrors.ErrInternal
	}

	return newAccessToken, nil
}

func (s *UserService) GenerateAndSendOTP(ctx context.Context, email string) error {
	if _, err := s.userRepository.GetByEmail(ctx, email); err != nil {
		s.logger.Warn("OTP requested for non-existent email", "email", email)
		return nil
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
	storedCode, err := s.otpRepository.Get(ctx, email)

	if err != nil || storedCode == "" {
		return customErrors.NewAppError(customErrors.ErrInvalidInput, "Invalid or expired OTP code")
	}

	if storedCode != code {
		return customErrors.NewAppError(customErrors.ErrInvalidInput, "Incorrect OTP code")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("failed to hash password", "error", err)
		return customErrors.ErrInternal
	}

	if err := s.userRepository.UpdatePasswordAndActivate(ctx, email, string(hashedPassword)); err != nil {
		return err
	}

	_ = s.otpRepository.Delete(ctx, email)
	s.logger.Info("Password updated successfully", "email", email)
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

func (s *UserService) generateToken(user *model.User, duration time.Duration) (string, error) {
	claims := model.UserClaims{
		UserID:   user.ID.String(),
		Username: user.Username,
		Role:     string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "warehouse-system",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
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

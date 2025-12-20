package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	customErrors "warehouse-management-system/internal/errors"
	"warehouse-management-system/internal/model"
	"warehouse-management-system/internal/service"
	"warehouse-management-system/mocks"
)

var testSecret = []byte("test-secret-key")

func TestUserService_EnsureAdminExists(t *testing.T) {
	t.Run("creates configured administrator", func(t *testing.T) {
		um := mocks.NewUserRepository(t)
		um.EXPECT().GetByUsername(mock.Anything, "course-admin").Return(nil, customErrors.ErrNotFound)
		um.EXPECT().CreateUser(mock.Anything, mock.MatchedBy(func(user *model.User) bool {
			return user.Username == "course-admin" && user.Email == "admin@example.com" &&
				user.Role == model.RoleAdmin && user.IsActive &&
				bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("strong-password-123")) == nil
		})).Return(nil)
		um.EXPECT().GetByUsername(mock.Anything, "admin").Return(nil, customErrors.ErrNotFound)

		svc := service.NewUserService(um, nil, nil, nil, newDiscardLogger(), testSecret)
		assert.NoError(t, svc.EnsureAdminExists(context.Background(), "course-admin", "admin@example.com", "strong-password-123"))
	})

	t.Run("rotates only legacy admin password", func(t *testing.T) {
		legacyHash, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
		require.NoError(t, err)
		legacy := &model.User{Username: "admin", Email: "legacy@example.com", PasswordHash: string(legacyHash)}
		um := mocks.NewUserRepository(t)
		um.EXPECT().GetByUsername(mock.Anything, "admin").Return(legacy, nil).Twice()
		um.EXPECT().UpdatePasswordAndActivate(mock.Anything, legacy.Email, mock.MatchedBy(func(hash string) bool {
			return bcrypt.CompareHashAndPassword([]byte(hash), []byte("replacement-password")) == nil
		})).Return(nil)

		svc := service.NewUserService(um, nil, nil, nil, newDiscardLogger(), testSecret)
		assert.NoError(t, svc.EnsureAdminExists(context.Background(), "admin", "admin@example.com", "replacement-password"))
	})

	t.Run("preserves a non-legacy password", func(t *testing.T) {
		secureHash, err := bcrypt.GenerateFromPassword([]byte("already-secure-password"), bcrypt.DefaultCost)
		require.NoError(t, err)
		admin := &model.User{Username: "admin", PasswordHash: string(secureHash)}
		um := mocks.NewUserRepository(t)
		um.EXPECT().GetByUsername(mock.Anything, "admin").Return(admin, nil).Twice()

		svc := service.NewUserService(um, nil, nil, nil, newDiscardLogger(), testSecret)
		assert.NoError(t, svc.EnsureAdminExists(context.Background(), "admin", "admin@example.com", "replacement-password"))
	})
}

func TestUserService_CreateUser(t *testing.T) {
	type args struct {
		username string
		email    string
		fullName string
		role     string
	}
	tests := []struct {
		name      string
		args      args
		prepare   func(um *mocks.UserRepository)
		wantError error
		checkRes  func(*testing.T, *model.User)
	}{
		{
			name: "Success Admin",
			args: args{username: "admin", email: "admin@mail.com", fullName: "Admin User", role: "admin"},
			prepare: func(um *mocks.UserRepository) {
				um.EXPECT().GetByUsername(mock.Anything, "admin").Return(nil, customErrors.ErrNotFound)
				um.EXPECT().GetByEmail(mock.Anything, "admin@mail.com").Return(nil, customErrors.ErrNotFound)

				um.EXPECT().CreateUser(mock.Anything, mock.MatchedBy(func(u *model.User) bool {
					return u.Username == "admin" && u.Role == model.RoleAdmin && !u.IsActive
				})).Return(nil)
			},
			wantError: nil,
			checkRes: func(t *testing.T, u *model.User) {
				assert.NotNil(t, u)
				assert.Equal(t, model.RoleAdmin, u.Role)
			},
		},
		{
			name: "Invalid Role",
			args: args{username: "u", email: "e", fullName: "f", role: "superuser"},
			prepare: func(um *mocks.UserRepository) {
			},
			wantError: customErrors.ErrInvalidInput,
		},
		{
			name: "Username Exists",
			args: args{username: "exist", email: "new@mail.com", fullName: "f", role: "worker"},
			prepare: func(um *mocks.UserRepository) {
				um.EXPECT().GetByUsername(mock.Anything, "exist").Return(&model.User{}, nil)
			},
			wantError: customErrors.ErrAlreadyExists,
		},
		{
			name: "Email Exists",
			args: args{username: "new", email: "exist@mail.com", fullName: "f", role: "worker"},
			prepare: func(um *mocks.UserRepository) {
				um.EXPECT().GetByUsername(mock.Anything, "new").Return(nil, customErrors.ErrNotFound)
				um.EXPECT().GetByEmail(mock.Anything, "exist@mail.com").Return(&model.User{}, nil)
			},
			wantError: customErrors.ErrAlreadyExists,
		},
		{
			name: "Repo Create Error",
			args: args{username: "u", email: "u@example.com", fullName: "f", role: "worker"},
			prepare: func(um *mocks.UserRepository) {
				um.EXPECT().GetByUsername(mock.Anything, "u").Return(nil, customErrors.ErrNotFound)
				um.EXPECT().GetByEmail(mock.Anything, "u@example.com").Return(nil, customErrors.ErrNotFound)
				um.EXPECT().CreateUser(mock.Anything, mock.Anything).Return(errors.New("db fail"))
			},
			wantError: errors.New("db fail"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			um := mocks.NewUserRepository(t)
			if tc.prepare != nil {
				tc.prepare(um)
			}

			svc := service.NewUserService(um, nil, nil, nil, newDiscardLogger(), testSecret)
			got, err := svc.CreateUser(context.Background(), tc.args.username, tc.args.email, tc.args.fullName, tc.args.role)

			if tc.wantError != nil {
				assert.Error(t, err)
				if errors.Is(tc.wantError, customErrors.ErrInvalidInput) || errors.Is(tc.wantError, customErrors.ErrAlreadyExists) {
					assert.ErrorIs(t, err, tc.wantError)
				} else {
					assert.Contains(t, err.Error(), tc.wantError.Error())
				}
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				if tc.checkRes != nil {
					tc.checkRes(t, got)
				}
			}
		})
	}
}

func TestUserService_Login(t *testing.T) {
	pass := "secret123"
	hashedPass, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)

	type args struct {
		username string
		password string
	}
	tests := []struct {
		name        string
		args        args
		prepare     func(um *mocks.UserRepository, sm *mocks.RefreshSessionRepository)
		wantError   error
		checkTokens bool
	}{
		{
			name: "Success",
			args: args{username: "user", password: pass},
			prepare: func(um *mocks.UserRepository, sm *mocks.RefreshSessionRepository) {
				u := &model.User{
					ID:           uuid.New(),
					Username:     "user",
					PasswordHash: string(hashedPass),
					IsActive:     true,
					Role:         model.RoleWorker,
				}
				um.EXPECT().GetByUsername(mock.Anything, "user").Return(u, nil)
				sm.EXPECT().Save(mock.Anything, mock.Anything, u.ID, 7*24*time.Hour).Return(nil)
			},
			wantError:   nil,
			checkTokens: true,
		},
		{
			name: "User Not Found",
			args: args{username: "ghost", password: pass},
			prepare: func(um *mocks.UserRepository, sm *mocks.RefreshSessionRepository) {
				um.EXPECT().GetByUsername(mock.Anything, "ghost").Return(nil, customErrors.ErrNotFound)
			},
			wantError: customErrors.ErrUnauthorized,
		},
		{
			name: "User Inactive",
			args: args{username: "inactive", password: pass},
			prepare: func(um *mocks.UserRepository, sm *mocks.RefreshSessionRepository) {
				u := &model.User{Username: "inactive", IsActive: false}
				um.EXPECT().GetByUsername(mock.Anything, "inactive").Return(u, nil)
			},
			wantError: customErrors.ErrUnauthorized,
		},
		{
			name: "Wrong Password",
			args: args{username: "user", password: "wrong"},
			prepare: func(um *mocks.UserRepository, sm *mocks.RefreshSessionRepository) {
				u := &model.User{
					Username:     "user",
					PasswordHash: string(hashedPass),
					IsActive:     true,
				}
				um.EXPECT().GetByUsername(mock.Anything, "user").Return(u, nil)
			},
			wantError: customErrors.ErrUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			um := mocks.NewUserRepository(t)
			sm := mocks.NewRefreshSessionRepository(t)
			tc.prepare(um, sm)

			svc := service.NewUserService(um, nil, nil, sm, newDiscardLogger(), testSecret)
			at, rt, user, err := svc.Login(context.Background(), tc.args.username, tc.args.password)

			if tc.wantError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tc.wantError)
				if errors.Is(tc.wantError, customErrors.ErrUnauthorized) {
					assert.Equal(t, "Invalid credentials", err.Error())
				}
				assert.Empty(t, at)
				assert.Empty(t, rt)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, at)
				assert.NotEmpty(t, rt)
				assert.NotNil(t, user)
			}
		})
	}
}

func TestUserService_GenerateAndSendOTP(t *testing.T) {
	email := "test@mail.com"
	tests := []struct {
		name    string
		email   string
		prepare func(um *mocks.UserRepository, om *mocks.OTPRepository, nm *mocks.NotificationService)
		wantErr bool
	}{
		{
			name:  "Success",
			email: email,
			prepare: func(um *mocks.UserRepository, om *mocks.OTPRepository, nm *mocks.NotificationService) {
				um.EXPECT().GetByEmail(mock.Anything, email).Return(&model.User{}, nil)

				om.EXPECT().Save(mock.Anything, email, mock.MatchedBy(func(code string) bool {
					return len(code) == 6
				}), 15*time.Minute).Return(nil)

				nm.EXPECT().SendEmail(email, mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name:  "User Not Found (Security)",
			email: "ghost@mail.com",
			prepare: func(um *mocks.UserRepository, om *mocks.OTPRepository, nm *mocks.NotificationService) {
				um.EXPECT().GetByEmail(mock.Anything, "ghost@mail.com").Return(nil, customErrors.ErrNotFound)
			},
			wantErr: false,
		},
		{
			name:  "OTP Save Error",
			email: email,
			prepare: func(um *mocks.UserRepository, om *mocks.OTPRepository, nm *mocks.NotificationService) {
				um.EXPECT().GetByEmail(mock.Anything, email).Return(&model.User{}, nil)
				om.EXPECT().Save(mock.Anything, email, mock.Anything, mock.Anything).Return(errors.New("redis fail"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			um := mocks.NewUserRepository(t)
			om := mocks.NewOTPRepository(t)
			nm := mocks.NewNotificationService(t)

			tc.prepare(um, om, nm)

			svc := service.NewUserService(um, om, nm, nil, newDiscardLogger(), testSecret)
			err := svc.GenerateAndSendOTP(context.Background(), tc.email)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserService_RecoverUsername(t *testing.T) {
	email := "test@mail.com"
	tests := []struct {
		name    string
		email   string
		prepare func(um *mocks.UserRepository, nm *mocks.NotificationService)
		wantErr bool
	}{
		{
			name:  "Success",
			email: email,
			prepare: func(um *mocks.UserRepository, nm *mocks.NotificationService) {
				user := &model.User{Username: "myuser", FullName: "Test User"}
				um.EXPECT().GetByEmail(mock.Anything, email).Return(user, nil)
				nm.EXPECT().SendEmail(email, mock.Anything, mock.MatchedBy(func(body string) bool {
					return len(body) > 0
				})).Return(nil)
			},
			wantErr: false,
		},
		{
			name:  "User Not Found (Silent)",
			email: "ghost@mail.com",
			prepare: func(um *mocks.UserRepository, nm *mocks.NotificationService) {
				um.EXPECT().GetByEmail(mock.Anything, "ghost@mail.com").Return(nil, customErrors.ErrNotFound)
			},
			wantErr: false,
		},
		{
			name:  "Email Send Error",
			email: email,
			prepare: func(um *mocks.UserRepository, nm *mocks.NotificationService) {
				um.EXPECT().GetByEmail(mock.Anything, email).Return(&model.User{}, nil)
				nm.EXPECT().SendEmail(mock.Anything, mock.Anything, mock.Anything).Return(errors.New("smtp fail"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			um := mocks.NewUserRepository(t)
			nm := mocks.NewNotificationService(t)
			tc.prepare(um, nm)

			svc := service.NewUserService(um, nil, nm, nil, newDiscardLogger(), testSecret)
			err := svc.RecoverUsername(context.Background(), tc.email)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserService_ResetPassword(t *testing.T) {
	email := "test@mail.com"
	code := "123456"
	newPass := "new-password-123"
	userID := uuid.New()

	tests := []struct {
		name    string
		prepare func(um *mocks.UserRepository, om *mocks.OTPRepository, sm *mocks.RefreshSessionRepository)
		wantErr error
	}{
		{
			name: "Success",
			prepare: func(um *mocks.UserRepository, om *mocks.OTPRepository, sm *mocks.RefreshSessionRepository) {
				om.EXPECT().Verify(mock.Anything, email, code, 5).Return(true, nil)
				um.EXPECT().GetByEmail(mock.Anything, email).Return(&model.User{ID: userID}, nil)
				sm.EXPECT().RevokeAll(mock.Anything, userID).Return(nil)
				um.EXPECT().UpdatePasswordAndActivate(mock.Anything, email, mock.MatchedBy(func(hash string) bool {
					return hash != newPass && len(hash) > 0
				})).Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "OTP Not Found",
			prepare: func(um *mocks.UserRepository, om *mocks.OTPRepository, sm *mocks.RefreshSessionRepository) {
				om.EXPECT().Verify(mock.Anything, email, code, 5).Return(false, customErrors.ErrNotFound)
			},
			wantErr: customErrors.ErrInvalidInput,
		},
		{
			name: "Wrong OTP",
			prepare: func(um *mocks.UserRepository, om *mocks.OTPRepository, sm *mocks.RefreshSessionRepository) {
				om.EXPECT().Verify(mock.Anything, email, code, 5).Return(false, nil)
			},
			wantErr: customErrors.ErrInvalidInput,
		},
		{
			name: "Update Repo Error",
			prepare: func(um *mocks.UserRepository, om *mocks.OTPRepository, sm *mocks.RefreshSessionRepository) {
				om.EXPECT().Verify(mock.Anything, email, code, 5).Return(true, nil)
				um.EXPECT().GetByEmail(mock.Anything, email).Return(&model.User{ID: userID}, nil)
				sm.EXPECT().RevokeAll(mock.Anything, userID).Return(nil)
				um.EXPECT().UpdatePasswordAndActivate(mock.Anything, email, mock.Anything).Return(errors.New("db fail"))
			},
			wantErr: errors.New("db fail"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			um := mocks.NewUserRepository(t)
			om := mocks.NewOTPRepository(t)
			sm := mocks.NewRefreshSessionRepository(t)
			tc.prepare(um, om, sm)

			svc := service.NewUserService(um, om, nil, sm, newDiscardLogger(), testSecret)
			err := svc.ResetPassword(context.Background(), email, code, newPass)

			if tc.wantErr != nil {
				assert.Error(t, err)
				if errors.Is(tc.wantErr, customErrors.ErrInvalidInput) {
					assert.ErrorIs(t, err, tc.wantErr)
				} else {
					assert.Contains(t, err.Error(), tc.wantErr.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserService_RefreshToken(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name    string
		token   string
		prepare func(um *mocks.UserRepository, sm *mocks.RefreshSessionRepository)
		wantErr error
	}{
		{
			name:  "Success",
			token: "opaque-refresh-token",
			prepare: func(um *mocks.UserRepository, sm *mocks.RefreshSessionRepository) {
				sm.EXPECT().Rotate(mock.Anything, "opaque-refresh-token", mock.Anything, 7*24*time.Hour).Return(userID, nil)
				um.EXPECT().GetByID(mock.Anything, userID).Return(&model.User{ID: userID, IsActive: true}, nil)
			},
			wantErr: nil,
		},
		{
			name:  "Invalid Session",
			token: "invalid",
			prepare: func(um *mocks.UserRepository, sm *mocks.RefreshSessionRepository) {
				sm.EXPECT().Rotate(mock.Anything, "invalid", mock.Anything, 7*24*time.Hour).Return(uuid.Nil, customErrors.ErrUnauthorized)
			},
			wantErr: customErrors.ErrUnauthorized,
		},
		{
			name:  "User Not Found",
			token: "missing-user",
			prepare: func(um *mocks.UserRepository, sm *mocks.RefreshSessionRepository) {
				sm.EXPECT().Rotate(mock.Anything, "missing-user", mock.Anything, 7*24*time.Hour).Return(userID, nil)
				um.EXPECT().GetByID(mock.Anything, userID).Return(nil, errors.New("nf"))
				sm.EXPECT().Revoke(mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: customErrors.ErrUnauthorized,
		},
		{
			name:  "User Inactive",
			token: "inactive-user",
			prepare: func(um *mocks.UserRepository, sm *mocks.RefreshSessionRepository) {
				sm.EXPECT().Rotate(mock.Anything, "inactive-user", mock.Anything, 7*24*time.Hour).Return(userID, nil)
				um.EXPECT().GetByID(mock.Anything, userID).Return(&model.User{ID: userID, IsActive: false}, nil)
				sm.EXPECT().Revoke(mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: customErrors.ErrUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			um := mocks.NewUserRepository(t)
			sm := mocks.NewRefreshSessionRepository(t)
			if tc.prepare != nil {
				tc.prepare(um, sm)
			}

			svc := service.NewUserService(um, nil, nil, sm, newDiscardLogger(), testSecret)
			newToken, newRefresh, err := svc.RefreshToken(context.Background(), tc.token)

			if tc.wantErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tc.wantErr)
				assert.Empty(t, newToken)
				assert.Empty(t, newRefresh)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, newToken)
				assert.NotEmpty(t, newRefresh)
			}
		})
	}
}

func TestUserService_GetList(t *testing.T) {
	tests := []struct {
		name    string
		page    int
		prepare func(um *mocks.UserRepository)
		wantErr bool
	}{
		{
			name: "Success",
			page: 1,
			prepare: func(um *mocks.UserRepository) {
				um.EXPECT().GetList(mock.Anything, 10, 0).Return([]*model.User{}, 0, nil)
			},
			wantErr: false,
		},
		{
			name: "Repo Error",
			page: 1,
			prepare: func(um *mocks.UserRepository) {
				um.EXPECT().GetList(mock.Anything, 10, 0).Return(nil, 0, errors.New("fail"))
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			um := mocks.NewUserRepository(t)
			tc.prepare(um)

			svc := service.NewUserService(um, nil, nil, nil, newDiscardLogger(), testSecret)
			_, _, err := svc.GetList(context.Background(), tc.page, 10)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserService_Delete(t *testing.T) {
	id := uuid.New()
	um := mocks.NewUserRepository(t)
	um.EXPECT().Delete(mock.Anything, id).Return(nil)

	svc := service.NewUserService(um, nil, nil, nil, newDiscardLogger(), testSecret)
	err := svc.Delete(context.Background(), id)
	assert.NoError(t, err)
}

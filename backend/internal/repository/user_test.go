package repository_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	customErrors "warehouse-management-system/internal/errors"
	"warehouse-management-system/internal/model"
	"warehouse-management-system/internal/repository"
)

func TestUserRepository_CreateUser(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	repo := repository.NewUserRepository(db, newDiscardLogger())

	defer func() {
		_, _ = db.Exec("DELETE FROM users")
	}()

	id := uuid.New()

	type args struct {
		user *model.User
	}

	tests := []struct {
		name      string
		args      args
		prepare   func()
		wantError error
	}{
		{
			name: "Success",
			args: args{
				user: &model.User{
					ID:           id,
					Username:     "testuser",
					Email:        "test@mail.com",
					FullName:     "Test User",
					Role:         model.RoleWorker,
					IsActive:     true,
					PasswordHash: "hash123",
				},
			},
			wantError: nil,
		},
		{
			name: "Duplicate Username",
			args: args{
				user: &model.User{
					ID:           uuid.New(),
					Username:     "dupuser",
					Email:        "unique@mail.com",
					FullName:     "Dup User",
					Role:         model.RoleAdmin,
					PasswordHash: "hash",
				},
			},
			prepare: func() {
				_, err := db.Exec("INSERT INTO users (id, username, email, full_name, role, password_hash) VALUES ($1, $2, $3, $4, $5, $6)",
					uuid.New(), "dupuser", "other@mail.com", "Orig", "admin", "hash")
				require.NoError(t, err)
			},
			wantError: customErrors.ErrAlreadyExists,
		},
		{
			name: "Duplicate Email",
			args: args{
				user: &model.User{
					ID:           uuid.New(),
					Username:     "uniqueuser",
					Email:        "dup@mail.com",
					FullName:     "Dup Email",
					Role:         model.RoleAdmin,
					PasswordHash: "hash",
				},
			},
			prepare: func() {
				_, err := db.Exec("INSERT INTO users (id, username, email, full_name, role, password_hash) VALUES ($1, $2, $3, $4, $5, $6)",
					uuid.New(), "origuser", "dup@mail.com", "Orig", "admin", "hash")
				require.NoError(t, err)
			},
			wantError: customErrors.ErrAlreadyExists,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.prepare != nil {
				tc.prepare()
			}

			err := repo.CreateUser(context.Background(), tc.args.user)

			if tc.wantError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				var count int
				err := db.QueryRow("SELECT COUNT(*) FROM users WHERE id = $1", tc.args.user.ID).Scan(&count)
				assert.NoError(t, err)
				assert.Equal(t, 1, count)
			}
		})
	}
}

func TestUserRepository_GetByUsername(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	repo := repository.NewUserRepository(db, newDiscardLogger())

	defer func() {
		_, _ = db.Exec("DELETE FROM users")
	}()

	id := uuid.New()
	_, err := db.Exec("INSERT INTO users (id, username, email, full_name, role, password_hash, is_active) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		id, "findme", "find@mail.com", "Find Me", "worker", "hash", true)
	require.NoError(t, err)

	tests := []struct {
		name      string
		username  string
		wantError error
		wantID    uuid.UUID
	}{
		{
			name:      "Success",
			username:  "findme",
			wantError: nil,
			wantID:    id,
		},
		{
			name:      "Not Found",
			username:  "ghost",
			wantError: customErrors.ErrNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := repo.GetByUsername(context.Background(), tc.username)

			if tc.wantError != nil {
				assert.ErrorIs(t, err, tc.wantError)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tc.wantID, got.ID)
				assert.Equal(t, tc.username, got.Username)
			}
		})
	}
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	repo := repository.NewUserRepository(db, newDiscardLogger())

	defer func() {
		_, _ = db.Exec("DELETE FROM users")
	}()

	id := uuid.New()
	email := "target@mail.com"
	_, err := db.Exec("INSERT INTO users (id, username, email, full_name, role, password_hash, is_active) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		id, "target", email, "Target", "admin", "hash", true)
	require.NoError(t, err)

	tests := []struct {
		name      string
		email     string
		wantError error
		wantID    uuid.UUID
	}{
		{
			name:      "Success",
			email:     email,
			wantError: nil,
			wantID:    id,
		},
		{
			name:      "Not Found",
			email:     "missing@mail.com",
			wantError: customErrors.ErrNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := repo.GetByEmail(context.Background(), tc.email)

			if tc.wantError != nil {
				assert.ErrorIs(t, err, tc.wantError)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tc.wantID, got.ID)
				assert.Equal(t, tc.email, got.Email)
			}
		})
	}
}

func TestUserRepository_GetByID(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	repo := repository.NewUserRepository(db, newDiscardLogger())

	defer func() {
		_, _ = db.Exec("DELETE FROM users")
	}()

	id := uuid.New()
	_, err := db.Exec("INSERT INTO users (id, username, email, full_name, role, password_hash, is_active) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		id, "byid", "byid@mail.com", "By ID", "worker", "hash", true)
	require.NoError(t, err)

	tests := []struct {
		name      string
		id        uuid.UUID
		wantError error
	}{
		{
			name:      "Success",
			id:        id,
			wantError: nil,
		},
		{
			name:      "Not Found",
			id:        uuid.New(),
			wantError: customErrors.ErrNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := repo.GetByID(context.Background(), tc.id)

			if tc.wantError != nil {
				assert.ErrorIs(t, err, tc.wantError)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tc.id, got.ID)
			}
		})
	}
}

func TestUserRepository_Update(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	repo := repository.NewUserRepository(db, newDiscardLogger())

	defer func() {
		_, _ = db.Exec("DELETE FROM users")
	}()

	id := uuid.New()
	_, err := db.Exec("INSERT INTO users (id, username, email, full_name, role, password_hash) VALUES ($1, $2, $3, $4, $5, $6)",
		id, "update_me", "old@mail.com", "Old Name", "worker", "hash")
	require.NoError(t, err)

	collisionID := uuid.New()
	_, err = db.Exec("INSERT INTO users (id, username, email, full_name, role, password_hash) VALUES ($1, $2, $3, $4, $5, $6)",
		collisionID, "collision", "collision@mail.com", "Col", "admin", "hash")
	require.NoError(t, err)

	tests := []struct {
		name      string
		input     *model.User
		wantError error
		checkRes  func(*testing.T)
	}{
		{
			name: "Success",
			input: &model.User{
				ID:       id,
				FullName: "New Name",
				Email:    "new@mail.com",
				Role:     model.RoleAdmin,
			},
			wantError: nil,
			checkRes: func(t *testing.T) {
				var name, email, role string
				err := db.QueryRow("SELECT full_name, email, role FROM users WHERE id = $1", id).Scan(&name, &email, &role)
				assert.NoError(t, err)
				assert.Equal(t, "New Name", name)
				assert.Equal(t, "new@mail.com", email)
				assert.Equal(t, "admin", role)
			},
		},
		{
			name: "Not Found",
			input: &model.User{
				ID:       uuid.New(),
				FullName: "Ghost",
				Email:    "ghost@mail.com",
				Role:     model.RoleWorker,
			},
			wantError: customErrors.ErrNotFound,
		},
		{
			name: "Duplicate Email",
			input: &model.User{
				ID:       id,
				FullName: "Try Dup",
				Email:    "collision@mail.com",
				Role:     model.RoleWorker,
			},
			wantError: customErrors.ErrAlreadyExists,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := repo.Update(context.Background(), tc.input)

			if tc.wantError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tc.checkRes != nil {
					tc.checkRes(t)
				}
			}
		})
	}
}

func TestUserRepository_Delete(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	repo := repository.NewUserRepository(db, newDiscardLogger())

	defer func() {
		_, _ = db.Exec("DELETE FROM users")
	}()

	id := uuid.New()
	_, err := db.Exec("INSERT INTO users (id, username, email, full_name, role, password_hash) VALUES ($1, $2, $3, $4, $5, $6)",
		id, "del_user", "del@mail.com", "Del", "worker", "hash")
	require.NoError(t, err)

	tests := []struct {
		name      string
		id        uuid.UUID
		wantError error
	}{
		{
			name:      "Success",
			id:        id,
			wantError: nil,
		},
		{
			name:      "Not Found",
			id:        uuid.New(),
			wantError: customErrors.ErrNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := repo.Delete(context.Background(), tc.id)

			if tc.wantError != nil {
				assert.ErrorIs(t, err, tc.wantError)
			} else {
				assert.NoError(t, err)
				var count int
				_ = db.QueryRow("SELECT COUNT(*) FROM users WHERE id = $1", tc.id).Scan(&count)
				assert.Equal(t, 0, count)
			}
		})
	}
}

func TestUserRepository_UpdatePasswordAndActivate(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	repo := repository.NewUserRepository(db, newDiscardLogger())

	defer func() {
		_, _ = db.Exec("DELETE FROM users")
	}()

	email := "pass@mail.com"
	_, err := db.Exec("INSERT INTO users (id, username, email, full_name, role, password_hash, is_active) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		uuid.New(), "passuser", email, "Pass", "worker", "oldhash", false)
	require.NoError(t, err)

	tests := []struct {
		name      string
		email     string
		newHash   string
		wantError error
		checkRes  func(*testing.T)
	}{
		{
			name:      "Success",
			email:     email,
			newHash:   "newsecrethash",
			wantError: nil,
			checkRes: func(t *testing.T) {
				var hash string
				var active bool
				err := db.QueryRow("SELECT password_hash, is_active FROM users WHERE email = $1", email).Scan(&hash, &active)
				assert.NoError(t, err)
				assert.Equal(t, "newsecrethash", hash)
				assert.True(t, active)
			},
		},
		{
			name:      "Not Found",
			email:     "unknown@mail.com",
			newHash:   "hash",
			wantError: customErrors.ErrNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := repo.UpdatePasswordAndActivate(context.Background(), tc.email, tc.newHash)

			if tc.wantError != nil {
				assert.ErrorIs(t, err, tc.wantError)
			} else {
				assert.NoError(t, err)
				if tc.checkRes != nil {
					tc.checkRes(t)
				}
			}
		})
	}
}

func TestUserRepository_GetList(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	repo := repository.NewUserRepository(db, newDiscardLogger())

	defer func() {
		_, _ = db.Exec("DELETE FROM users")
	}()

	names := []string{"Alice", "Bob", "Charlie"}
	for _, name := range names {
		_, err := db.Exec("INSERT INTO users (id, username, email, full_name, role, password_hash) VALUES ($1, $2, $3, $4, $5, $6)",
			uuid.New(), name, name+"@mail.com", name, "worker", "hash")
		require.NoError(t, err)
	}

	tests := []struct {
		name      string
		limit     int
		offset    int
		wantCount int
		wantTotal int
		checkRes  func(*testing.T, []*model.User)
	}{
		{
			name:      "All Users",
			limit:     10,
			offset:    0,
			wantCount: 3,
			wantTotal: 3,
			checkRes: func(t *testing.T, users []*model.User) {
				assert.Equal(t, "Charlie", users[0].FullName)
				assert.Equal(t, "Bob", users[1].FullName)
				assert.Equal(t, "Alice", users[2].FullName)
			},
		},
		{
			name:      "Pagination",
			limit:     1,
			offset:    0,
			wantCount: 1,
			wantTotal: 3,
			checkRes: func(t *testing.T, users []*model.User) {
				assert.Equal(t, "Charlie", users[0].FullName)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			list, total, err := repo.GetList(context.Background(), tc.limit, tc.offset)
			assert.NoError(t, err)
			assert.Len(t, list, tc.wantCount)
			assert.Equal(t, tc.wantTotal, total)
			if tc.checkRes != nil {
				tc.checkRes(t, list)
			}
		})
	}
}

package repository_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	customErrors "warehouse-management-system/internal/errors"
	"warehouse-management-system/internal/model"
	"warehouse-management-system/internal/repository"
)

func TestCounterpartyRepository_Create(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	repo := repository.NewCounterpartyRepository(db, newDiscardLogger())

	defer func() {
		_, _ = db.Exec("DELETE FROM counterparties")
	}()

	id := uuid.New()

	type args struct {
		counterparty *model.Counterparty
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
				counterparty: &model.Counterparty{
					ID:          uuid.New(),
					Name:        "Test Client",
					Type:        model.CounterpartyClient,
					PhoneNumber: "1234567890",
					Email:       "test@client.com",
				},
			},
			wantError: nil,
		},
		{
			name: "Duplicate ID",
			args: args{
				counterparty: &model.Counterparty{
					ID:          id,
					Name:        "Duplicate Client",
					Type:        model.CounterpartyClient,
					PhoneNumber: "0000000000",
					Email:       "dup@client.com",
				},
			},
			prepare: func() {
				_, err := db.Exec("INSERT INTO counterparties (id, name, type, phone_number, email) VALUES ($1, $2, $3, $4, $5)",
					id, "Original", "client", "111", "orig@mail.com")
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

			err := repo.Create(context.Background(), tc.args.counterparty)

			if tc.wantError != nil {
				assert.ErrorIs(t, err, tc.wantError)
			} else {
				assert.NoError(t, err)
				var count int
				err := db.QueryRow("SELECT COUNT(*) FROM counterparties WHERE id = $1", tc.args.counterparty.ID).Scan(&count)
				assert.NoError(t, err)
				assert.Equal(t, 1, count)
			}
		})
	}
}

func TestCounterpartyRepository_GetByID(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	repo := repository.NewCounterpartyRepository(db, newDiscardLogger())

	defer func() {
		_, _ = db.Exec("DELETE FROM counterparties")
	}()

	existingID := uuid.New()
	_, err := db.Exec("INSERT INTO counterparties (id, name, type, phone_number, email) VALUES ($1, $2, $3, $4, $5)",
		existingID, "Target Client", "client", "555-55-55", "target@mail.com")
	require.NoError(t, err)

	tests := []struct {
		name      string
		id        uuid.UUID
		wantError error
		wantName  string
	}{
		{
			name:      "Success",
			id:        existingID,
			wantError: nil,
			wantName:  "Target Client",
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
				assert.Equal(t, tc.wantName, got.Name)
				assert.Equal(t, tc.id, got.ID)
			}
		})
	}
}

func TestCounterpartyRepository_Update(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	repo := repository.NewCounterpartyRepository(db, newDiscardLogger())

	defer func() {
		_, _ = db.Exec("DELETE FROM counterparties")
	}()

	id := uuid.New()
	_, err := db.Exec("INSERT INTO counterparties (id, name, type, phone_number, email) VALUES ($1, $2, $3, $4, $5)",
		id, "Old Name", "client", "111", "old@mail.com")
	require.NoError(t, err)

	tests := []struct {
		name      string
		input     *model.Counterparty
		wantError error
		checkRes  func(*testing.T)
	}{
		{
			name: "Success",
			input: &model.Counterparty{
				ID:          id,
				Name:        "New Name",
				PhoneNumber: "999",
				Email:       "new@mail.com",
			},
			wantError: nil,
			checkRes: func(t *testing.T) {
				var name, phone string
				err := db.QueryRow("SELECT name, phone_number FROM counterparties WHERE id = $1", id).Scan(&name, &phone)
				assert.NoError(t, err)
				assert.Equal(t, "New Name", name)
				assert.Equal(t, "999", phone)
			},
		},
		{
			name: "Not Found",
			input: &model.Counterparty{
				ID:          uuid.New(),
				Name:        "Ghost",
				PhoneNumber: "000",
				Email:       "ghost@mail.com",
			},
			wantError: customErrors.ErrNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := repo.Update(context.Background(), tc.input)

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

func TestCounterpartyRepository_Delete(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	repo := repository.NewCounterpartyRepository(db, newDiscardLogger())

	defer func() {
		_, _ = db.Exec("DELETE FROM counterparties")
	}()

	id := uuid.New()
	_, err := db.Exec("INSERT INTO counterparties (id, name, type, phone_number, email) VALUES ($1, $2, $3, $4, $5)",
		id, "To Delete", "supplier", "123", "del@mail.com")
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
				_ = db.QueryRow("SELECT COUNT(*) FROM counterparties WHERE id = $1", tc.id).Scan(&count)
				assert.Equal(t, 0, count)
			}
		})
	}
}

func TestCounterpartyRepository_GetList(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	repo := repository.NewCounterpartyRepository(db, newDiscardLogger())

	defer func() {
		_, _ = db.Exec("DELETE FROM counterparties")
	}()

	for i := 0; i < 5; i++ {
		name := fmt.Sprintf("Client %d", i)
		_, err := db.Exec("INSERT INTO counterparties (id, name, type, phone_number, email) VALUES ($1, $2, $3, $4, $5)",
			uuid.New(), name, "client", "000", "c@mail.com")
		require.NoError(t, err)
	}
	for i := 0; i < 5; i++ {
		name := fmt.Sprintf("Supplier %d", i)
		_, err := db.Exec("INSERT INTO counterparties (id, name, type, phone_number, email) VALUES ($1, $2, $3, $4, $5)",
			uuid.New(), name, "supplier", "111", "s@mail.com")
		require.NoError(t, err)
	}

	tests := []struct {
		name       string
		limit      int
		offset     int
		typeFilter string
		wantCount  int
		wantTotal  int
	}{
		{
			name:       "All Page 1",
			limit:      10,
			offset:     0,
			typeFilter: "",
			wantCount:  10,
			wantTotal:  10,
		},
		{
			name:       "Filter Clients",
			limit:      10,
			offset:     0,
			typeFilter: "client",
			wantCount:  5,
			wantTotal:  5,
		},
		{
			name:       "Filter Suppliers",
			limit:      10,
			offset:     0,
			typeFilter: "supplier",
			wantCount:  5,
			wantTotal:  5,
		},
		{
			name:       "Pagination",
			limit:      2,
			offset:     0,
			typeFilter: "",
			wantCount:  2,
			wantTotal:  10,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			list, total, err := repo.GetList(context.Background(), tc.limit, tc.offset, tc.typeFilter)
			assert.NoError(t, err)
			assert.Len(t, list, tc.wantCount)
			assert.Equal(t, tc.wantTotal, total)

			if tc.typeFilter != "" {
				for _, item := range list {
					assert.Equal(t, model.CounterpartyType(tc.typeFilter), item.Type)
				}
			}
		})
	}
}

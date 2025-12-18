package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	customErrors "warehouse-management-system/internal/errors"
	"warehouse-management-system/internal/model"
	"warehouse-management-system/internal/service"
	"warehouse-management-system/mocks"
)

func TestCounterpartyService_Create(t *testing.T) {
	type args struct {
		name    string
		typeStr string
		phone   string
		email   string
	}
	tests := []struct {
		name      string
		args      args
		prepare   func(m *mocks.CounterpartyRepository)
		wantError error
		checkRes  func(*testing.T, *model.Counterparty)
	}{
		{
			name: "Success Client",
			args: args{name: "Client A", typeStr: "client", phone: "123", email: "a@a.com"},
			prepare: func(m *mocks.CounterpartyRepository) {
				m.EXPECT().Create(mock.Anything, mock.MatchedBy(func(c *model.Counterparty) bool {
					return c.Name == "Client A" && c.Type == model.CounterpartyClient && c.ID != uuid.Nil
				})).Return(nil)
			},
			wantError: nil,
			checkRes: func(t *testing.T, c *model.Counterparty) {
				assert.NotNil(t, c)
				assert.Equal(t, model.CounterpartyClient, c.Type)
			},
		},
		{
			name: "Success Supplier",
			args: args{name: "Supplier B", typeStr: "supplier", phone: "456", email: "b@b.com"},
			prepare: func(m *mocks.CounterpartyRepository) {
				m.EXPECT().Create(mock.Anything, mock.MatchedBy(func(c *model.Counterparty) bool {
					return c.Type == model.CounterpartySupplier
				})).Return(nil)
			},
			wantError: nil,
			checkRes: func(t *testing.T, c *model.Counterparty) {
				assert.Equal(t, model.CounterpartySupplier, c.Type)
			},
		},
		{
			name: "Empty Name",
			args: args{name: "", typeStr: "client"},
			prepare: func(m *mocks.CounterpartyRepository) {
			},
			wantError: customErrors.ErrInvalidInput,
		},
		{
			name: "Invalid Type",
			args: args{name: "Valid", typeStr: "alien"},
			prepare: func(m *mocks.CounterpartyRepository) {
			},
			wantError: customErrors.ErrInvalidInput,
		},
		{
			name: "Repo Error",
			args: args{name: "Valid", typeStr: "client"},
			prepare: func(m *mocks.CounterpartyRepository) {
				m.EXPECT().Create(mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			wantError: errors.New("db error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := mocks.NewCounterpartyRepository(t)
			if tc.prepare != nil {
				tc.prepare(repo)
			}

			svc := service.NewCounterpartyService(repo, newDiscardLogger())
			got, err := svc.Create(context.Background(), tc.args.name, tc.args.typeStr, tc.args.phone, tc.args.email)

			if tc.wantError != nil {
				assert.Error(t, err)
				if errors.Is(tc.wantError, customErrors.ErrInvalidInput) {
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

func TestCounterpartyService_GetByID(t *testing.T) {
	id := uuid.New()
	expected := &model.Counterparty{ID: id, Name: "Test"}

	repo := mocks.NewCounterpartyRepository(t)
	repo.EXPECT().GetByID(mock.Anything, id).Return(expected, nil)

	svc := service.NewCounterpartyService(repo, newDiscardLogger())
	got, err := svc.GetByID(context.Background(), id)

	assert.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestCounterpartyService_Update(t *testing.T) {
	id := uuid.New()
	type args struct {
		id    uuid.UUID
		name  string
		phone string
		email string
	}
	tests := []struct {
		name      string
		args      args
		prepare   func(m *mocks.CounterpartyRepository)
		wantError error
	}{
		{
			name: "Success",
			args: args{id: id, name: "New Name", phone: "999", email: "new@mail.com"},
			prepare: func(m *mocks.CounterpartyRepository) {
				existing := &model.Counterparty{ID: id, Name: "Old Name"}
				m.EXPECT().GetByID(mock.Anything, id).Return(existing, nil)
				m.EXPECT().Update(mock.Anything, mock.MatchedBy(func(c *model.Counterparty) bool {
					return c.ID == id && c.Name == "New Name" && c.PhoneNumber == "999" && c.Email == "new@mail.com"
				})).Return(nil)
			},
			wantError: nil,
		},
		{
			name: "Empty Name",
			args: args{id: id, name: ""},
			prepare: func(m *mocks.CounterpartyRepository) {
			},
			wantError: customErrors.ErrInvalidInput,
		},
		{
			name: "GetByID Error",
			args: args{id: id, name: "Valid"},
			prepare: func(m *mocks.CounterpartyRepository) {
				m.EXPECT().GetByID(mock.Anything, id).Return(nil, errors.New("not found"))
			},
			wantError: errors.New("not found"),
		},
		{
			name: "Update Repo Error",
			args: args{id: id, name: "Valid"},
			prepare: func(m *mocks.CounterpartyRepository) {
				existing := &model.Counterparty{ID: id}
				m.EXPECT().GetByID(mock.Anything, id).Return(existing, nil)
				m.EXPECT().Update(mock.Anything, existing).Return(errors.New("db fail"))
			},
			wantError: errors.New("db fail"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := mocks.NewCounterpartyRepository(t)
			if tc.prepare != nil {
				tc.prepare(repo)
			}

			svc := service.NewCounterpartyService(repo, newDiscardLogger())
			got, err := svc.Update(context.Background(), tc.args.id, tc.args.name, tc.args.phone, tc.args.email)

			if tc.wantError != nil {
				assert.Error(t, err)
				if errors.Is(tc.wantError, customErrors.ErrInvalidInput) {
					assert.ErrorIs(t, err, tc.wantError)
				} else {
					assert.Contains(t, err.Error(), tc.wantError.Error())
				}
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tc.args.name, got.Name)
			}
		})
	}
}

func TestCounterpartyService_Delete(t *testing.T) {
	id := uuid.New()
	repo := mocks.NewCounterpartyRepository(t)
	repo.EXPECT().Delete(mock.Anything, id).Return(nil)

	svc := service.NewCounterpartyService(repo, newDiscardLogger())
	err := svc.Delete(context.Background(), id)

	assert.NoError(t, err)
}

func TestCounterpartyService_GetList(t *testing.T) {
	type args struct {
		page       int
		pageSize   int
		typeFilter string
	}
	expectedList := []*model.Counterparty{{ID: uuid.New(), Name: "C1"}}

	tests := []struct {
		name      string
		args      args
		prepare   func(m *mocks.CounterpartyRepository)
		wantList  []*model.Counterparty
		wantCount int
		wantErr   bool
	}{
		{
			name: "Success Defaults",
			args: args{page: 0, pageSize: 0, typeFilter: ""},
			prepare: func(m *mocks.CounterpartyRepository) {
				m.EXPECT().GetList(mock.Anything, 10, 0, "").Return(expectedList, 1, nil)
			},
			wantList:  expectedList,
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "Success Filtered",
			args: args{page: 2, pageSize: 5, typeFilter: "client"},
			prepare: func(m *mocks.CounterpartyRepository) {
				m.EXPECT().GetList(mock.Anything, 5, 5, "client").Return(expectedList, 5, nil)
			},
			wantList:  expectedList,
			wantCount: 5,
			wantErr:   false,
		},
		{
			name: "Repo Error",
			args: args{page: 1, pageSize: 10},
			prepare: func(m *mocks.CounterpartyRepository) {
				m.EXPECT().GetList(mock.Anything, 10, 0, "").Return(nil, 0, errors.New("fail"))
			},
			wantList:  nil,
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := mocks.NewCounterpartyRepository(t)
			tc.prepare(repo)

			svc := service.NewCounterpartyService(repo, newDiscardLogger())
			got, count, err := svc.GetList(context.Background(), tc.args.page, tc.args.pageSize, tc.args.typeFilter)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantList, got)
				assert.Equal(t, tc.wantCount, count)
			}
		})
	}
}

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

func TestOrderService_Create(t *testing.T) {
	cpID := uuid.New()
	type args struct {
		order *model.Order
	}
	tests := []struct {
		name      string
		args      args
		prepare   func(rm *mocks.OrderRepository, cm *mocks.OrderCounterpartyService, pm *mocks.OrderProductService)
		wantError error
		checkRes  func(*testing.T, *model.Order)
	}{
		{
			name: "Success Inbound",
			args: args{
				order: &model.Order{
					CounterpartyID: cpID,
					OrderType:      model.OrderInbound,
					Items:          []model.OrderItem{{ProductID: uuid.New(), Quantity: 10}},
				},
			},
			prepare: func(rm *mocks.OrderRepository, cm *mocks.OrderCounterpartyService, pm *mocks.OrderProductService) {
				cm.EXPECT().GetByID(mock.Anything, cpID).Return(&model.Counterparty{ID: cpID, Type: model.CounterpartySupplier}, nil)
				rm.EXPECT().Create(mock.Anything, mock.MatchedBy(func(o *model.Order) bool {
					return o.Status == model.StatusPending && o.ID != uuid.Nil
				})).Return(nil)
			},
			wantError: nil,
			checkRes: func(t *testing.T, o *model.Order) {
				assert.NotNil(t, o)
				assert.Equal(t, model.OrderInbound, o.OrderType)
			},
		},
		{
			name: "Success Outbound",
			args: args{
				order: &model.Order{
					CounterpartyID: cpID,
					OrderType:      model.OrderOutbound,
					Destination:    "Some Address",
					Items:          []model.OrderItem{{ProductID: uuid.New(), Quantity: 5}},
				},
			},
			prepare: func(rm *mocks.OrderRepository, cm *mocks.OrderCounterpartyService, pm *mocks.OrderProductService) {
				cm.EXPECT().GetByID(mock.Anything, cpID).Return(&model.Counterparty{ID: cpID, Type: model.CounterpartyClient}, nil)
				rm.EXPECT().Create(mock.Anything, mock.MatchedBy(func(o *model.Order) bool {
					return o.OrderType == model.OrderOutbound
				})).Return(nil)
			},
			wantError: nil,
		},
		{
			name: "Counterparty Service Error",
			args: args{order: &model.Order{CounterpartyID: cpID}},
			prepare: func(rm *mocks.OrderRepository, cm *mocks.OrderCounterpartyService, pm *mocks.OrderProductService) {
				cm.EXPECT().GetByID(mock.Anything, cpID).Return(nil, errors.New("cp error"))
			},
			wantError: errors.New("cp error"),
		},
		{
			name: "Inbound Wrong CP Type",
			args: args{order: &model.Order{CounterpartyID: cpID, OrderType: model.OrderInbound}},
			prepare: func(rm *mocks.OrderRepository, cm *mocks.OrderCounterpartyService, pm *mocks.OrderProductService) {
				cm.EXPECT().GetByID(mock.Anything, cpID).Return(&model.Counterparty{ID: cpID, Type: model.CounterpartyClient}, nil)
			},
			wantError: customErrors.ErrInvalidInput,
		},
		{
			name: "Outbound Wrong CP Type",
			args: args{order: &model.Order{CounterpartyID: cpID, OrderType: model.OrderOutbound, Destination: "Addr"}},
			prepare: func(rm *mocks.OrderRepository, cm *mocks.OrderCounterpartyService, pm *mocks.OrderProductService) {
				cm.EXPECT().GetByID(mock.Anything, cpID).Return(&model.Counterparty{ID: cpID, Type: model.CounterpartySupplier}, nil)
			},
			wantError: customErrors.ErrInvalidInput,
		},
		{
			name: "Outbound Missing Destination",
			args: args{order: &model.Order{CounterpartyID: cpID, OrderType: model.OrderOutbound, Destination: ""}},
			prepare: func(rm *mocks.OrderRepository, cm *mocks.OrderCounterpartyService, pm *mocks.OrderProductService) {
				cm.EXPECT().GetByID(mock.Anything, cpID).Return(&model.Counterparty{ID: cpID, Type: model.CounterpartyClient}, nil)
			},
			wantError: customErrors.ErrInvalidInput,
		},
		{
			name: "Repo Create Error",
			args: args{
				order: &model.Order{
					CounterpartyID: cpID,
					OrderType:      model.OrderInbound,
					Items:          []model.OrderItem{},
				},
			},
			prepare: func(rm *mocks.OrderRepository, cm *mocks.OrderCounterpartyService, pm *mocks.OrderProductService) {
				cm.EXPECT().GetByID(mock.Anything, cpID).Return(&model.Counterparty{ID: cpID, Type: model.CounterpartySupplier}, nil)
				rm.EXPECT().Create(mock.Anything, mock.Anything).Return(errors.New("db fail"))
			},
			wantError: errors.New("db fail"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rm := mocks.NewOrderRepository(t)
			cm := mocks.NewOrderCounterpartyService(t)
			pm := mocks.NewOrderProductService(t)

			if tc.prepare != nil {
				tc.prepare(rm, cm, pm)
			}

			svc := service.NewOrderService(rm, cm, pm, newDiscardLogger())
			got, err := svc.Create(context.Background(), tc.args.order)

			if tc.wantError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantError.Error())
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

func TestOrderService_GetByID(t *testing.T) {
	id := uuid.New()
	expectedOrder := &model.Order{ID: id, Status: model.StatusPending}

	rm := mocks.NewOrderRepository(t)
	rm.EXPECT().GetByID(mock.Anything, id).Return(expectedOrder, nil)

	svc := service.NewOrderService(rm, nil, nil, newDiscardLogger())
	got, err := svc.GetByID(context.Background(), id)

	assert.NoError(t, err)
	assert.Equal(t, expectedOrder, got)
}

func TestOrderService_Update(t *testing.T) {
	id := uuid.New()
	prodID1 := uuid.New()
	prodID2 := uuid.New()

	type args struct {
		id        uuid.UUID
		newStatus string
	}
	tests := []struct {
		name      string
		args      args
		prepare   func(rm *mocks.OrderRepository, pm *mocks.OrderProductService)
		wantError error
	}{
		{
			name: "Success Status Change Pending to Processing",
			args: args{id: id, newStatus: string(model.StatusProcessing)},
			prepare: func(rm *mocks.OrderRepository, pm *mocks.OrderProductService) {
				order := &model.Order{ID: id, Status: model.StatusPending}
				rm.EXPECT().GetByID(mock.Anything, id).Return(order, nil)
				rm.EXPECT().Update(mock.Anything, mock.MatchedBy(func(o *model.Order) bool {
					return o.Status == model.StatusProcessing
				})).Return(nil)
			},
			wantError: nil,
		},
		{
			name: "Success Complete Inbound (Income)",
			args: args{id: id, newStatus: string(model.StatusCompleted)},
			prepare: func(rm *mocks.OrderRepository, pm *mocks.OrderProductService) {
				order := &model.Order{
					ID:        id,
					Status:    model.StatusProcessing,
					OrderType: model.OrderInbound,
					Items:     []model.OrderItem{{ProductID: prodID1, Quantity: 10}},
				}
				rm.EXPECT().GetByID(mock.Anything, id).Return(order, nil)
				pm.EXPECT().ChangeStock(mock.Anything, prodID1, 10, "income").Return(nil)
				rm.EXPECT().Update(mock.Anything, mock.MatchedBy(func(o *model.Order) bool {
					return o.Status == model.StatusCompleted
				})).Return(nil)
			},
			wantError: nil,
		},
		{
			name: "Success Complete Outbound (Expense)",
			args: args{id: id, newStatus: string(model.StatusCompleted)},
			prepare: func(rm *mocks.OrderRepository, pm *mocks.OrderProductService) {
				order := &model.Order{
					ID:        id,
					Status:    model.StatusProcessing,
					OrderType: model.OrderOutbound,
					Items:     []model.OrderItem{{ProductID: prodID1, Quantity: 5}},
				}
				rm.EXPECT().GetByID(mock.Anything, id).Return(order, nil)
				pm.EXPECT().ChangeStock(mock.Anything, prodID1, 5, "expense").Return(nil)
				rm.EXPECT().Update(mock.Anything, mock.MatchedBy(func(o *model.Order) bool {
					return o.Status == model.StatusCompleted
				})).Return(nil)
			},
			wantError: nil,
		},
		{
			name: "Cannot Update Canceled Order",
			args: args{id: id, newStatus: string(model.StatusProcessing)},
			prepare: func(rm *mocks.OrderRepository, pm *mocks.OrderProductService) {
				order := &model.Order{ID: id, Status: model.StatusCanceled}
				rm.EXPECT().GetByID(mock.Anything, id).Return(order, nil)
			},
			wantError: customErrors.ErrInvalidInput,
		},
		{
			name: "Stock Update Fail with Rollback",
			args: args{id: id, newStatus: string(model.StatusCompleted)},
			prepare: func(rm *mocks.OrderRepository, pm *mocks.OrderProductService) {
				order := &model.Order{
					ID:        id,
					Status:    model.StatusProcessing,
					OrderType: model.OrderInbound,
					Items: []model.OrderItem{
						{ProductID: prodID1, Quantity: 10},
						{ProductID: prodID2, Quantity: 20},
					},
				}
				rm.EXPECT().GetByID(mock.Anything, id).Return(order, nil)
				pm.EXPECT().ChangeStock(mock.Anything, prodID1, 10, "income").Return(nil)
				pm.EXPECT().ChangeStock(mock.Anything, prodID2, 20, "income").Return(errors.New("stock error"))

				pm.EXPECT().ChangeStock(mock.Anything, prodID1, 10, "expense").Return(nil)
			},
			wantError: errors.New("stock error"),
		},
		{
			name: "Repo Update Fail after Stock Update",
			args: args{id: id, newStatus: string(model.StatusCompleted)},
			prepare: func(rm *mocks.OrderRepository, pm *mocks.OrderProductService) {
				order := &model.Order{
					ID:        id,
					Status:    model.StatusProcessing,
					OrderType: model.OrderInbound,
					Items:     []model.OrderItem{{ProductID: prodID1, Quantity: 10}},
				}
				rm.EXPECT().GetByID(mock.Anything, id).Return(order, nil)
				pm.EXPECT().ChangeStock(mock.Anything, prodID1, 10, "income").Return(nil)
				rm.EXPECT().Update(mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			wantError: errors.New("db error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rm := mocks.NewOrderRepository(t)
			pm := mocks.NewOrderProductService(t)

			if tc.prepare != nil {
				tc.prepare(rm, pm)
			}

			svc := service.NewOrderService(rm, nil, pm, newDiscardLogger())
			got, err := svc.Update(context.Background(), tc.args.id, tc.args.newStatus)

			if tc.wantError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantError.Error())
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, model.OrderStatus(tc.args.newStatus), got.Status)
			}
		})
	}
}

func TestOrderService_Delete(t *testing.T) {
	id := uuid.New()
	tests := []struct {
		name      string
		prepare   func(rm *mocks.OrderRepository)
		wantError error
	}{
		{
			name: "Success Pending",
			prepare: func(rm *mocks.OrderRepository) {
				rm.EXPECT().GetByID(mock.Anything, id).Return(&model.Order{ID: id, Status: model.StatusPending}, nil)
				rm.EXPECT().Delete(mock.Anything, id).Return(nil)
			},
			wantError: nil,
		},
		{
			name: "Success Canceled",
			prepare: func(rm *mocks.OrderRepository) {
				rm.EXPECT().GetByID(mock.Anything, id).Return(&model.Order{ID: id, Status: model.StatusCanceled}, nil)
				rm.EXPECT().Delete(mock.Anything, id).Return(nil)
			},
			wantError: nil,
		},
		{
			name: "Fail Processing",
			prepare: func(rm *mocks.OrderRepository) {
				rm.EXPECT().GetByID(mock.Anything, id).Return(&model.Order{ID: id, Status: model.StatusProcessing}, nil)
			},
			wantError: customErrors.ErrInvalidInput,
		},
		{
			name: "GetByID Error",
			prepare: func(rm *mocks.OrderRepository) {
				rm.EXPECT().GetByID(mock.Anything, id).Return(nil, errors.New("not found"))
			},
			wantError: errors.New("not found"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rm := mocks.NewOrderRepository(t)
			if tc.prepare != nil {
				tc.prepare(rm)
			}

			svc := service.NewOrderService(rm, nil, nil, newDiscardLogger())
			err := svc.Delete(context.Background(), id)

			if tc.wantError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOrderService_GetList(t *testing.T) {
	type args struct {
		page     int
		pageSize int
	}
	expectedList := []*model.Order{{ID: uuid.New()}}

	tests := []struct {
		name      string
		args      args
		prepare   func(rm *mocks.OrderRepository)
		wantList  []*model.Order
		wantCount int
		wantError error
	}{
		{
			name: "Success Default",
			args: args{page: 0, pageSize: 0},
			prepare: func(rm *mocks.OrderRepository) {
				rm.EXPECT().GetList(mock.Anything, 10, 0).Return(expectedList, 1, nil)
			},
			wantList:  expectedList,
			wantCount: 1,
			wantError: nil,
		},
		{
			name: "Success Paging",
			args: args{page: 2, pageSize: 5},
			prepare: func(rm *mocks.OrderRepository) {
				rm.EXPECT().GetList(mock.Anything, 5, 5).Return(expectedList, 10, nil)
			},
			wantList:  expectedList,
			wantCount: 10,
			wantError: nil,
		},
		{
			name: "Repo Error",
			args: args{page: 1, pageSize: 10},
			prepare: func(rm *mocks.OrderRepository) {
				rm.EXPECT().GetList(mock.Anything, 10, 0).Return(nil, 0, errors.New("db fail"))
			},
			wantList:  nil,
			wantCount: 0,
			wantError: errors.New("db fail"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rm := mocks.NewOrderRepository(t)
			if tc.prepare != nil {
				tc.prepare(rm)
			}

			svc := service.NewOrderService(rm, nil, nil, newDiscardLogger())
			got, count, err := svc.GetList(context.Background(), tc.args.page, tc.args.pageSize)

			if tc.wantError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.wantError, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantList, got)
				assert.Equal(t, tc.wantCount, count)
			}
		})
	}
}

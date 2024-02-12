package eventschemas

import (
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	"github.com/kofc7186/fundraiser-manager/pkg/types"
)

const (
	OrderCreatedType   = "org.kofc7186.fundraiserManager.order.created"
	OrderUpdatedType   = "org.kofc7186.fundraiserManager.order.updated"
	OrderDeletedType   = "org.kofc7186.fundraiserManager.order.deleted"
	OrderStartedType   = "org.kofc7186.fundraiserManager.order.started"
	OrderReleasedType  = "org.kofc7186.fundraiserManager.order.released"
	OrderPreparedType  = "org.kofc7186.fundraiserManager.order.prepared"
	OrderDeliveredType = "org.kofc7186.fundraiserManager.order.delivered"
	OrderCanceledType  = "org.kofc7186.fundraiserManager.order.canceled"
)

type BaseOrder struct {
	Order          types.Order `json:"order"`
	IdempotencyKey string      `json:"idempotencyKey"`
}

type OrderCreated struct {
	BaseOrder
}

func NewOrderCreated(order *types.Order) (*cloudevents.Event, error) {
	event := newEvent(OrderCreatedType)
	event.SetSubject(order.ID)

	oc := &OrderCreated{
		BaseOrder: BaseOrder{
			Order:          *order,
			IdempotencyKey: uuid.Must(uuid.NewV7()).String(),
		},
	}

	if err := event.SetData(applicationJSON, oc); err != nil {
		return nil, err
	}
	return event, nil
}

type OrderUpdated struct {
	BaseOrder
	OldOrder      types.Order `json:"oldOrder"`
	UpdatedFields []string    `json:"updatedFields"`
}

func NewOrderUpdated(oldOrder, newOrder *types.Order, fieldMask []string) (*cloudevents.Event, error) {
	event := newEvent(OrderUpdatedType)
	event.SetSubject(newOrder.ID)

	ou := &OrderUpdated{
		BaseOrder: BaseOrder{
			Order:          *newOrder,
			IdempotencyKey: uuid.Must(uuid.NewV7()).String(),
		},
		OldOrder:      *oldOrder,
		UpdatedFields: fieldMask,
	}

	if err := event.SetData(applicationJSON, ou); err != nil {
		return nil, err
	}
	return event, nil
}

type OrderDeleted struct {
	BaseOrder
}

func NewOrderDeleted(order *types.Order) (*cloudevents.Event, error) {
	event := newEvent(OrderDeletedType)
	event.SetSubject(order.ID)

	od := &OrderDeleted{
		BaseOrder: BaseOrder{
			Order:          *order,
			IdempotencyKey: uuid.Must(uuid.NewV7()).String(),
		},
	}

	if err := event.SetData(applicationJSON, od); err != nil {
		return nil, err
	}
	return event, nil
}

type OrderStarted struct {
	StartTime time.Time `json:"startTime"`
}

func NewOrderStarted(id string) (*cloudevents.Event, error) {
	event := newEvent(OrderStartedType)
	event.SetSubject(id)

	os := &OrderStarted{
		StartTime: time.Now(),
	}

	if err := event.SetData(applicationJSON, os); err != nil {
		return nil, err
	}
	return event, nil
}

type OrderReleased struct {
	ReleaseTime time.Time `json:"releaseTime"`
}

func NewOrderReleased(id string) (*cloudevents.Event, error) {
	event := newEvent(OrderReleasedType)
	event.SetSubject(id)

	os := &OrderReleased{
		ReleaseTime: time.Now(),
	}

	if err := event.SetData(applicationJSON, os); err != nil {
		return nil, err
	}
	return event, nil
}

type OrderPrepared struct {
	PrepareTime time.Time `json:"prepareTime"`
}

func NewOrderPrepared(id string) (*cloudevents.Event, error) {
	event := newEvent(OrderPreparedType)
	event.SetSubject(id)

	os := &OrderPrepared{
		PrepareTime: time.Now(),
	}

	if err := event.SetData(applicationJSON, os); err != nil {
		return nil, err
	}
	return event, nil
}

type OrderDelivered struct {
	DeliveryTime time.Time `json:"deliveryTime"`
}

func NewOrderDelivered(id string) (*cloudevents.Event, error) {
	event := newEvent(OrderDeliveredType)
	event.SetSubject(id)

	os := &OrderDelivered{
		DeliveryTime: time.Now(),
	}

	if err := event.SetData(applicationJSON, os); err != nil {
		return nil, err
	}
	return event, nil
}

type OrderCanceled struct {
	CancelTime time.Time `json:"cancelTime"`
}

func NewOrderCanceled(id string) (*cloudevents.Event, error) {
	event := newEvent(OrderCanceledType)
	event.SetSubject(id)

	os := &OrderCanceled{
		CancelTime: time.Now(),
	}

	if err := event.SetData(applicationJSON, os); err != nil {
		return nil, err
	}
	return event, nil
}

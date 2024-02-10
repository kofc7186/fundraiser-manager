package eventschemas

import (
	"github.com/kofc7186/fundraiser-manager/pkg/square/types/webhooks"
	"github.com/kofc7186/fundraiser-manager/pkg/types"
)

const (
	OrderReceivedType = "org.kofc7186.fundraiserManager.orderReceived"
	OrderUpdatedType  = "org.kofc7186.fundraiserManager.orderUpdated"
)

type BaseOrder struct {
	Order          *types.Order `json:"order"`
	IdempotencyKey string       `json:"idempotencyKey"`
}

type OrderReceived struct {
	BaseOrder
	Raw *webhooks.OrderCreated `json:"raw"`
}

type OrderUpdated struct {
	BaseOrder
	Raw *webhooks.OrderUpdated `json:"raw"`
}

/*

func NewOrderReceived(squareOrderCreatedEvent *webhooks.OrderCreated) (*cloudevents.Event, error) {
	event := newEvent(OrderReceivedType)
	event.SetSubject(squareOrderCreatedEvent.Data.Object.OrderCreated.OrderId)

	p, err := types.CreateInternalOrderFromSquareOrder(squareOrderCreatedEvent.Data.Object.OrderCreated)
	if err != nil {
		return nil, err
	}

	pr := &OrderReceived{
		BaseOrder: BaseOrder{
			Order:          p,
			IdempotencyKey: squareOrderCreatedEvent.EventID,
		},
		Raw: squareOrderCreatedEvent,
	}

	if err = event.SetData(applicationJSON, pr); err != nil {
		return nil, err
	}
	return event, nil
}

func NewOrderUpdated(squareOrderUpdatedEvent *webhooks.OrderUpdated) (*cloudevents.Event, error) {
	event := newEvent(OrderUpdatedType)
	event.SetSubject(squareOrderUpdatedEvent.Data.Object.OrderUpdated.OrderId)

	p, err := types.CreateInternalOrderFromSquareOrder(squareOrderUpdatedEvent.Data.Object.Order)
	if err != nil {
		return nil, err
	}

	pu := &OrderUpdated{
		BaseOrder: BaseOrder{
			Order:          p,
			IdempotencyKey: squareOrderUpdatedEvent.EventID,
		},
		Raw: squareOrderUpdatedEvent,
	}

	if err = event.SetData(applicationJSON, pu); err != nil {
		return nil, err
	}
	return event, nil
}

*/

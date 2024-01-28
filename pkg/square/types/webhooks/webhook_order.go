package webhooks

import (
	"github.com/kofc7186/fundraiser-manager/pkg/square/types/models"
)

const (
	SQUARE_WEBHOOK_ORDER_CREATED = "order.created"
	SQUARE_WEBHOOK_ORDER_UPDATED = "order.updated"
)

type OrderCreated struct {
	WebhookBase
	Data OrderCreatedEventData `json:"data"`
}

type OrderCreatedEventData struct {
	Type   string                  `json:"type"`
	ID     string                  `json:"id"`
	Object OrderCreatedEventObject `json:"object"`
}

type OrderCreatedEventObject struct {
	Order models.Order `json:"order"`
}

type OrderUpdated struct {
	WebhookBase
	Data OrderUpdatedEventData `json:"data"`
}

type OrderUpdatedEventData struct {
	Type   string                  `json:"type"`
	ID     string                  `json:"id"`
	Object OrderUpdatedEventObject `json:"object"`
}

type OrderUpdatedEventObject struct {
	Order models.Order `json:"order"`
}

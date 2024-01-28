package webhooks

import (
	"github.com/kofc7186/fundraiser-manager/pkg/square/types/models"
)

const (
	SQUARE_WEBHOOK_CUSTOMER_CREATED = "customer.created"
	SQUARE_WEBHOOK_CUSTOMER_UPDATED = "customer.updated"
)

type CustomerCreated struct {
	WebhookBase
	Data CustomerCreatedEventData `json:"data"`
}

type CustomerCreatedEventData struct {
	Type   string                     `json:"type"`
	ID     string                     `json:"id"`
	Object CustomerCreatedEventObject `json:"object"`
}

type CustomerCreatedEventObject struct {
	Customer models.Customer `json:"customer"`
}

type CustomerUpdated struct {
	WebhookBase
	Data CustomerUpdatedEventData `json:"data"`
}

type CustomerUpdatedEventData struct {
	Type   string                     `json:"type"`
	ID     string                     `json:"id"`
	Object CustomerUpdatedEventObject `json:"object"`
}

type CustomerUpdatedEventObject struct {
	Customer models.Customer `json:"customer"`
}

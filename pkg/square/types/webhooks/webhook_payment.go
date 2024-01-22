package webhooks

import (
	"github.com/kofc7186/fundraiser-manager/pkg/square/types/models"
)

const (
	SQUARE_WEBHOOK_PAYMENT_CREATED = "payment.created"
	SQUARE_WEBHOOK_PAYMENT_UPDATED = "payment.updated"
)

type PaymentCreated struct {
	WebhookBase
	Data PaymentCreatedEventData `json:"data"`
}

type PaymentCreatedEventData struct {
	Type   string                    `json:"type"`
	ID     string                    `json:"id"`
	Object PaymentCreatedEventObject `json:"object"`
}

type PaymentCreatedEventObject struct {
	Payment models.Payment `json:"payment"`
}

type PaymentUpdated struct {
	WebhookBase
	Data PaymentUpdatedEventData `json:"data"`
}

type PaymentUpdatedEventData struct {
	Type   string                    `json:"type"`
	ID     string                    `json:"id"`
	Object PaymentUpdatedEventObject `json:"object"`
}

type PaymentUpdatedEventObject struct {
	Payment models.Payment `json:"payment"`
}

package webhooks

import (
	"github.com/kofc7186/fundraiser-manager/pkg/square/types/models"
)

const (
	SQUARE_WEBHOOK_REFUND_CREATED = "refund.created"
	SQUARE_WEBHOOK_REFUND_UPDATED = "refund.updated"
)

type RefundCreated struct {
	WebhookBase
	Data RefundCreatedEventData `json:"data"`
}

type RefundCreatedEventData struct {
	Type   string                   `json:"type"`
	ID     string                   `json:"id"`
	Object RefundCreatedEventObject `json:"object"`
}

type RefundCreatedEventObject struct {
	Refund models.PaymentRefund `json:"refund"`
}

type RefundUpdated struct {
	WebhookBase
	Data RefundUpdatedEventData `json:"data"`
}

type RefundUpdatedEventData struct {
	Type   string                   `json:"type"`
	ID     string                   `json:"id"`
	Object RefundUpdatedEventObject `json:"object"`
}

type RefundUpdatedEventObject struct {
	Refund models.PaymentRefund `json:"refund"`
}

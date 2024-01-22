package webhooks

import "time"

type WebhookBase struct {
	MerchantID string    `json:"merchant_id"`
	Type       string    `json:"type"`
	EventID    string    `json:"event_id"`
	CreatedAt  time.Time `json:"created_at"`
}

func (wb *WebhookBase) ID() string {
	return wb.EventID
}

type SquareWebhookEvent interface {
	ID() string
}

package types

import (
	"fmt"
	"time"

	"github.com/kofc7186/fundraiser-manager/pkg/square/types/models"
)

type RefundStatus string

const (
	REFUND_STATUS_UNKNOWN   RefundStatus = ""
	REFUND_STATUS_PENDING   RefundStatus = "PENDING"
	REFUND_STATUS_REJECTED  RefundStatus = "REJECTED"
	REFUND_STATUS_COMPLETED RefundStatus = "COMPLETED"
	REFUND_STATUS_FAILED    RefundStatus = "FAILED"
)

func parseRefundStatus(status string) (RefundStatus, error) {
	switch RefundStatus(status) {
	case REFUND_STATUS_PENDING:
		return REFUND_STATUS_PENDING, nil
	case REFUND_STATUS_REJECTED:
		return REFUND_STATUS_REJECTED, nil
	case REFUND_STATUS_COMPLETED:
		return REFUND_STATUS_COMPLETED, nil
	case REFUND_STATUS_FAILED:
		return REFUND_STATUS_FAILED, nil
	}
	return REFUND_STATUS_UNKNOWN, fmt.Errorf("%s is not a valid RefundStatus", status)
}

type Refund struct {
	Expiration        time.Time       `json:"expiration"`
	FeeAmount         float64         `json:"feeAmount"`
	ID                string          `json:"id"`
	IdempotencyKeys   map[string]bool `json:"idempotencyKeys"`
	Reason            string          `json:"reason"`
	RefundAmount      float64         `json:"refundAmount"`
	SquarePaymentID   string          `json:"squarePaymentID"`
	SquareOrderID     string          `json:"squareOrderID"`
	SquareUpdatedTime time.Time       `json:"squareUpdatedTime"`
	Status            RefundStatus    `json:"status"`
	Unlinked          bool            `json:"unlinked"`
}

func CreateInternalRefundFromSquareRefund(squareRefund models.PaymentRefund) (*Refund, error) {
	r := &Refund{
		ID:              squareRefund.Id,
		Reason:          squareRefund.Reason,
		RefundAmount:    float64(squareRefund.AmountMoney.Amount / 100),
		SquarePaymentID: squareRefund.PaymentId,
		SquareOrderID:   squareRefund.OrderId,
		Unlinked:        squareRefund.Unlinked,
	}

	var err error
	// per https://developer.squareup.com/reference/square/payments-api/webhooks/refund.created
	// this will be an RFC3339 timestamp
	r.SquareUpdatedTime, err = time.Parse(time.RFC3339, squareRefund.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if r.Status, err = parseRefundStatus(squareRefund.Status); err != nil {
		return nil, err
	}

	for _, fee := range squareRefund.ProcessingFee {
		r.FeeAmount += float64(fee.AmountMoney.Amount / 100)
	}

	return r, nil
}

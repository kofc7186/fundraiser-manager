package types

import (
	"fmt"
	"time"

	"github.com/kofc7186/fundraiser-manager/pkg/square/types/models"
)

type PaymentStatus string
type PaymentSource string

const (
	PAYMENT_STATUS_UNKNOWN   PaymentStatus = ""
	PAYMENT_STATUS_PENDING   PaymentStatus = "PENDING"
	PAYMENT_STATUS_APPROVED  PaymentStatus = "APPROVED"
	PAYMENT_STATUS_COMPLETED PaymentStatus = "COMPLETED"
	PAYMENT_STATUS_CANCELED  PaymentStatus = "CANCELED"
	PAYMENT_STATUS_FAILED    PaymentStatus = "FAILED"

	PAYMENT_SOURCE_UNKNOWN   PaymentSource = ""
	PAYMENT_SOURCE_ONLINE    PaymentSource = "ONLINE"
	PAYMENT_SOURCE_IN_PERSON PaymentSource = "IN_PERSON"
)

func parsePaymentStatus(status string) (PaymentStatus, error) {
	switch PaymentStatus(status) {
	case PAYMENT_STATUS_PENDING:
		return PAYMENT_STATUS_PENDING, nil
	case PAYMENT_STATUS_APPROVED:
		return PAYMENT_STATUS_APPROVED, nil
	case PAYMENT_STATUS_COMPLETED:
		return PAYMENT_STATUS_COMPLETED, nil
	case PAYMENT_STATUS_CANCELED:
		return PAYMENT_STATUS_CANCELED, nil
	case PAYMENT_STATUS_FAILED:
		return PAYMENT_STATUS_FAILED, nil
	}
	return PAYMENT_STATUS_UNKNOWN, fmt.Errorf("%s is not a valid PaymentStatus", status)
}

func parsePaymentSource(source string) (PaymentSource, error) {
	switch PaymentSource(source) {
	case "ONLINE_STORE":
		return PAYMENT_SOURCE_ONLINE, nil
	case "SQUARE_POS":
		return PAYMENT_SOURCE_IN_PERSON, nil
	}
	return PAYMENT_SOURCE_UNKNOWN, fmt.Errorf("%s is not a valid PaymentSource", source)
}

type Payment struct {
	Expiration        time.Time       `json:"expiration"`
	FeeAmount         float64         `json:"feeAmount"`
	ID                string          `json:"id"`
	IdempotencyKeys   map[string]bool `json:"idempotencyKeys"`
	Note              string          `json:"note"`
	ReceiptURL        string          `json:"receiptURL"`
	RefundAmount      float64         `json:"refundAmount,omitempty"`
	Source            PaymentSource   `json:"source"`
	SquareCustomerID  string          `json:"squareCustomerID,omitempty"`
	SquareOrderID     string          `json:"squareOrderID"`
	SquareRefundIDs   []string        `json:"squareRefundIDs,omitempty"`
	SquareUpdatedTime time.Time       `json:"squareUpdatedTime"`
	Status            PaymentStatus   `json:"status"`
	TipAmount         float64         `json:"tipAmount"`
	TotalAmount       float64         `json:"totalAmount"`
}

func CreateInternalPaymentFromSquarePayment(squarePayment models.Payment) (*Payment, error) {
	p := &Payment{
		ID:               squarePayment.Id,
		Note:             squarePayment.Note,
		ReceiptURL:       squarePayment.ReceiptUrl,
		RefundAmount:     float64(squarePayment.RefundedMoney.Amount / 100),
		SquareCustomerID: squarePayment.CustomerId,
		SquareOrderID:    squarePayment.OrderId,
		SquareRefundIDs:  squarePayment.RefundIds,
		TipAmount:        float64(squarePayment.TipMoney.Amount / 100),
		TotalAmount:      float64(squarePayment.TotalMoney.Amount / 100),
	}

	var err error
	if squarePayment.ApplicationDetails != nil {
		p.Source, err = parsePaymentSource(squarePayment.ApplicationDetails.SquareProduct)
		if err != nil {
			return nil, err
		}
	} else {
		p.Source = PAYMENT_SOURCE_UNKNOWN
	}

	if p.Status, err = parsePaymentStatus(squarePayment.Status); err != nil {
		return nil, err
	}

	for _, fee := range squarePayment.ProcessingFee {
		p.FeeAmount += float64(fee.AmountMoney.Amount / 100)
	}

	// per https://developer.squareup.com/reference/square/payments-api/webhooks/payment.created
	// this will be an RFC3339 timestamp
	p.SquareUpdatedTime, err = time.Parse(time.RFC3339, squarePayment.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return p, nil
}

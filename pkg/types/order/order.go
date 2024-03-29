package order

import (
	"errors"
	"fmt"
	"time"

	"github.com/kofc7186/fundraiser-manager/pkg/square/types/models"
	paymentType "github.com/kofc7186/fundraiser-manager/pkg/types/payment"
)

// this is the state of the order object according to Square
type SquareOrderState string

const (
	SQUARE_ORDER_STATE_UNKNOWN   SquareOrderState = ""
	SQUARE_ORDER_STATE_OPEN      SquareOrderState = "OPEN"
	SQUARE_ORDER_STATE_COMPLETED SquareOrderState = "COMPLETED"
	SQUARE_ORDER_STATE_CANCELED  SquareOrderState = "CANCELED"
	SQUARE_ORDER_STATE_DRAFT     SquareOrderState = "DRAFT"
)

func parseSquareOrderState(state string) (SquareOrderState, error) {
	switch SquareOrderState(state) {
	case SQUARE_ORDER_STATE_OPEN:
		return SQUARE_ORDER_STATE_OPEN, nil
	case SQUARE_ORDER_STATE_COMPLETED:
		return SQUARE_ORDER_STATE_COMPLETED, nil
	case SQUARE_ORDER_STATE_CANCELED:
		return SQUARE_ORDER_STATE_CANCELED, nil
	case SQUARE_ORDER_STATE_DRAFT:
		return SQUARE_ORDER_STATE_DRAFT, nil
	}
	return SQUARE_ORDER_STATE_UNKNOWN, fmt.Errorf("%q is not a valid SquareOrderState", state)
}

// This is the status of the order as it flows through fundraiser-manager
type OrderStatus string

const (
	ORDER_STATUS_UNKNOWN  OrderStatus = ""
	ORDER_STATUS_ONLINE   OrderStatus = "ONLINE"
	ORDER_STATUS_PRESENT  OrderStatus = "PRESENT"
	ORDER_STATUS_LABELED  OrderStatus = "LABELED"
	ORDER_STATUS_READY    OrderStatus = "READY"
	ORDER_STATUS_CLOSED   OrderStatus = "CLOSED"
	ORDER_STATUS_CANCELED OrderStatus = "CANCELED"
)

func parseOrderStatus(status string) (OrderStatus, error) {
	switch OrderStatus(status) {
	case ORDER_STATUS_ONLINE:
		return ORDER_STATUS_ONLINE, nil
	case ORDER_STATUS_PRESENT:
		return ORDER_STATUS_PRESENT, nil
	case ORDER_STATUS_LABELED:
		return ORDER_STATUS_LABELED, nil
	case ORDER_STATUS_READY:
		return ORDER_STATUS_READY, nil
	case ORDER_STATUS_CLOSED:
		return ORDER_STATUS_CLOSED, nil
	case ORDER_STATUS_CANCELED:
		return ORDER_STATUS_CANCELED, nil
	}
	return ORDER_STATUS_UNKNOWN, fmt.Errorf("%q is not a valid OrderStatus", status)
}

// This is the type of the item within an order
type OrderItemType string

const (
	ORDER_ITEM_TYPE_UNKNOWN       OrderItemType = ""
	ORDER_ITEM_TYPE_ITEM          OrderItemType = "ITEM"
	ORDER_ITEM_TYPE_CUSTOM_AMOUNT OrderItemType = "CUSTOM_AMOUNT"
	ORDER_ITEM_TYPE_GIFT_CARD     OrderItemType = "GIFT_CARD"
)

func parseOrderItemType(itemType string) (OrderItemType, error) {
	switch OrderItemType(itemType) {
	case ORDER_ITEM_TYPE_ITEM:
		return ORDER_ITEM_TYPE_ITEM, nil
	case ORDER_ITEM_TYPE_CUSTOM_AMOUNT:
		return ORDER_ITEM_TYPE_CUSTOM_AMOUNT, nil
	case ORDER_ITEM_TYPE_GIFT_CARD:
		return ORDER_ITEM_TYPE_GIFT_CARD, nil
	}
	return ORDER_ITEM_TYPE_UNKNOWN, fmt.Errorf("%q is not a valid OrderItemType", itemType)
}

type OrderStatusTransition struct {
	PreviousStatus OrderStatus
	Status         OrderStatus
	Timestamp      time.Time
}

type OrderItem struct {
	SquareCatalogObjectID string          `json:"squareCatalogObjectID" firestore:"squareCatalogObjectID"`
	Modifiers             []OrderModifier `json:"modifiers" firestore:"modifiers"`
	Name                  string          `json:"name" firestore:"name"`
	Note                  string          `json:"note" firestore:"note"`
	Quantity              string          `json:"quantity" firestore:"quantity"`
	SquareItemType        OrderItemType   `json:"squareItemType" firestore:"squareItemType"`
	Variation             string          `json:"variation" firestore:"variation"`
}

type OrderModifier struct {
	Name                  string `json:"name" firestore:"name"`
	Quantity              string `json:"quantity" firestore:"quantity"`
	SquareCatalogObjectID string `json:"squareCatalogObjectID" firestore:"squareCatalogObjectID"`
}

type Order struct {
	CreatedTime       time.Time                 `json:"createdTime" firestore:"createdTime"`
	DisplayName       string                    `json:"displayName" firestore:"displayName"`
	EmailAddress      string                    `json:"emailAddress" firestore:"emailAddress"`
	Expedite          bool                      `json:"expedite" firestore:"expedite"`
	Expiration        time.Time                 `json:"expiration" firestore:"expiration"`
	FeeAmount         float64                   `json:"feeAmount" firestore:"feeAmount"`
	FirstName         string                    `json:"firstName" firestore:"firstName"`
	LastName          string                    `json:"lastName" firestore:"lastName"`
	ID                string                    `json:"id" firestore:"id"`
	IdempotencyKeys   map[string]bool           `json:"idempotencyKeys" firestore:"idempotencyKeys"`
	Items             []OrderItem               `json:"items" firestore:"items"`
	KnightOfColumbus  bool                      `json:"isKnight" firestore:"isKnight"`
	LabelIDs          []string                  `json:"labelIDs" firestore:"labelIDs"`
	Number            uint16                    `json:"number" firestore:"number"` // This should be autogenerated by Firestore upon insert
	Note              string                    `json:"note" firestore:"note"`
	PhoneNumber       string                    `json:"phoneNumber" firestore:"phoneNumber"`
	ReceiptURL        string                    `json:"receiptURL" firestore:"receiptURL"`
	Source            paymentType.PaymentSource `json:"source" firestore:"source"`
	SquareCustomerID  string                    `json:"squareCustomerID" firestore:"squareCustomerID"`
	SquareOrderState  SquareOrderState          `json:"squareOrderState" firestore:"squareOrderState"`
	SquarePaymentID   string                    `json:"squarePaymentID" firestore:"squarePaymentID"`
	SquareUpdatedTime time.Time                 `json:"squareUpdatedTime" firestore:"squareUpdatedTime"`
	Status            OrderStatus               `json:"status" firestore:"status"`
	StatusTransitions []OrderStatusTransition   `json:"statusTransitions" firestore:"statusTransitions"`
	TipAmount         float64                   `json:"tipAmount" firestore:"tipAmount"`
	TotalAmount       float64                   `json:"totalAmount" firestore:"totalAmount"`
	Version           int32                     `json:"version" firestore:"version"`
}

func CreateInternalOrderFromSquareOrder(squareOrder models.Order) (*Order, error) {
	o := &Order{
		Expedite: false, // make default explicit
		ID:       squareOrder.Id,
		Status:   ORDER_STATUS_UNKNOWN,
		Version:  squareOrder.Version,
	}

	var err error
	if squareOrder.CustomerId != "" {
		o.SquareCustomerID = squareOrder.CustomerId
	}

	if len(squareOrder.Fulfillments) == 1 {
		if pickupDetails := squareOrder.Fulfillments[0].PickupDetails; pickupDetails != nil {
			o.Note = pickupDetails.Note

			if recipient := pickupDetails.Recipient; recipient != nil {
				// if customerID wasn't explicitly set before, let's try to get it from the fulfillment details
				if o.SquareCustomerID != "" {
					o.SquareCustomerID = recipient.CustomerId
				}
				if recipient.DisplayName != "" {
					o.DisplayName = recipient.DisplayName
					// TODO: set first and last name based on this
				}
				if recipient.EmailAddress != "" {
					o.EmailAddress = recipient.EmailAddress
				}
				if recipient.PhoneNumber != "" {
					o.PhoneNumber = recipient.PhoneNumber
				}
			}
		}
	}

	// TODO: orders can technically be associated with multiple payments, but our code assumes it is a 1:1 mapping
	if len(squareOrder.Tenders) == 1 {
		tender := squareOrder.Tenders[0]

		if o.SquarePaymentID == "" {
			o.SquarePaymentID = tender.PaymentId
		}
		if o.SquareCustomerID == "" {
			o.SquareCustomerID = tender.CustomerId
		}
	}

	for _, item := range squareOrder.LineItems {
		orderItem := OrderItem{
			Name:                  item.Name,
			Note:                  item.Note,
			Quantity:              item.Quantity,
			SquareCatalogObjectID: item.CatalogObjectId,
			Variation:             item.VariationName,
		}

		orderItem.SquareItemType, err = parseOrderItemType(item.ItemType)
		if err != nil {
			return nil, err
		}

		for _, modifier := range item.Modifiers {
			orderItemModifier := OrderModifier{
				Name:                  modifier.Name,
				Quantity:              modifier.Quantity,
				SquareCatalogObjectID: modifier.CatalogObjectId,
			}
			orderItem.Modifiers = append(orderItem.Modifiers, orderItemModifier)
		}
		o.Items = append(o.Items, orderItem)
	}

	if o.SquareOrderState, err = parseSquareOrderState(squareOrder.State); err != nil {
		return nil, err
	}

	// per https://developer.squareup.com/reference/square/orders-api/webhooks/order.created
	// this will be an RFC3339 timestamp
	o.CreatedTime, err = time.Parse(time.RFC3339, squareOrder.CreatedAt)
	if err != nil {
		return nil, err
	}

	// per https://developer.squareup.com/reference/square/orders-api/webhooks/order.created
	// this will be an RFC3339 timestamp
	o.SquareUpdatedTime, err = time.Parse(time.RFC3339, squareOrder.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return o, nil
}

func CreateOrderFromPayment(payment paymentType.Payment) (*Order, error) {
	switch payment.Status {
	case paymentType.PAYMENT_STATUS_CANCELED, paymentType.PAYMENT_STATUS_FAILED:
		return nil, errors.New("invalid payment state")
	}

	o := &Order{
		DisplayName:      fmt.Sprintf("%s %s", payment.FirstName, payment.LastName),
		EmailAddress:     payment.EmailAddress,
		Expedite:         false, // make default explicit
		FeeAmount:        payment.FeeAmount,
		FirstName:        payment.FirstName,
		ID:               payment.SquareOrderID,
		LastName:         payment.LastName,
		ReceiptURL:       payment.ReceiptURL,
		Source:           payment.Source,
		Status:           ORDER_STATUS_UNKNOWN,
		SquareCustomerID: payment.SquareCustomerID,
		SquarePaymentID:  payment.ID,
		TipAmount:        payment.TipAmount,
		TotalAmount:      payment.TotalAmount,
	}

	return o, nil
}

func (o *Order) Update(eventType string, proposedOrder *Order) error {
	for eventTypeRegexp := range updateMap {
		if eventTypeRegexp.MatchString(eventType) {
			continue
		}
	}

	return nil
}

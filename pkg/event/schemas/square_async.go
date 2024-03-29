package schemas

import (
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/kofc7186/fundraiser-manager/pkg/square/types/models"
	"github.com/kofc7186/fundraiser-manager/pkg/types/customer"
	"github.com/kofc7186/fundraiser-manager/pkg/types/order"
	"github.com/kofc7186/fundraiser-manager/pkg/types/payment"
)

const (
	SquareGetPaymentRequestType        = "org.kofc7186.fundraiserManager.square.getPayment.request"
	SquareGetPaymentResponseType       = "org.kofc7186.fundraiserManager.square.getPayment.response"
	SquareListPaymentsRequestType      = "org.kofc7186.fundraiserManager.square.listPayments.request"
	SquareRetrieveOrderRequestType     = "org.kofc7186.fundraiserManager.square.retrieveOrder.request"
	SquareRetrieveOrderResponseType    = "org.kofc7186.fundraiserManager.square.retrieveOrder.response"
	SquareRetrieveCustomerRequestType  = "org.kofc7186.fundraiserManager.square.retrieveCustomer.request"
	SquareRetrieveCustomerResponseType = "org.kofc7186.fundraiserManager.square.retrieveCustomer.response"
)

func NewSquareGetPaymentRequest(id string) *cloudevents.Event {
	event := newEvent(SquareGetPaymentRequestType)
	event.SetSubject(id)

	return event
}

type SquareGetPaymentResponse struct {
	BasePayment
	RequestSource string
	Raw           models.GetPaymentResponse
}

func NewSquareGetPaymentResponse(source string, response models.GetPaymentResponse) (*cloudevents.Event, error) {
	event := newEvent(SquareGetPaymentResponseType)
	event.SetSubject(response.Payment.Id)

	payment, err := payment.CreateInternalPaymentFromSquarePayment(*response.Payment)
	if err != nil {
		return nil, err
	}

	sgpc := &SquareGetPaymentResponse{
		BasePayment: BasePayment{
			Payment:        payment,
			IdempotencyKey: "",
		},
		RequestSource: source,
		Raw:           response,
	}

	_ = event.SetData(applicationJSON, sgpc)
	return event, nil
}

type SquareListPaymentsRequest struct {
	BeginTime time.Time `json:"beginTime"`
	EndTime   time.Time `json:"endTime"`
}

func NewSquareListPaymentsRequest(beginTime, endTime time.Time) *cloudevents.Event {
	event := newEvent(SquareListPaymentsRequestType)

	slpr := &SquareListPaymentsRequest{
		BeginTime: beginTime,
		EndTime:   endTime,
	}
	_ = event.SetData(applicationJSON, slpr)
	return event
}

func NewSquareRetrieveOrderRequest(id string) *cloudevents.Event {
	event := newEvent(SquareRetrieveOrderRequestType)
	event.SetSubject(id)

	return event
}

type SquareRetrieveOrderResponse struct {
	BaseOrder
	RequestSource string
	Raw           models.RetrieveOrderResponse
}

func NewSquareRetrieveOrderResponse(source string, response models.RetrieveOrderResponse) (*cloudevents.Event, error) {
	event := newEvent(SquareRetrieveOrderResponseType)
	event.SetSubject(response.Order.Id)

	order, err := order.CreateInternalOrderFromSquareOrder(*response.Order)
	if err != nil {
		return nil, err
	}

	sgoc := &SquareRetrieveOrderResponse{
		BaseOrder: BaseOrder{
			Order:          order,
			IdempotencyKey: "",
		},
		RequestSource: source,
		Raw:           response,
	}

	_ = event.SetData(applicationJSON, sgoc)
	return event, nil
}

func NewSquareRetrieveCustomerRequest(id string) *cloudevents.Event {
	event := newEvent(SquareRetrieveCustomerRequestType)
	event.SetSubject(id)

	return event
}

type SquareRetrieveCustomerResponse struct {
	BaseCustomer
	RequestSource string
	Raw           models.RetrieveCustomerResponse
}

func NewSquareRetrieveCustomerResponse(source string, response models.RetrieveCustomerResponse) (*cloudevents.Event, error) {
	event := newEvent(SquareRetrieveCustomerResponseType)
	event.SetSubject(response.Customer.Id)

	customer, err := customer.CreateInternalCustomerFromSquareCustomer(*response.Customer)
	if err != nil {
		return nil, err
	}

	sgcc := &SquareRetrieveCustomerResponse{
		BaseCustomer: BaseCustomer{
			Customer:       customer,
			IdempotencyKey: "",
		},
		RequestSource: source,
		Raw:           response,
	}

	_ = event.SetData(applicationJSON, sgcc)
	return event, nil
}

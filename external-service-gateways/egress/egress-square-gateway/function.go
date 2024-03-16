package egresssquaregateway

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"cloud.google.com/go/pubsub"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/cloudevents/sdk-go/v2/event"

	retryablehttp "github.com/hashicorp/go-retryablehttp"

	"github.com/antihax/optional"
	eventschemas "github.com/kofc7186/fundraiser-manager/pkg/event/schemas"
	"github.com/kofc7186/fundraiser-manager/pkg/logging"
	"github.com/kofc7186/fundraiser-manager/pkg/square/api"
	"github.com/kofc7186/fundraiser-manager/pkg/square/types/models"
	"github.com/kofc7186/fundraiser-manager/pkg/util"

	_ "github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
)

const (
	FUNCTION_NAME       = "egress-square-gateway"
	PUBLISH_TIMEOUT_SEC = 2 * time.Second
)

var paymentResponseTopic *pubsub.Topic
var orderResponseTopic *pubsub.Topic
var customerResponseTopic *pubsub.Topic

var squareClient *api.APIClient

func init() {
	slog.SetDefault(logging.FunctionLogger(FUNCTION_NAME))

	psClient, err := pubsub.NewClient(context.Background(), util.GetEnvOrPanic("GCP_PROJECT"))
	if err != nil {
		panic(err)
	}

	SQUARE_PAYMENT_RESPONSE_TOPIC_PATH := util.GetEnvOrPanic("SQUARE_PAYMENT_RESPONSE_TOPIC_PATH")
	paymentResponseTopic = psClient.Topic(SQUARE_PAYMENT_RESPONSE_TOPIC_PATH)
	if ok, err := paymentResponseTopic.Exists(context.Background()); !ok || err != nil {
		panic(fmt.Sprintf("existence check for %s failed: %v", SQUARE_PAYMENT_RESPONSE_TOPIC_PATH, err))
	}

	SQUARE_ORDER_RESPONSE_TOPIC_PATH := util.GetEnvOrPanic("SQUARE_ORDER_RESPONSE_TOPIC_PATH")
	orderResponseTopic = psClient.Topic(SQUARE_ORDER_RESPONSE_TOPIC_PATH)
	if ok, err := orderResponseTopic.Exists(context.Background()); !ok || err != nil {
		panic(fmt.Sprintf("existence check for %s failed: %v", SQUARE_ORDER_RESPONSE_TOPIC_PATH, err))
	}

	SQUARE_CUSTOMER_RESPONSE_TOPIC_PATH := util.GetEnvOrPanic("SQUARE_CUSTOMER_RESPONSE_TOPIC_PATH")
	customerResponseTopic = psClient.Topic(SQUARE_CUSTOMER_RESPONSE_TOPIC_PATH)
	if ok, err := customerResponseTopic.Exists(context.Background()); !ok || err != nil {
		panic(fmt.Sprintf("existence check for %s failed: %v", SQUARE_CUSTOMER_RESPONSE_TOPIC_PATH, err))
	}

	// initialize Square Client
	var configuration *api.Configuration

	SQUARE_ENVIRONMENT := util.GetEnvOrPanic("SQUARE_ENVIRONMENT")
	if SQUARE_ENVIRONMENT == "sandbox" {
		configuration = api.NewSandboxConfiguration()
	} else {
		configuration = api.NewConfiguration()
	}
	if SQUARE_VERSION, ok := os.LookupEnv("SQUARE_VERSION"); ok {
		configuration.AddDefaultHeader("Square-Version", SQUARE_VERSION)
	}

	retryClient := retryablehttp.NewClient()
	retryClient.Logger = logging.Logger
	configuration.HTTPClient = retryClient.StandardClient()

	// configure authentication credentials
	configuration.AddDefaultHeader("Authorization", fmt.Sprintf("Bearer %s", util.GetEnvOrPanic("SQUARE_ACCESS_TOKEN")))

	squareClient = api.NewAPIClient(configuration)

	// do this last so we are ensured to have all the required clients established above
	functions.CloudEvent("EgressSquarePaymentGateway", EgressSquarePaymentGateway)
	functions.CloudEvent("EgressSquareOrderGateway", EgressSquareOrderGateway)
	functions.CloudEvent("EgressSquareCustomerGateway", EgressSquareCustomerGateway)
}

// EgressSquarePaymentGateway invokes the Square API to get the payment object for the specified request
func EgressSquarePaymentGateway(ctx context.Context, e event.Event) error {
	// there are two CloudEvents - one for the pubsub message "event", and then the data within
	var msg eventschemas.MessagePublishedData
	if err := e.DataAs(&msg); err != nil {
		slog.ErrorContext(ctx, err.Error(), "event", e)
		return err
	}

	nestedEvent := &event.Event{}
	if err := nestedEvent.UnmarshalJSON(msg.Message.Data); err != nil {
		return err
	}

	var responseEvents []*event.Event
	switch nestedEvent.Type() {
	case eventschemas.SquareGetPaymentRequestType:
		paymentID := nestedEvent.Subject()

		payment, httpResponse, err := squareClient.PaymentsApi.GetPayment(ctx, paymentID)
		if err != nil {
			slog.ErrorContext(ctx, "error getting payment from Square", "paymentID", paymentID, "error", err, "httpResponse", httpResponse)
			return err
		}

		if len(payment.Errors) != 0 {
			return fmt.Errorf("error(s) calling GetPayment: %v", payment.Errors)
		}

		responseEvent, err := eventschemas.NewSquareGetPaymentResponse(nestedEvent.Source(), payment)
		if err != nil {
			return err
		}
		responseEvents = append(responseEvents, responseEvent)
	case eventschemas.SquareListPaymentsRequestType:
		slpr := &eventschemas.SquareListPaymentsRequest{}
		if err := nestedEvent.DataAs(slpr); err != nil {
			return err
		}
		payments, httpResponse, err := squareClient.PaymentsApi.ListPayments(ctx, &api.PaymentsApiListPaymentsOpts{
			BeginTime: optional.NewString(slpr.BeginTime.Format(time.RFC3339)),
			EndTime:   optional.NewString(slpr.EndTime.Format(time.RFC3339)),
		})
		if err != nil {
			slog.ErrorContext(ctx, "error listing payments from Square", "error", err, "httpResponse", httpResponse)
			return err
		}

		if len(payments.Errors) != 0 {
			return fmt.Errorf("error(s) calling ListPayments: %v", payments.Errors)
		}
		for _, payment := range payments.Payments {
			responseEvent, err := eventschemas.NewSquareGetPaymentResponse(nestedEvent.Source(), models.GetPaymentResponse{Payment: &payment})
			if err != nil {
				return err
			}
			responseEvents = append(responseEvents, responseEvent)
		}
	}

	for _, responseEvent := range responseEvents {
		respBytes, err := responseEvent.MarshalJSON()
		if err != nil {
			return err
		}

		timeoutContext, cancel := context.WithTimeout(context.Background(), PUBLISH_TIMEOUT_SEC)
		defer cancel()

		publishResult := paymentResponseTopic.Publish(timeoutContext, &pubsub.Message{Data: respBytes})
		// this call blocks until complete or timeout occurs
		if _, err := publishResult.Get(timeoutContext); err != nil {
			slog.ErrorContext(ctx, err.Error())
			return err
		}
	}

	return nil
}

// EgressSquareOrderGateway invokes the Square API to get the order object for the specified request
func EgressSquareOrderGateway(ctx context.Context, e event.Event) error {
	// there are two CloudEvents - one for the pubsub message "event", and then the data within
	var msg eventschemas.MessagePublishedData
	if err := e.DataAs(&msg); err != nil {
		slog.ErrorContext(ctx, err.Error(), "event", e)
		return err
	}

	nestedEvent := &event.Event{}
	if err := nestedEvent.UnmarshalJSON(msg.Message.Data); err != nil {
		return err
	}

	orderID := nestedEvent.Subject()

	order, httpResponse, err := squareClient.OrdersApi.RetrieveOrder(ctx, orderID)
	if err != nil {
		slog.ErrorContext(ctx, "error getting order from Square", "orderID", orderID, "error", err, "httpResponse", httpResponse)
		return err
	}

	if len(order.Errors) != 0 {
		return fmt.Errorf("error(s) calling RetrieveOrder: %v", order.Errors)
	}

	responseEvent, err := eventschemas.NewSquareRetrieveOrderResponse(nestedEvent.Source(), order)
	if err != nil {
		return err
	}

	respBytes, err := responseEvent.MarshalJSON()
	if err != nil {
		return err
	}

	timeoutContext, cancel := context.WithTimeout(context.Background(), PUBLISH_TIMEOUT_SEC)
	defer cancel()

	publishResult := orderResponseTopic.Publish(timeoutContext, &pubsub.Message{Data: respBytes})
	// this call blocks until complete or timeout occurs
	if _, err := publishResult.Get(timeoutContext); err != nil {
		slog.ErrorContext(ctx, err.Error())
		return err
	}

	return nil
}

// EgressSquareCustomerGateway invokes the Square API to get the customer object for the specified request
func EgressSquareCustomerGateway(ctx context.Context, e event.Event) error {
	// there are two CloudEvents - one for the pubsub message "event", and then the data within
	var msg eventschemas.MessagePublishedData
	if err := e.DataAs(&msg); err != nil {
		slog.ErrorContext(ctx, err.Error(), "event", e)
		return err
	}

	nestedEvent := &event.Event{}
	if err := nestedEvent.UnmarshalJSON(msg.Message.Data); err != nil {
		return err
	}

	customerID := nestedEvent.Subject()

	customer, httpResponse, err := squareClient.CustomersApi.RetrieveCustomer(ctx, customerID)
	if err != nil {
		slog.ErrorContext(ctx, "error getting customer from Square", "customerID", customerID, "error", err, "httpResponse", httpResponse)
		return err
	}

	if len(customer.Errors) != 0 {
		return fmt.Errorf("error(s) calling RetrieveCustomer: %v", customer.Errors)
	}

	responseEvent, err := eventschemas.NewSquareRetrieveCustomerResponse(nestedEvent.Source(), customer)
	if err != nil {
		return err
	}

	respBytes, err := responseEvent.MarshalJSON()
	if err != nil {
		return err
	}

	timeoutContext, cancel := context.WithTimeout(context.Background(), PUBLISH_TIMEOUT_SEC)
	defer cancel()

	publishResult := customerResponseTopic.Publish(timeoutContext, &pubsub.Message{Data: respBytes})
	// this call blocks until complete or timeout occurs
	if _, err := publishResult.Get(timeoutContext); err != nil {
		slog.ErrorContext(ctx, err.Error())
		return err
	}

	return nil
}

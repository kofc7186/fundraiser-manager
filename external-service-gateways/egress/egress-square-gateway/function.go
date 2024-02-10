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

	eventschemas "github.com/kofc7186/fundraiser-manager/pkg/event-schemas"
	"github.com/kofc7186/fundraiser-manager/pkg/logging"
	"github.com/kofc7186/fundraiser-manager/pkg/square/api"
	"github.com/kofc7186/fundraiser-manager/pkg/util"

	_ "github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
)

const PUBLISH_TIMEOUT_SEC = 2 * time.Second

var paymentResponseTopic *pubsub.Topic
var orderResponseTopic *pubsub.Topic
var customerResponseTopic *pubsub.Topic

var squareClient *api.APIClient

var SQUARE_ACCESS_TOKEN string
var SQUARE_ENVIRONMENT string
var SQUARE_PAYMENT_RESPONSE_TOPIC_PATH string
var SQUARE_ORDER_RESPONSE_TOPIC_PATH string
var SQUARE_CUSTOMER_RESPONSE_TOPIC_PATH string

func init() {
	slog.SetDefault(logging.Logger)

	psClient, err := pubsub.NewClient(context.Background(), util.GetEnvOrPanic("GCP_PROJECT"))
	if err != nil {
		panic(err)
	}

	SQUARE_PAYMENT_RESPONSE_TOPIC_PATH = util.GetEnvOrPanic("SQUARE_PAYMENT_RESPONSE_TOPIC_PATH")
	paymentResponseTopic = psClient.Topic(SQUARE_PAYMENT_RESPONSE_TOPIC_PATH)
	if ok, err := paymentResponseTopic.Exists(context.Background()); !ok || err != nil {
		panic(fmt.Sprintf("existence check for %s failed: %v", SQUARE_PAYMENT_RESPONSE_TOPIC_PATH, err))
	}

	SQUARE_ORDER_RESPONSE_TOPIC_PATH = util.GetEnvOrPanic("SQUARE_ORDER_RESPONSE_TOPIC_PATH")
	orderResponseTopic = psClient.Topic(SQUARE_ORDER_RESPONSE_TOPIC_PATH)
	if ok, err := orderResponseTopic.Exists(context.Background()); !ok || err != nil {
		panic(fmt.Sprintf("existence check for %s failed: %v", SQUARE_ORDER_RESPONSE_TOPIC_PATH, err))
	}

	SQUARE_CUSTOMER_RESPONSE_TOPIC_PATH = util.GetEnvOrPanic("SQUARE_CUSTOMER_RESPONSE_TOPIC_PATH")
	customerResponseTopic = psClient.Topic(SQUARE_CUSTOMER_RESPONSE_TOPIC_PATH)
	if ok, err := customerResponseTopic.Exists(context.Background()); !ok || err != nil {
		panic(fmt.Sprintf("existence check for %s failed: %v", SQUARE_CUSTOMER_RESPONSE_TOPIC_PATH, err))
	}

	// initialize Square Client
	var configuration *api.Configuration

	SQUARE_ENVIRONMENT = util.GetEnvOrPanic("SQUARE_ENVIRONMENT")
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
	SQUARE_ACCESS_TOKEN = util.GetEnvOrPanic("SQUARE_ACCESS_TOKEN")
	configuration.AddDefaultHeader("Authorization", fmt.Sprintf("Bearer %s", SQUARE_ACCESS_TOKEN))

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
		slog.Error(err.Error(), "event", e)
		return err
	}

	nestedEvent := &event.Event{}
	if err := nestedEvent.UnmarshalJSON(msg.Message.Data); err != nil {
		return err
	}

	paymentID := nestedEvent.Subject()

	payment, _, err := squareClient.PaymentsApi.GetPayment(ctx, paymentID)
	if err != nil {
		// TODO: print http response headers if we have an error
		return err
	}

	if len(payment.Errors) != 0 {
		return fmt.Errorf("error(s) calling GetPayment: %v", payment.Errors)
	}

	responseEvent, err := eventschemas.NewSquareGetPaymentCompleted(*payment.Payment)
	if err != nil {
		return err
	}
	responseEvent.SetExtension("request_source", nestedEvent.Source())

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

	return nil
}

// EgressSquareOrderGateway invokes the Square API to get the order object for the specified request
func EgressSquareOrderGateway(ctx context.Context, e event.Event) error {
	// there are two CloudEvents - one for the pubsub message "event", and then the data within
	var msg eventschemas.MessagePublishedData
	if err := e.DataAs(&msg); err != nil {
		slog.Error(err.Error(), "event", e)
		return err
	}

	nestedEvent := &event.Event{}
	if err := nestedEvent.UnmarshalJSON(msg.Message.Data); err != nil {
		return err
	}

	orderID := nestedEvent.Subject()

	order, _, err := squareClient.OrdersApi.RetrieveOrder(ctx, orderID)
	if err != nil {
		// TODO: print http response headers if we have an error
		return err
	}

	if len(order.Errors) != 0 {
		return fmt.Errorf("error(s) calling RetrieveOrders: %v", order.Errors)
	}

	responseEvent, err := eventschemas.NewSquareGetOrderCompleted(*order.Order)
	if err != nil {
		return err
	}
	responseEvent.SetExtension("request_source", nestedEvent.Source())

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
		slog.Error(err.Error(), "event", e)
		return err
	}

	nestedEvent := &event.Event{}
	if err := nestedEvent.UnmarshalJSON(msg.Message.Data); err != nil {
		return err
	}

	customerID := nestedEvent.Subject()

	customer, _, err := squareClient.CustomersApi.RetrieveCustomer(ctx, customerID)
	if err != nil {
		// TODO: print http response headers if we have an error
		return err
	}

	if len(customer.Errors) != 0 {
		return fmt.Errorf("error(s) calling RetrieveCustomers: %v", customer.Errors)
	}

	responseEvent, err := eventschemas.NewSquareGetCustomerCompleted(*customer.Customer)
	if err != nil {
		return err
	}
	responseEvent.SetExtension("request_source", nestedEvent.Source())

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

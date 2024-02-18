package squarewebhookingress

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"

	"cloud.google.com/go/pubsub"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	eventschemas "github.com/kofc7186/fundraiser-manager/pkg/event/schemas"
	"github.com/kofc7186/fundraiser-manager/pkg/logging"
	squarewebhooktypes "github.com/kofc7186/fundraiser-manager/pkg/square/types/webhooks"
	"github.com/kofc7186/fundraiser-manager/pkg/square/webhooks"
	"github.com/kofc7186/fundraiser-manager/pkg/util"

	_ "github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
)

const (
	FUNCTION_NAME       = "square-webhook-ingress"
	PUBLISH_TIMEOUT_SEC = 2 * time.Second
)

var squareOrderRequestTopic *pubsub.Topic
var squarePaymentWebhookTopic *pubsub.Topic
var squareRefundWebhookTopic *pubsub.Topic
var squareCustomerWebhookTopic *pubsub.Topic
var SQUARE_SIGNATURE_KEY string
var WEBHOOK_URL string

func init() {
	slog.SetDefault(logging.FunctionLogger(FUNCTION_NAME))

	// if we don't have these environment variables set, we should panic ASAP
	SQUARE_SIGNATURE_KEY = util.GetEnvOrPanic("SQUARE_SIGNATURE_KEY")
	SQUARE_ORDER_REQUEST_TOPIC := util.GetEnvOrPanic("SQUARE_ORDER_REQUEST_TOPIC")
	SQUARE_PAYMENT_WEBHOOK_TOPIC := util.GetEnvOrPanic("SQUARE_PAYMENT_WEBHOOK_TOPIC")
	SQUARE_REFUND_WEBHOOK_TOPIC := util.GetEnvOrPanic("SQUARE_REFUND_WEBHOOK_TOPIC")
	SQUARE_CUSTOMER_WEBHOOK_TOPIC := util.GetEnvOrPanic("SQUARE_CUSTOMER_WEBHOOK_TOPIC")
	WEBHOOK_URL = util.GetEnvOrPanic("WEBHOOK_URL")

	psClient, err := pubsub.NewClient(context.Background(), util.GetEnvOrPanic("GCP_PROJECT"))
	if err != nil {
		panic(err)
	}

	squareOrderRequestTopic = psClient.Topic(SQUARE_ORDER_REQUEST_TOPIC)
	if ok, err := squareOrderRequestTopic.Exists(context.Background()); !ok || err != nil {
		panic(fmt.Sprintf("existence check for %s failed: %v", SQUARE_ORDER_REQUEST_TOPIC, err))
	}

	squarePaymentWebhookTopic = psClient.Topic(SQUARE_PAYMENT_WEBHOOK_TOPIC)
	if ok, err := squarePaymentWebhookTopic.Exists(context.Background()); !ok || err != nil {
		panic(fmt.Sprintf("existence check for %s failed: %v", SQUARE_PAYMENT_WEBHOOK_TOPIC, err))
	}

	squareRefundWebhookTopic = psClient.Topic(SQUARE_REFUND_WEBHOOK_TOPIC)
	if ok, err := squareRefundWebhookTopic.Exists(context.Background()); !ok || err != nil {
		panic(fmt.Sprintf("existence check for %s failed: %v", SQUARE_REFUND_WEBHOOK_TOPIC, err))
	}

	squareCustomerWebhookTopic = psClient.Topic(SQUARE_CUSTOMER_WEBHOOK_TOPIC)
	if ok, err := squareCustomerWebhookTopic.Exists(context.Background()); !ok || err != nil {
		panic(fmt.Sprintf("existence check for %s failed: %v", SQUARE_CUSTOMER_WEBHOOK_TOPIC, err))
	}

	// do this last so we are ensured to have all the required clients established above
	functions.HTTP("WebhookRouter", WebhookRouter)
}

// WebhookRouter is the function that routes the incoming request based on
func WebhookRouter(w http.ResponseWriter, r *http.Request) {
	// parse and validate the input came from Square
	webhookEvent, err := webhooks.VerifySquareWebhook(r, SQUARE_SIGNATURE_KEY, WEBHOOK_URL)
	if err != nil {
		slog.ErrorContext(r.Context(), err.Error())
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	// create the correct internal event
	var internalEvent *cloudevents.Event
	var pubTopic *pubsub.Topic
	switch t := webhookEvent.(type) {
	case *squarewebhooktypes.PaymentCreated:
		internalEvent, err = eventschemas.NewPaymentCreatedFromSquare(t)
		pubTopic = squarePaymentWebhookTopic
	case *squarewebhooktypes.PaymentUpdated:
		internalEvent, err = eventschemas.NewPaymentUpdatedFromSquare(t)
		pubTopic = squarePaymentWebhookTopic
	case *squarewebhooktypes.RefundCreated:
		internalEvent, err = eventschemas.NewRefundCreatedFromSquare(t)
		pubTopic = squareRefundWebhookTopic
	case *squarewebhooktypes.RefundUpdated:
		internalEvent, err = eventschemas.NewRefundUpdatedFromSquare(t)
		pubTopic = squareRefundWebhookTopic
	case *squarewebhooktypes.CustomerCreated:
		internalEvent, err = eventschemas.NewCustomerCreatedFromSquare(t)
		pubTopic = squareCustomerWebhookTopic
	case *squarewebhooktypes.CustomerUpdated:
		internalEvent, err = eventschemas.NewCustomerUpdatedFromSquare(t)
		pubTopic = squareCustomerWebhookTopic
	// these two types are different; since the Square webhook doesn't include the 'order' object, we immediately have to fetch it
	case *squarewebhooktypes.OrderCreated:
		internalEvent = eventschemas.NewSquareRetrieveOrderRequest(t.Data.Object.OrderCreated.OrderId)
		internalEvent.SetSource(squarewebhooktypes.SQUARE_WEBHOOK_ORDER_CREATED)
		pubTopic = squareOrderRequestTopic
	case *squarewebhooktypes.OrderUpdated:
		internalEvent = eventschemas.NewSquareRetrieveOrderRequest(t.Data.Object.OrderUpdated.OrderId)
		internalEvent.SetSource(squarewebhooktypes.SQUARE_WEBHOOK_ORDER_UPDATED)
		pubTopic = squareOrderRequestTopic
	default:
		err = errors.New("unsupported webhook event received")
		slog.ErrorContext(r.Context(), err.Error())
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	slog.DebugContext(r.Context(), fmt.Sprintf("event %T received", webhookEvent), "event", internalEvent)
	if err != nil {
		slog.ErrorContext(r.Context(), fmt.Sprintf("error creating internal event: %v", err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	// publish to correct topic
	eventJSON, err := internalEvent.MarshalJSON()
	if err != nil {
		slog.ErrorContext(r.Context(), err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	timeoutContext, cancel := context.WithTimeout(context.Background(), PUBLISH_TIMEOUT_SEC)
	defer cancel()

	publishResult := pubTopic.Publish(timeoutContext, &pubsub.Message{Data: eventJSON})
	messageID, err := publishResult.Get(timeoutContext) // this call blocks until complete or timeout occurs
	if err != nil {
		slog.ErrorContext(r.Context(), err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	// respond back to Square that we've successfully ingested the webhook event
	slog.DebugContext(r.Context(), fmt.Sprintf("successfully published %s", messageID), "event_id", webhookEvent.ID())
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("\"ok\""))
}

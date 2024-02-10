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
	eventschemas "github.com/kofc7186/fundraiser-manager/pkg/event-schemas"
	"github.com/kofc7186/fundraiser-manager/pkg/logging"
	whtypes "github.com/kofc7186/fundraiser-manager/pkg/square/types/webhooks"
	"github.com/kofc7186/fundraiser-manager/pkg/square/webhooks"
	"github.com/kofc7186/fundraiser-manager/pkg/util"

	_ "github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
)

const PUBLISH_TIMEOUT_SEC = 2 * time.Second

var squareOrderRequestTopic *pubsub.Topic
var paymentEventsTopic *pubsub.Topic
var refundEventsTopic *pubsub.Topic
var customerEventsTopic *pubsub.Topic
var SQUARE_SIGNATURE_KEY string
var WEBHOOK_URL string
var SQUARE_ORDER_REQUEST_TOPIC string
var PAYMENT_EVENTS_TOPIC string
var REFUND_EVENTS_TOPIC string
var CUSTOMER_EVENTS_TOPIC string

func init() {
	slog.SetDefault(logging.Logger)

	// if we don't have these environment variables set, we should panic ASAP
	SQUARE_SIGNATURE_KEY = util.GetEnvOrPanic("SQUARE_SIGNATURE_KEY")
	SQUARE_ORDER_REQUEST_TOPIC = util.GetEnvOrPanic("SQUARE_ORDER_REQUEST_TOPIC")
	PAYMENT_EVENTS_TOPIC = util.GetEnvOrPanic("PAYMENT_EVENTS_TOPIC")
	REFUND_EVENTS_TOPIC = util.GetEnvOrPanic("REFUND_EVENTS_TOPIC")
	CUSTOMER_EVENTS_TOPIC = util.GetEnvOrPanic("CUSTOMER_EVENTS_TOPIC")
	WEBHOOK_URL = util.GetEnvOrPanic("WEBHOOK_URL")

	psClient, err := pubsub.NewClient(context.Background(), util.GetEnvOrPanic("GCP_PROJECT"))
	if err != nil {
		panic(err)
	}

	squareOrderRequestTopic = psClient.Topic(SQUARE_ORDER_REQUEST_TOPIC)
	if ok, err := squareOrderRequestTopic.Exists(context.Background()); !ok || err != nil {
		panic(fmt.Sprintf("existence check for %s failed: %v", SQUARE_ORDER_REQUEST_TOPIC, err))
	}

	paymentEventsTopic = psClient.Topic(PAYMENT_EVENTS_TOPIC)
	if ok, err := paymentEventsTopic.Exists(context.Background()); !ok || err != nil {
		panic(fmt.Sprintf("existence check for %s failed: %v", PAYMENT_EVENTS_TOPIC, err))
	}

	refundEventsTopic = psClient.Topic(REFUND_EVENTS_TOPIC)
	if ok, err := refundEventsTopic.Exists(context.Background()); !ok || err != nil {
		panic(fmt.Sprintf("existence check for %s failed: %v", REFUND_EVENTS_TOPIC, err))
	}

	customerEventsTopic = psClient.Topic(CUSTOMER_EVENTS_TOPIC)
	if ok, err := customerEventsTopic.Exists(context.Background()); !ok || err != nil {
		panic(fmt.Sprintf("existence check for %s failed: %v", CUSTOMER_EVENTS_TOPIC, err))
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
	case *whtypes.PaymentCreated:
		internalEvent, err = eventschemas.NewPaymentReceived(t)
		pubTopic = paymentEventsTopic
	case *whtypes.PaymentUpdated:
		internalEvent, err = eventschemas.NewPaymentUpdated(t)
		pubTopic = paymentEventsTopic
	case *whtypes.RefundCreated:
		internalEvent, err = eventschemas.NewRefundReceived(t)
		pubTopic = refundEventsTopic
	case *whtypes.RefundUpdated:
		internalEvent, err = eventschemas.NewRefundUpdated(t)
		pubTopic = refundEventsTopic
	case *whtypes.CustomerCreated:
		internalEvent, err = eventschemas.NewCustomerReceived(t)
		pubTopic = customerEventsTopic
	case *whtypes.CustomerUpdated:
		internalEvent, err = eventschemas.NewCustomerUpdated(t)
		pubTopic = customerEventsTopic
	// these two types are different; since the Square webhook doesn't include the 'order' object, we immediately have to fetch it
	case *whtypes.OrderCreated:
		internalEvent = eventschemas.NewSquareGetOrderRequested(t.Data.Object.OrderCreated.OrderId)
		pubTopic = squareOrderRequestTopic
	case *whtypes.OrderUpdated:
		internalEvent = eventschemas.NewSquareGetOrderRequested(t.Data.Object.OrderUpdated.OrderId)
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

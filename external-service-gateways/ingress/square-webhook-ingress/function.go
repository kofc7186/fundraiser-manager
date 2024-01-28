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

var orderEventsTopic *pubsub.Topic
var paymentEventsTopic *pubsub.Topic
var refundEventsTopic *pubsub.Topic
var SQUARE_SIGNATURE_KEY string
var WEBHOOK_URL string
var ORDER_EVENTS_TOPIC string
var PAYMENT_EVENTS_TOPIC string
var REFUND_EVENTS_TOPIC string

func init() {
	slog.SetDefault(logging.Logger)

	// if we don't have these environment variables set, we should panic ASAP
	SQUARE_SIGNATURE_KEY = util.GetEnvOrPanic("SQUARE_SIGNATURE_KEY")
	ORDER_EVENTS_TOPIC = util.GetEnvOrPanic("ORDER_EVENTS_TOPIC")
	PAYMENT_EVENTS_TOPIC = util.GetEnvOrPanic("PAYMENT_EVENTS_TOPIC")
	REFUND_EVENTS_TOPIC = util.GetEnvOrPanic("REFUND_EVENTS_TOPIC")
	WEBHOOK_URL = util.GetEnvOrPanic("WEBHOOK_URL")

	psClient, err := pubsub.NewClient(context.Background(), util.GetEnvOrPanic("GCP_PROJECT"))
	if err != nil {
		panic(err)
	}

	orderEventsTopic = psClient.Topic(ORDER_EVENTS_TOPIC)
	if ok, err := orderEventsTopic.Exists(context.Background()); !ok || err != nil {
		panic(fmt.Sprintf("existence check for %s failed: %v", ORDER_EVENTS_TOPIC, err))
	}

	paymentEventsTopic = psClient.Topic(PAYMENT_EVENTS_TOPIC)
	if ok, err := paymentEventsTopic.Exists(context.Background()); !ok || err != nil {
		panic(fmt.Sprintf("existence check for %s failed: %v", PAYMENT_EVENTS_TOPIC, err))
	}

	refundEventsTopic = psClient.Topic(REFUND_EVENTS_TOPIC)
	if ok, err := refundEventsTopic.Exists(context.Background()); !ok || err != nil {
		panic(fmt.Sprintf("existence check for %s failed: %v", REFUND_EVENTS_TOPIC, err))
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
	switch t := webhookEvent.(type) {
	case *whtypes.OrderCreated:
		internalEvent, err = eventschemas.NewOrderReceived(t)
	case *whtypes.OrderUpdated:
		internalEvent, err = eventschemas.NewOrderUpdated(t)
	case *whtypes.PaymentCreated:
		internalEvent, err = eventschemas.NewPaymentReceived(t)
	case *whtypes.PaymentUpdated:
		internalEvent, err = eventschemas.NewPaymentUpdated(t)
	case *whtypes.RefundCreated:
		internalEvent, err = eventschemas.NewRefundReceived(t)
	case *whtypes.RefundUpdated:
		internalEvent, err = eventschemas.NewRefundUpdated(t)
	default:
		err = errors.New("unsupported webhook event received")
		slog.ErrorContext(r.Context(), err.Error())
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	if err != nil {
		slog.ErrorContext(r.Context(), fmt.Sprintf("error creating internal event: %v", err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	// publish to payments topic

	eventJSON, err := internalEvent.MarshalJSON()
	if err != nil {
		slog.ErrorContext(r.Context(), err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	timeoutContext, cancel := context.WithTimeout(context.Background(), PUBLISH_TIMEOUT_SEC)
	defer cancel()

	publishResult := paymentEventsTopic.Publish(timeoutContext, &pubsub.Message{Data: eventJSON})
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

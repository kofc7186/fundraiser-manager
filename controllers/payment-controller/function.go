package paymentcontroller

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/pubsub"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/googleapis/google-cloudevents-go/cloud/firestoredata"

	eventschemas "github.com/kofc7186/fundraiser-manager/pkg/event/schemas"
	"github.com/kofc7186/fundraiser-manager/pkg/logging"
	paymentType "github.com/kofc7186/fundraiser-manager/pkg/types/payment"
	refundType "github.com/kofc7186/fundraiser-manager/pkg/types/refund"
	"github.com/kofc7186/fundraiser-manager/pkg/util"

	_ "github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
)

const (
	FUNCTION_NAME       = "payment-controller"
	PUBLISH_TIMEOUT_SEC = 2 * time.Second
)

var firestoreClient *firestore.Client
var paymentDocPath string

var paymentEventsTopic *pubsub.Topic
var squarePaymentRequestTopic *pubsub.Topic

var expirationTime time.Time

func init() {
	slog.SetDefault(logging.FunctionLogger(FUNCTION_NAME))

	psClient, err := pubsub.NewClient(context.Background(), util.GetEnvOrPanic("GCP_PROJECT"))
	if err != nil {
		panic(err)
	}

	PAYMENT_EVENTS_TOPIC := util.GetEnvOrPanic("PAYMENT_EVENTS_TOPIC")
	paymentEventsTopic = psClient.Topic(PAYMENT_EVENTS_TOPIC)
	if ok, err := paymentEventsTopic.Exists(context.Background()); !ok || err != nil {
		panic(fmt.Sprintf("existence check for %s failed: %v", PAYMENT_EVENTS_TOPIC, err))
	}

	SQUARE_PAYMENT_REQUEST_TOPIC := util.GetEnvOrPanic("SQUARE_PAYMENT_REQUEST_TOPIC")
	squarePaymentRequestTopic = psClient.Topic(SQUARE_PAYMENT_REQUEST_TOPIC)
	if ok, err := squarePaymentRequestTopic.Exists(context.Background()); !ok || err != nil {
		panic(fmt.Sprintf("existence check for %s failed: %v", SQUARE_PAYMENT_REQUEST_TOPIC, err))
	}

	firestoreClient, err = firestore.NewClient(context.Background(), util.GetEnvOrPanic("GCP_PROJECT"))
	if err != nil {
		panic(err)
	}

	paymentDocPath = fmt.Sprintf("fundraisers/%s/payments", util.GetEnvOrPanic("FUNDRAISER_ID"))

	expirationTime, err = time.Parse(time.RFC3339, util.GetEnvOrPanic("EXPIRATION_TIME"))
	if err != nil {
		panic(err)
	}

	// do this last so we are ensured to have all the required clients established above
	functions.CloudEvent("ProcessSquarePaymentWebhookEvent", ProcessSquarePaymentWebhookEvent)
	functions.CloudEvent("ProcessSquarePaymentResponse", ProcessSquarePaymentWebhookEvent) // Square API responses just get written like inbound webhooks
	functions.CloudEvent("ProcessCDCEvent", ProcessCDCEvent)
	functions.CloudEvent("RefundWatcher", RefundWatcher)
}

// ProcessSquarePaymentWebhookEvent
func ProcessSquarePaymentWebhookEvent(ctx context.Context, e event.Event) error {
	// there are two CloudEvents - one for the pubsub message "event", and then the data within
	var msg eventschemas.MessagePublishedData
	if err := e.DataAs(&msg); err != nil {
		slog.ErrorContext(ctx, err.Error(), "event", e)
		return err
	}

	nestedEvent := &event.Event{}
	if err := nestedEvent.UnmarshalJSON(msg.Message.Data); err != nil {
		slog.ErrorContext(ctx, err.Error(), "event", e)
		return err
	}

	return writeSquarePaymentToFirestore(ctx, nestedEvent)
}

func writeSquarePaymentToFirestore(ctx context.Context, e *event.Event) error {
	paymentCreateRequest := false
	attemptedWrite := false

	var idempotencyKey string
	var proposedPayment *paymentType.Payment
	switch e.Type() {
	case eventschemas.PaymentCreatedFromSquareType:
		paymentCreateRequest = true

		pr := &eventschemas.PaymentReceivedFromSquare{}
		if err := e.DataAs(pr); err != nil {
			slog.ErrorContext(ctx, err.Error(), "event", e)
			return err
		}
		idempotencyKey = pr.BasePayment.IdempotencyKey
		proposedPayment = pr.BasePayment.Payment
	case eventschemas.PaymentUpdatedFromSquareType:
		pu := &eventschemas.PaymentUpdatedFromSquare{}
		if err := e.DataAs(pu); err != nil {
			slog.ErrorContext(ctx, err.Error(), "event", e)
			return err
		}
		idempotencyKey = pu.BasePayment.IdempotencyKey
		proposedPayment = pu.BasePayment.Payment
	case eventschemas.SquareGetPaymentResponseType:
		sgpr := &eventschemas.SquareGetPaymentResponse{}
		if err := e.DataAs(sgpr); err != nil {
			slog.ErrorContext(ctx, err.Error(), "event", e)
			return err
		}
		idempotencyKey = sgpr.BasePayment.IdempotencyKey
		proposedPayment = sgpr.BasePayment.Payment
	default:
		// TODO: slog
		return nil
	}

	// make sure to update the map to denote that we've processed this event already
	//
	// the boolean here is only to allow Firestore to map back to Go struct; the important
	// thing is that the key is put into the map
	proposedPayment.IdempotencyKeys = make(map[string]bool, 1)
	if idempotencyKey != "" {
		proposedPayment.IdempotencyKeys[idempotencyKey] = true
	}

	// ensure the firestore expiration timestamp is written in the appropriate field
	proposedPayment.Expiration = expirationTime

	docRef := firestoreClient.Doc(fmt.Sprintf("%s/%s", paymentDocPath, proposedPayment.ID))
	transaction := func(ctx context.Context, t *firestore.Transaction) error {
		docSnap, err := t.Get(docRef)
		if err != nil {
			if status.Code(err) == codes.NotFound {
				// document doesn't yet exist, so just write it
				attemptedWrite = true
				return t.Set(docRef, proposedPayment)
			}
			// document exists but there was some error, bail
			slog.ErrorContext(ctx, err.Error(), "event", e)
			return err
		}

		if paymentCreateRequest {
			// if we get a create request and the doc already exists via an out of order event, just squelch it
			// TODO: test this case
			return nil
		}

		// since the document already exists and we have an update event, let's make sure
		// we really should update it
		persistedPayment := &paymentType.Payment{}
		if err := docSnap.DataTo(persistedPayment); err != nil {
			slog.ErrorContext(ctx, err.Error(), "event", e)
			return err
		}
		// search the map to see if we've observed the idempotency key before
		if _, ok := persistedPayment.IdempotencyKeys[idempotencyKey]; ok {
			// we've already processed this update from square, so ignore it
			slog.DebugContext(ctx, "skipped duplicate event seen from Square", "idempotencyKey", idempotencyKey)
			return nil
		}

		// check to see if this square event is out of order
		if persistedPayment.SquareUpdatedTime.After(proposedPayment.SquareUpdatedTime) {
			// we've already processed a newer update from square, so ignore it
			slog.DebugContext(ctx, "skipped out of order event seen from Square", "idempotencyKey", idempotencyKey)
			return nil
		}

		// copy over idempotency keys from what we've seen before
		for key, val := range persistedPayment.IdempotencyKeys {
			proposedPayment.IdempotencyKeys[key] = val
		}

		// if we get here, we have a newer proposal for payment so let's write it
		attemptedWrite = true
		return t.Set(docRef, proposedPayment)
	}

	if err := firestoreClient.RunTransaction(ctx, transaction); err != nil {
		slog.ErrorContext(ctx, err.Error(), "event", e)
		return err
	}

	// if we got here and attemptedWrite is true, then we wrote the document successfully
	if attemptedWrite {
		slog.InfoContext(ctx, fmt.Sprintf("payment %v written at %v", docRef.ID, docRef.Path))
	}
	return nil
}

// ProcessCDCEvent generates internal domain events from changes to firestore payments collection
func ProcessCDCEvent(ctx context.Context, e event.Event) error {
	var data firestoredata.DocumentEventData
	if err := proto.Unmarshal(e.Data(), &data); err != nil {
		slog.ErrorContext(ctx, err.Error(), "event", e)
		return fmt.Errorf("proto.Unmarshal: %w", err)
	}

	var internalEvent *event.Event
	if data.GetValue() == nil {
		// the payment document was deleted
		payment := &paymentType.Payment{}
		err := util.ParseFirebaseDocument(data.OldValue, payment)
		if err != nil {
			slog.ErrorContext(ctx, err.Error(), "event", e)
			return err
		}
		internalEvent, err = eventschemas.NewPaymentDeleted(payment)
		if err != nil {
			slog.ErrorContext(ctx, err.Error(), "event", e)
			return err
		}
	} else if data.GetOldValue() == nil {
		// the payment document was created
		payment := &paymentType.Payment{}
		err := util.ParseFirebaseDocument(data.Value, payment)
		if err != nil {
			slog.ErrorContext(ctx, err.Error(), "event", e)
			return err
		}
		internalEvent, err = eventschemas.NewPaymentCreated(payment)
		if err != nil {
			slog.ErrorContext(ctx, err.Error(), "event", e)
			return err
		}
	} else {
		// the payment document was updated
		payment := &paymentType.Payment{}
		err := util.ParseFirebaseDocument(data.Value, payment)
		if err != nil {
			slog.ErrorContext(ctx, err.Error(), "event", e)
			return err
		}
		oldPayment := &paymentType.Payment{}
		if err = util.ParseFirebaseDocument(data.OldValue, oldPayment); err != nil {
			slog.ErrorContext(ctx, err.Error(), "event", e)
			return err
		}
		internalEvent, err = eventschemas.NewPaymentUpdated(oldPayment, payment, data.UpdateMask.FieldPaths)
		if err != nil {
			slog.ErrorContext(ctx, err.Error(), "event", e)
			return err
		}
	}

	eventJSON, err := internalEvent.MarshalJSON()
	if err != nil {
		slog.ErrorContext(ctx, err.Error(), "event", e)
		return err
	}
	timeoutContext, cancel := context.WithTimeout(context.Background(), PUBLISH_TIMEOUT_SEC)
	defer cancel()

	publishResult := paymentEventsTopic.Publish(timeoutContext, &pubsub.Message{Data: eventJSON})
	messageID, err := publishResult.Get(timeoutContext) // this call blocks until complete or timeout occurs
	if err != nil {
		slog.ErrorContext(ctx, err.Error(), "event", e)
		return err
	}
	slog.InfoContext(ctx, fmt.Sprintf("published %s", internalEvent.Type()), "messageID", messageID, "paymentID", internalEvent.Subject())
	return nil
}

// RefundWatcher updates the payment object based on observed refunds
func RefundWatcher(ctx context.Context, e event.Event) error {
	// there are two CloudEvents - one for the pubsub message "event", and then the data within
	var msg eventschemas.MessagePublishedData
	if err := e.DataAs(&msg); err != nil {
		slog.ErrorContext(ctx, err.Error(), "event", e)
		return err
	}

	nestedEvent := &event.Event{}
	if err := nestedEvent.UnmarshalJSON(msg.Message.Data); err != nil {
		slog.ErrorContext(ctx, err.Error(), "event", e)
		return err
	}

	var idempotencyKey string
	var refundToProcess *refundType.Refund

	switch nestedEvent.Type() {
	case eventschemas.RefundCreatedType:
		rc := &eventschemas.RefundCreated{}
		if err := nestedEvent.DataAs(rc); err != nil {
			slog.ErrorContext(ctx, err.Error(), "event", e)
			return err
		}
		idempotencyKey = rc.IdempotencyKey
		refundToProcess = rc.Refund
	case eventschemas.RefundUpdatedType:
		ru := &eventschemas.RefundUpdated{}
		if err := nestedEvent.DataAs(ru); err != nil {
			slog.ErrorContext(ctx, err.Error(), "event", e)
			return err
		}
		idempotencyKey = ru.IdempotencyKey
		refundToProcess = ru.Refund
	case eventschemas.RefundDeletedType:
		rd := &eventschemas.RefundDeleted{}
		if err := nestedEvent.DataAs(rd); err != nil {
			slog.ErrorContext(ctx, err.Error(), "event", e)
			return err
		}
		idempotencyKey = rd.IdempotencyKey
		refundToProcess = rd.Refund
	default:
		slog.DebugContext(ctx, fmt.Sprintf("squelching %q event", e.Type()), "event", nestedEvent.String())
		return nil
	}

	// if we have a new refund, find the matching internal Payment object
	paymentDocRef := firestoreClient.Doc(fmt.Sprintf("%s/%s", paymentDocPath, refundToProcess.SquarePaymentID))
	transaction := func(ctx context.Context, t *firestore.Transaction) error {
		paymentDocSnap, err := t.Get(paymentDocRef)
		if err != nil {
			if status.Code(err) == codes.NotFound {
				// payment object doesn't yet exist, so just write it
				getPaymentEvent := eventschemas.NewSquareGetPaymentRequest(refundToProcess.SquarePaymentID)
				eventJSON, err := getPaymentEvent.MarshalJSON()
				if err != nil {
					slog.ErrorContext(ctx, err.Error(), "event", e)
					return err
				}
				timeoutContext, cancel := context.WithTimeout(context.Background(), PUBLISH_TIMEOUT_SEC)
				defer cancel()

				publishResult := squarePaymentRequestTopic.Publish(timeoutContext, &pubsub.Message{Data: eventJSON})
				messageID, err := publishResult.Get(timeoutContext) // this call blocks until complete or timeout occurs
				if err != nil {
					slog.ErrorContext(ctx, err.Error(), "event", e)
					return err
				}
				slog.InfoContext(ctx, "published SquareGetPaymentRequest during refund processing", "messageID", messageID, "paymentID", refundToProcess.SquarePaymentID)
				return nil
			}
			// document exists but there was some database error, bail
			slog.ErrorContext(ctx, err.Error(), "event", e)
			return err
		}
		// payment object exists, let's denote the fact we have a related refund
		persistedPayment := &paymentType.Payment{}
		if err := paymentDocSnap.DataTo(persistedPayment); err != nil {
			slog.ErrorContext(ctx, err.Error(), "event", e)
			return err
		}

		// search the map to see if we've observed the idempotency key before (i.e. processed this event before)
		if _, ok := persistedPayment.IdempotencyKeys[idempotencyKey]; ok {
			// we've already processed this, so ignore the duplicate
			slog.DebugContext(ctx, "already processed update for this refund", "idempotencyKey", idempotencyKey, "refundID", refundToProcess.ID, "paymentID", persistedPayment.ID)
			return nil
		}

		// let's update the map with this event
		persistedPayment.IdempotencyKeys[idempotencyKey] = true

		switch nestedEvent.Type() {
		case eventschemas.RefundCreatedType:
			switch refundToProcess.Status {
			case refundType.REFUND_STATUS_PENDING, refundType.REFUND_STATUS_COMPLETED:
				persistedPayment.RefundAmount += refundToProcess.RefundAmount
				persistedPayment.FeeAmount -= refundToProcess.FeeAmount
				persistedPayment.SquareRefundIDs = append(persistedPayment.SquareRefundIDs, refundToProcess.ID)
			default:
				slog.DebugContext(ctx, fmt.Sprintf("ignoring refund create with %q status", refundToProcess.Status), "event", nestedEvent.String())
				// fall through to write idempotencyKey update
			}
		case eventschemas.RefundUpdatedType:
			switch refundToProcess.Status {
			case refundType.REFUND_STATUS_PENDING, refundType.REFUND_STATUS_COMPLETED:
				if slices.Contains[[]string, string](persistedPayment.SquareRefundIDs, refundToProcess.ID) {
					slog.DebugContext(ctx, "already processed refund against payment", "event", nestedEvent)
				} else {
					persistedPayment.RefundAmount += refundToProcess.RefundAmount
					persistedPayment.FeeAmount -= refundToProcess.FeeAmount
					persistedPayment.SquareRefundIDs = append(persistedPayment.SquareRefundIDs, refundToProcess.ID)
				}
			case refundType.REFUND_STATUS_FAILED:
				// this can happen if there is zero Square balance, and the withdrawal fails for some reason
				if slices.Contains[[]string, string](persistedPayment.SquareRefundIDs, refundToProcess.ID) {
					persistedPayment.RefundAmount -= refundToProcess.RefundAmount
					persistedPayment.FeeAmount += refundToProcess.FeeAmount
				}
			default:
				slog.DebugContext(ctx, fmt.Sprintf("ignoring refund update with %q status", refundToProcess.Status), "event", nestedEvent.String())
				// fall through to write idempotencyKey update
			}
		case eventschemas.RefundDeletedType:
			if slices.Contains[[]string, string](persistedPayment.SquareRefundIDs, refundToProcess.ID) {
				persistedPayment.RefundAmount -= refundToProcess.RefundAmount
				persistedPayment.FeeAmount += refundToProcess.FeeAmount
			} else {
				slog.DebugContext(ctx, "ignoring refund delete", "event", nestedEvent.String())
				// fall through to write idempotencyKey update
			}
		}

		if err := t.Set(paymentDocRef, persistedPayment); err != nil {
			slog.ErrorContext(ctx, err.Error(), "event", e)
			return err
		}

		slog.DebugContext(ctx, "updated payment for refund change", "event", nestedEvent.String())
		return nil
	}

	return firestoreClient.RunTransaction(ctx, transaction)
}

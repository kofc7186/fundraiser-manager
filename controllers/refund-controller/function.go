package refundcontroller

import (
	"context"
	"fmt"
	"log/slog"
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
	refundtype "github.com/kofc7186/fundraiser-manager/pkg/types/refund"
	"github.com/kofc7186/fundraiser-manager/pkg/util"

	_ "github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
)

const (
	FUNCTION_NAME       = "refund-controller"
	PUBLISH_TIMEOUT_SEC = 2 * time.Second
)

var firestoreClient *firestore.Client
var refundDocPath string

var refundEventsTopic *pubsub.Topic

var expirationTime time.Time

func init() {
	slog.SetDefault(logging.FunctionLogger(FUNCTION_NAME))

	psClient, err := pubsub.NewClient(context.Background(), util.GetEnvOrPanic("GCP_PROJECT"))
	if err != nil {
		panic(err)
	}

	REFUND_EVENTS_TOPIC := util.GetEnvOrPanic("REFUND_EVENTS_TOPIC")
	refundEventsTopic = psClient.Topic(REFUND_EVENTS_TOPIC)
	if ok, err := refundEventsTopic.Exists(context.Background()); !ok || err != nil {
		panic(fmt.Sprintf("existence check for %s failed: %v", REFUND_EVENTS_TOPIC, err))
	}

	firestoreClient, err = firestore.NewClient(context.Background(), util.GetEnvOrPanic("GCP_PROJECT"))
	if err != nil {
		panic(err)
	}

	refundDocPath = fmt.Sprintf("fundraisers/%s/refunds", util.GetEnvOrPanic("FUNDRAISER_ID"))

	expirationTime, err = time.Parse(time.RFC3339, util.GetEnvOrPanic("EXPIRATION_TIME"))
	if err != nil {
		panic(err)
	}

	// do this last so we are ensured to have all the required clients established above
	functions.CloudEvent("ProcessSquareRefundWebhookEvent", ProcessSquareRefundWebhookEvent)
	functions.CloudEvent("ProcessCDCEvent", ProcessCDCEvent)
}

// ProcessSquareRefundWebhookEvent
func ProcessSquareRefundWebhookEvent(ctx context.Context, e event.Event) error {
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

	return writeSquareRefundToFirestore(ctx, nestedEvent)
}

func writeSquareRefundToFirestore(ctx context.Context, e *event.Event) error {
	refundCreateRequest := false
	attemptedWrite := false

	var idempotencyKey string
	var proposedRefund *refundtype.Refund
	switch e.Type() {
	case eventschemas.RefundCreatedFromSquareType:
		refundCreateRequest = true

		rr := &eventschemas.RefundCreatedFromSquare{}
		if err := e.DataAs(rr); err != nil {
			return err
		}
		idempotencyKey = rr.BaseRefund.IdempotencyKey
		proposedRefund = rr.BaseRefund.Refund
	case eventschemas.RefundUpdatedFromSquareType:
		ru := &eventschemas.RefundUpdatedFromSquare{}
		if err := e.DataAs(ru); err != nil {
			return err
		}
		idempotencyKey = ru.BaseRefund.IdempotencyKey
		proposedRefund = ru.BaseRefund.Refund
	default:
		// TODO: slog
		return nil
	}

	if proposedRefund.Unlinked {
		slog.InfoContext(ctx, "received info about an unlinked refund, squelching", "event", e)
		return nil
	}

	// make sure to update the map to denote that we've processed this event already
	//
	// the boolean here is only to allow Firestore to map back to Go struct; the important
	// thing is that the key is put into the map
	proposedRefund.IdempotencyKeys = make(map[string]bool, 1)
	if idempotencyKey != "" {
		proposedRefund.IdempotencyKeys[idempotencyKey] = true
	}

	// ensure the firestore expiration timestamp is written in the appropriate field
	proposedRefund.Expiration = expirationTime

	docRef := firestoreClient.Doc(fmt.Sprintf("%s/%s", refundDocPath, proposedRefund.ID))
	transaction := func(ctx context.Context, t *firestore.Transaction) error {
		docSnap, err := t.Get(docRef)
		if err != nil {
			if status.Code(err) == codes.NotFound {
				// document doesn't yet exist, so just write it
				attemptedWrite = true
				return t.Set(docRef, proposedRefund)
			}
			// document exists but there was some error, bail
			return err
		}

		if refundCreateRequest {
			// if we get a create request and the doc already exists via an out of order event, just squelch it
			// TODO: test this case
			return nil
		}

		// since the document already exists and we have an update event, let's make sure
		// we really should update it
		persistedRefund := &refundtype.Refund{}
		if err := docSnap.DataTo(persistedRefund); err != nil {
			return err
		}
		// search the map to see if we've observed the idempotency key before
		if _, ok := persistedRefund.IdempotencyKeys[idempotencyKey]; ok {
			// we've already processed this update from square, so ignore it
			slog.DebugContext(ctx, "skipped duplicate event seen from Square", "idempotencyKey", idempotencyKey)
			return nil
		}

		// check to see if this square event is out of order
		if persistedRefund.SquareUpdatedTime.After(proposedRefund.SquareUpdatedTime) {
			// we've already processed a newer update from square, so ignore it
			slog.DebugContext(ctx, "skipped out of order event seen from Square", "idempotencyKey", idempotencyKey)
			return nil
		}

		// copy over idempotency keys from what we've seen before
		for key, val := range persistedRefund.IdempotencyKeys {
			proposedRefund.IdempotencyKeys[key] = val
		}

		// if we get here, we have a newer proposal for payment so let's write it
		attemptedWrite = true
		return t.Set(docRef, proposedRefund)
	}

	if err := firestoreClient.RunTransaction(ctx, transaction); err != nil {
		return err
	}

	// if we got here and attemptedWrite is true, then we wrote the document successfully
	if attemptedWrite {
		slog.InfoContext(ctx, fmt.Sprintf("refund %v written at %v", docRef.ID, docRef.Path))
	}
	return nil
}

// ProcessCDCEvent generates internal domain events from changes to firestore refunds collection
func ProcessCDCEvent(ctx context.Context, e event.Event) error {
	var data firestoredata.DocumentEventData
	if err := proto.Unmarshal(e.Data(), &data); err != nil {
		return fmt.Errorf("proto.Unmarshal: %w", err)
	}

	var internalEvent *event.Event
	if data.GetValue() == nil {
		// the refund document was deleted
		refund := &refundtype.Refund{}
		err := util.ParseFirebaseDocument(data.OldValue, refund)
		if err != nil {
			return err
		}
		internalEvent, err = eventschemas.NewRefundDeleted(refund)
		if err != nil {
			return err
		}
	} else if data.GetOldValue() == nil {
		// the payment document was created
		refund := &refundtype.Refund{}
		err := util.ParseFirebaseDocument(data.Value, refund)
		if err != nil {
			return err
		}
		internalEvent, err = eventschemas.NewRefundCreated(refund)
		if err != nil {
			return err
		}
	} else {
		// the payment document was updated
		refund := &refundtype.Refund{}
		err := util.ParseFirebaseDocument(data.Value, refund)
		if err != nil {
			return err
		}
		oldRefund := &refundtype.Refund{}
		if err = util.ParseFirebaseDocument(data.OldValue, oldRefund); err != nil {
			return err
		}
		internalEvent, err = eventschemas.NewRefundUpdated(oldRefund, refund, data.UpdateMask.FieldPaths)
		if err != nil {
			return err
		}
	}

	eventJSON, err := internalEvent.MarshalJSON()
	if err != nil {
		return err
	}
	timeoutContext, cancel := context.WithTimeout(context.Background(), PUBLISH_TIMEOUT_SEC)
	defer cancel()

	publishResult := refundEventsTopic.Publish(timeoutContext, &pubsub.Message{Data: eventJSON})
	messageID, err := publishResult.Get(timeoutContext) // this call blocks until complete or timeout occurs
	if err != nil {
		return err
	}
	slog.InfoContext(ctx, fmt.Sprintf("published %s", internalEvent.Type()), "messageID", messageID, "refundID", internalEvent.Subject())
	return nil
}

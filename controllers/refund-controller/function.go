package refundcontroller

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/cloudevents/sdk-go/v2/event"

	eventschemas "github.com/kofc7186/fundraiser-manager/pkg/event-schemas"
	"github.com/kofc7186/fundraiser-manager/pkg/logging"
	"github.com/kofc7186/fundraiser-manager/pkg/types"
	"github.com/kofc7186/fundraiser-manager/pkg/util"

	_ "github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
)

const PUBLISH_TIMEOUT_SEC = 2 * time.Second

var firestoreClient *firestore.Client
var refundDocPath string

var expirationTime time.Time

func init() {
	slog.SetDefault(logging.Logger)

	var err error
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
	functions.CloudEvent("RefundEvent", ProcessRefundEvent)
}

// ProcessRefundEvent
func ProcessRefundEvent(ctx context.Context, e event.Event) error {
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

	return writeSquareRefundToFirestore(ctx, nestedEvent)
}

func writeSquareRefundToFirestore(ctx context.Context, e *event.Event) error {
	refundCreateRequest := false

	var idempotencyKey string
	var proposedRefund *types.Refund
	switch e.Type() {
	case eventschemas.RefundReceivedType:
		refundCreateRequest = true

		rr := &eventschemas.RefundReceived{}
		if err := e.DataAs(rr); err != nil {
			return err
		}
		idempotencyKey = rr.BaseRefund.IdempotencyKey
		proposedRefund = rr.BaseRefund.Refund
	case eventschemas.RefundUpdatedType:
		ru := &eventschemas.RefundUpdated{}
		if err := e.DataAs(ru); err != nil {
			return err
		}
		idempotencyKey = ru.BaseRefund.IdempotencyKey
		proposedRefund = ru.BaseRefund.Refund
	}

	// make sure to update the map to denote that we've processed this event already
	//
	// the boolean here is only to allow Firestore to map back to Go struct; the important
	// thing is that the key is put into the map
	proposedRefund.IdempotencyKeys = make(map[string]bool, 1)
	proposedRefund.IdempotencyKeys[idempotencyKey] = true

	// ensure the firestore expiration timestamp is written in the appropriate field
	proposedRefund.Expiration = expirationTime

	docRef := firestoreClient.Doc(fmt.Sprintf("%s/%s", refundDocPath, proposedRefund.ID))
	transaction := func(ctx context.Context, t *firestore.Transaction) error {
		docSnap, err := t.Get(docRef)
		if err != nil {
			if status.Code(err) == codes.NotFound {
				// document doesn't yet exist, so just write it
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
		persistedRefund := &types.Refund{}
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
		return t.Set(docRef, proposedRefund)
	}

	if err := firestoreClient.RunTransaction(ctx, transaction); err != nil {
		return err
	}

	slog.InfoContext(ctx, fmt.Sprintf("refund %v written at %v", docRef.ID, docRef.Path))
	return nil
}

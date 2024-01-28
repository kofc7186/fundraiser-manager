package customercontroller

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
var customerDocPath string

var expirationTime time.Time

func init() {
	slog.SetDefault(logging.Logger)

	var err error
	firestoreClient, err = firestore.NewClient(context.Background(), util.GetEnvOrPanic("GCP_PROJECT"))
	if err != nil {
		panic(err)
	}

	customerDocPath = fmt.Sprintf("fundraisers/%s/customers", util.GetEnvOrPanic("FUNDRAISER_ID"))

	expirationTime, err = time.Parse(time.RFC3339, util.GetEnvOrPanic("EXPIRATION_TIME"))
	if err != nil {
		panic(err)
	}

	// do this last so we are ensured to have all the required clients established above
	functions.CloudEvent("CustomerEvent", ProcessCustomerEvent)
}

// ProcessCustomerEvent
func ProcessCustomerEvent(ctx context.Context, e event.Event) error {
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

	return writeSquareCustomerToFirestore(ctx, nestedEvent)
}

func writeSquareCustomerToFirestore(ctx context.Context, e *event.Event) error {
	customerCreateRequest := false

	var idempotencyKey string
	var proposedCustomer *types.Customer
	switch e.Type() {
	case eventschemas.CustomerReceivedType:
		customerCreateRequest = true

		cr := &eventschemas.CustomerReceived{}
		if err := e.DataAs(cr); err != nil {
			return err
		}
		idempotencyKey = cr.BaseCustomer.IdempotencyKey
		proposedCustomer = cr.BaseCustomer.Customer
	case eventschemas.CustomerUpdatedType:
		cu := &eventschemas.CustomerUpdated{}
		if err := e.DataAs(cu); err != nil {
			return err
		}
		idempotencyKey = cu.BaseCustomer.IdempotencyKey
		proposedCustomer = cu.BaseCustomer.Customer
	}

	docRef := firestoreClient.Doc(fmt.Sprintf("%s/%s", customerDocPath, proposedCustomer.ID))
	transaction := func(ctx context.Context, t *firestore.Transaction) error {
		docSnap, err := t.Get(docRef)
		if err != nil {
			if status.Code(err) == codes.NotFound {
				// document doesn't yet exist, so just write it
				return t.Set(docRef, proposedCustomer)
			}
			// document exists but there was some error, bail
			return err
		}

		if customerCreateRequest {
			// if we get a create request and the doc already exists via an out of order event, just squelch it
			// TODO: test this case
			return nil
		}

		// since the document already exists and we have an update event, let's make sure
		// we really should update it
		persistedcustomer := &types.Customer{}
		if err := docSnap.DataTo(persistedcustomer); err != nil {
			return err
		}
		// search the map to see if we've observed the idempotency key before
		if _, ok := persistedcustomer.IdempotencyKeys[idempotencyKey]; ok {
			// we've already processed this update from square, so ignore it
			slog.DebugContext(ctx, "skipped duplicate event seen from Square", "idempotencyKey", idempotencyKey)
			return nil
		}

		// check to see if this square event is out of order
		if persistedcustomer.SquareUpdatedTime.After(proposedCustomer.SquareUpdatedTime) {
			// we've already processed a newer update from square, so ignore it
			slog.DebugContext(ctx, "skipped out of order event seen from Square", "idempotencyKey", idempotencyKey)
			return nil
		}

		// make sure to update the map to denote that we've processed this event already
		//
		// the boolean here is only to allow Firestore to map back to Go struct; the important
		// thing is that the key is put into the map
		proposedCustomer.IdempotencyKeys[idempotencyKey] = true

		// ensure the firestore expiration timestamp is written in the appropriate field
		proposedCustomer.Expiration = expirationTime

		// if we get here, we have a newer proposal for payment so let's write it
		return t.Set(docRef, proposedCustomer)
	}

	if err := firestoreClient.RunTransaction(ctx, transaction); err != nil {
		return err
	}

	slog.InfoContext(ctx, fmt.Sprintf("customer %v written at %v", docRef.ID, docRef.Path))
	return nil
}

// TODO: have a trigger on order create/update to query customer table, and fetch it if we don't find it

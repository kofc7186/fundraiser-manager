package ordercontroller

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/cloudevents/sdk-go/v2/event"

	"github.com/googleapis/google-cloudevents-go/cloud/firestoredata"

	eventschemas "github.com/kofc7186/fundraiser-manager/pkg/event-schemas"
	"github.com/kofc7186/fundraiser-manager/pkg/logging"
	"github.com/kofc7186/fundraiser-manager/pkg/types"
	"github.com/kofc7186/fundraiser-manager/pkg/util"

	_ "github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
)

const PUBLISH_TIMEOUT_SEC = 2 * time.Second

var firestoreClient *firestore.Client
var orderDocPath string

var expirationTime time.Time

func init() {
	slog.SetDefault(logging.Logger)

	var err error
	firestoreClient, err = firestore.NewClient(context.Background(), util.GetEnvOrPanic("GCP_PROJECT"))
	if err != nil {
		panic(err)
	}

	orderDocPath = fmt.Sprintf("fundraisers/%s/orders", util.GetEnvOrPanic("FUNDRAISER_ID"))

	expirationTime, err = time.Parse(time.RFC3339, util.GetEnvOrPanic("EXPIRATION_TIME"))
	if err != nil {
		panic(err)
	}

	// do this last so we are ensured to have all the required clients established above
	functions.CloudEvent("OrderEvent", ProcessOrderEvent)
	functions.CloudEvent("CustomerWatcher", CustomerWatcher)
	functions.CloudEvent("ProcessSquareOrderResponse", ProcessOrderEvent) // Square API responses are how we process webhooks
}

// ProcessOrderEvent
func ProcessOrderEvent(ctx context.Context, e event.Event) error {
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

	return writeSquareOrderToFirestore(ctx, nestedEvent)
}

func writeSquareOrderToFirestore(ctx context.Context, e *event.Event) error {
	orderCreateRequest := false

	var idempotencyKey string
	var proposedOrder *types.Order
	switch e.Type() {
	case eventschemas.SquareGetOrderCompletedType:
		sgoc := &eventschemas.SquareGetOrderCompleted{}
		if err := e.DataAs(sgoc); err != nil {
			return err
		}
		idempotencyKey = sgoc.BaseOrder.IdempotencyKey
		proposedOrder = sgoc.BaseOrder.Order
	}

	// make sure to update the map to denote that we've processed this event already
	//
	// the boolean here is only to allow Firestore to map back to Go struct; the important
	// thing is that the key is put into the map
	proposedOrder.IdempotencyKeys = make(map[string]bool, 1)
	if idempotencyKey != "" {
		proposedOrder.IdempotencyKeys[idempotencyKey] = true
	}

	// ensure the firestore expiration timestamp is written in the appropriate field
	proposedOrder.Expiration = expirationTime

	docRef := firestoreClient.Doc(fmt.Sprintf("%s/%s", orderDocPath, proposedOrder.ID))
	transaction := func(ctx context.Context, t *firestore.Transaction) error {
		docSnap, err := t.Get(docRef)
		if err != nil {
			if status.Code(err) == codes.NotFound {
				// document doesn't yet exist, so just write it
				return t.Set(docRef, proposedOrder)
			}
			// document exists but there was some error, bail
			return err
		}

		if orderCreateRequest {
			// if we get a create request and the doc already exists via an out of order event, just squelch it
			// TODO: test this case
			return nil
		}

		// since the document already exists and we have an update event, let's make sure
		// we really should update it
		persistedOrder := &types.Order{}
		if err := docSnap.DataTo(persistedOrder); err != nil {
			return err
		}
		// search the map to see if we've observed the idempotency key before
		if _, ok := persistedOrder.IdempotencyKeys[idempotencyKey]; ok {
			// we've already processed this update from square, so ignore it
			slog.DebugContext(ctx, "skipped duplicate event seen from Square", "idempotencyKey", idempotencyKey)
			return nil
		}

		// check to see if this square event is out of order
		if persistedOrder.SquareUpdatedTime.After(proposedOrder.SquareUpdatedTime) {
			// we've already processed a newer update from square, so ignore it
			slog.DebugContext(ctx, "skipped out of order event seen from Square", "idempotencyKey", idempotencyKey)
			return nil
		}

		// copy over idempotency keys from what we've seen before
		for key, val := range persistedOrder.IdempotencyKeys {
			proposedOrder.IdempotencyKeys[key] = val
		}

		// if we get here, we have a newer proposal for customer so let's write it
		return t.Set(docRef, proposedOrder)
	}

	if err := firestoreClient.RunTransaction(ctx, transaction); err != nil {
		return err
	}

	slog.InfoContext(ctx, fmt.Sprintf("order %v written at %v", docRef.ID, docRef.Path))
	return nil
}

// CustomerWatcher is triggered on any create/update/delete events on an Customer object
// It should scan all orders we know of, and if any are for that customer, we should attempt
// to update the relevant fields (first/last/display name, email, phone number, knight)
func CustomerWatcher(ctx context.Context, e event.Event) error {
	var data firestoredata.DocumentEventData
	if err := proto.Unmarshal(e.Data(), &data); err != nil {
		return fmt.Errorf("proto.Unmarshal: %w", err)
	}

	if data.GetValue() == nil {
		// the order document was deleted, nothing for us to do
		return nil
	}

	// we're here because an order was either just created or updated
	if customerID, ok := data.Value.Fields["id"]; ok {
		customerIDString := customerID.GetStringValue()
		if customerIDString == "" {
			// customerID is unset in the order, therefore we can't find a match in Firestore, so just bail
			return nil
		}
		docs := firestoreClient.Collection(orderDocPath).Where("squareCustomerID", "==", customerIDString)
		return firestoreClient.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
			docSnaps, err := tx.Documents(docs).GetAll()
			if err != nil {
				return err
			}
			for _, docSnap := range docSnaps {
				// if we're here, we have updated customer information for a valid order
				order := types.Order{}
				if err := docSnap.DataTo(&order); err != nil {
					slog.ErrorContext(ctx, err.Error())
					continue
				}

				updates := []firestore.Update{}
				// default to a create event, where we'd want to try to update all fields
				updateFields := []string{"emailAddress", "phoneNumber", "lastName", "firstName", "isKnight"}
				if data.UpdateMask != nil {
					// this is an update event, so only touch the fields that have changed in the customer object
					updateFields = data.UpdateMask.FieldPaths
				}
				for _, field := range updateFields {
					switch field {
					case "emailAddress", "phoneNumber":
						if order.EmailAddress == "" || order.PhoneNumber == "" {
							updates = append(updates, firestore.Update{
								Path:  field,
								Value: data.Value.Fields[field].GetStringValue(),
							})
						}
					case "lastName":
						if order.LastName == "" {
							updates = append(updates, firestore.Update{
								Path:  field,
								Value: data.Value.Fields[field].GetStringValue(),
							})
						}
						// TODO: set display name
					case "firstName":
						if order.FirstName == "" {
							updates = append(updates, firestore.Update{
								Path:  field,
								Value: data.Value.Fields[field].GetStringValue(),
							})
						}
						// TODO: set display name
					case "isKnight":
						updates = append(updates, firestore.Update{
							Path:  field,
							Value: data.Value.Fields[field].GetBooleanValue(),
						})
					}
				}

				// update order with Customer-sourced information
				wr, err := docSnap.Ref.Update(ctx, updates)
				if err != nil {
					slog.ErrorContext(ctx, "failed to update order with new customer info", "error", err)
					continue // we quietly continue here so as to not fail the entire txn
				}
				slog.DebugContext(ctx, "updated order with new customer info", "orderID", order.ID, "updateTime", wr.UpdateTime)
			}
			return nil
		})
	}
	return nil
}

package ordercontroller

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/pubsub"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/cloudevents/sdk-go/v2/event"

	"github.com/googleapis/google-cloudevents-go/cloud/firestoredata"
	eventschemas "github.com/kofc7186/fundraiser-manager/pkg/event-schemas"
	"github.com/kofc7186/fundraiser-manager/pkg/logging"
	squarewebhooktypes "github.com/kofc7186/fundraiser-manager/pkg/square/types/webhooks"
	"github.com/kofc7186/fundraiser-manager/pkg/types"
	"github.com/kofc7186/fundraiser-manager/pkg/util"

	_ "github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
)

const PUBLISH_TIMEOUT_SEC = 2 * time.Second

var firestoreClient *firestore.Client
var fundraiserDocPath string
var orderDocPath string

var orderEventsTopic *pubsub.Topic

var expirationTime time.Time

func init() {
	slog.SetDefault(logging.Logger)

	psClient, err := pubsub.NewClient(context.Background(), util.GetEnvOrPanic("GCP_PROJECT"))
	if err != nil {
		panic(err)
	}

	firestoreClient, err = firestore.NewClient(context.Background(), util.GetEnvOrPanic("GCP_PROJECT"))
	if err != nil {
		panic(err)
	}

	ORDER_EVENTS_TOPIC := util.GetEnvOrPanic("ORDER_EVENTS_TOPIC")
	orderEventsTopic = psClient.Topic(ORDER_EVENTS_TOPIC)
	if ok, err := orderEventsTopic.Exists(context.Background()); !ok || err != nil {
		panic(fmt.Sprintf("existence check for %s failed: %v", ORDER_EVENTS_TOPIC, err))
	}

	fundraiserDocPath = fmt.Sprintf("fundraisers/%s", util.GetEnvOrPanic("FUNDRAISER_ID"))
	orderDocPath = fmt.Sprintf("%s/orders", fundraiserDocPath)

	expirationTime, err = time.Parse(time.RFC3339, util.GetEnvOrPanic("EXPIRATION_TIME"))
	if err != nil {
		panic(err)
	}

	// do this last so we are ensured to have all the required clients established above
	functions.CloudEvent("ProcessSquareRetrieveOrderResponse", ProcessSquareRetrieveOrderResponse) // Square API responses are how we process order.created/order.updated webhooks
	functions.CloudEvent("ProcessCDCEvent", ProcessCDCEvent)
	functions.CloudEvent("CustomerWatcher", CustomerWatcher)
	functions.CloudEvent("PaymentWatcher", PaymentWatcher)
}

// ProcessOrderEvent
func ProcessSquareRetrieveOrderResponse(ctx context.Context, e event.Event) error {
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

	if e.Source() == squarewebhooktypes.SQUARE_WEBHOOK_ORDER_CREATED {
		orderCreateRequest = true
	}

	var idempotencyKey string
	var proposedOrder types.Order
	switch e.Type() {
	case eventschemas.SquareRetrieveOrderResponseType:
		sgoc := &eventschemas.SquareRetrieveOrderResponse{}
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
	transaction := func(ctx context.Context, tx *firestore.Transaction) error {
		var orderNumber uint16 = 1000
		docSnap, err := tx.Get(docRef)
		if err != nil {
			if status.Code(err) == codes.NotFound {
				// order document doesn't yet exist, so just write it
				fundraiserDocRef := firestoreClient.Doc(fundraiserDocPath)
				fundraiserDocSnap, err := tx.Get(fundraiserDocRef)
				if err != nil {
					if status.Code(err) == codes.NotFound {
						// this is extremely unlikely, but create it if it doesn't exist
						if err := tx.Update(fundraiserDocRef, []firestore.Update{{Path: "orderNumber", Value: orderNumber}}); err != nil {
							return err
						}
					}
				} else {
					orderNumberInterface, err := fundraiserDocSnap.DataAt("orderNumber")
					if err != nil {
						if status.Code(err) == codes.NotFound {
							if err := tx.Update(fundraiserDocRef, []firestore.Update{{Path: "orderNumber", Value: orderNumber}}); err != nil {
								return err
							}
						} else {
							return err
						}
					} else {
						if err := tx.Update(fundraiserDocRef, []firestore.Update{{Path: "orderNumber", Value: firestore.Increment(1)}}); err != nil {
							return err
						}
					}
					orderNumber = orderNumberInterface.(uint16) + 1
				}
				proposedOrder.Number = orderNumber
				return tx.Set(docRef, proposedOrder)
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
		return tx.Set(docRef, proposedOrder)
	}

	if err := firestoreClient.RunTransaction(ctx, transaction); err != nil {
		return err
	}

	slog.InfoContext(ctx, fmt.Sprintf("order %v written at %v", docRef.ID, docRef.Path))
	return nil
}

// ProcessCDCEvent generates internal domain events from changes to firestore orders collection
func ProcessCDCEvent(ctx context.Context, e event.Event) error {
	var data firestoredata.DocumentEventData
	if err := proto.Unmarshal(e.Data(), &data); err != nil {
		return fmt.Errorf("proto.Unmarshal: %w", err)
	}

	var internalEvent *event.Event
	if data.GetValue() == nil {
		// the order document was deleted
		b, err := json.Marshal(util.ParseFirebaseDocument(data.OldValue))
		if err != nil {
			return err
		}
		order := &types.Order{}
		if err := json.Unmarshal(b, order); err != nil {
			return err
		}
		internalEvent, err = eventschemas.NewOrderDeleted(order)
		if err != nil {
			return err
		}
	} else if data.GetOldValue() == nil {
		// the order document was created
		b, err := json.Marshal(util.ParseFirebaseDocument(data.Value))
		if err != nil {
			return err
		}
		order := &types.Order{}
		if err := json.Unmarshal(b, order); err != nil {
			return err
		}
		internalEvent, err = eventschemas.NewOrderCreated(order)
		if err != nil {
			return err
		}
	} else {
		// the order document was updated
		orderBytes, err := json.Marshal(util.ParseFirebaseDocument(data.Value))
		if err != nil {
			return err
		}
		order := &types.Order{}
		if err := json.Unmarshal(orderBytes, order); err != nil {
			return err
		}
		oldOrderBytes, err := json.Marshal(util.ParseFirebaseDocument(data.OldValue))
		if err != nil {
			return err
		}
		oldOrder := &types.Order{}
		if err := json.Unmarshal(oldOrderBytes, oldOrder); err != nil {
			return err
		}
		internalEvent, err = eventschemas.NewOrderUpdated(oldOrder, order, data.UpdateMask.FieldPaths)
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

	publishResult := orderEventsTopic.Publish(timeoutContext, &pubsub.Message{Data: eventJSON})
	messageID, err := publishResult.Get(timeoutContext) // this call blocks until complete or timeout occurs
	if err != nil {
		return err
	}
	slog.InfoContext(ctx, fmt.Sprintf("published %s", internalEvent.Type()), "messageID", messageID, "orderID", internalEvent.Subject())
	return nil
}

// PaymentWatcher updates relevant order objects based on observed payment events
func PaymentWatcher(ctx context.Context, e event.Event) error {
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

	var idempotencyKey string
	var paymentToProcess types.Payment
	var fieldsInPaymentUpdate []string

	switch nestedEvent.Type() {
	case eventschemas.PaymentCreatedType:
		pc := &eventschemas.PaymentCreated{}
		if err := nestedEvent.DataAs(pc); err != nil {
			return err
		}
		idempotencyKey = pc.IdempotencyKey
		paymentToProcess = pc.Payment
	case eventschemas.PaymentUpdatedType:
		pu := &eventschemas.PaymentUpdated{}
		if err := nestedEvent.DataAs(pu); err != nil {
			return err
		}
		idempotencyKey = pu.IdempotencyKey
		paymentToProcess = pu.Payment
		fieldsInPaymentUpdate = pu.UpdatedFields
	default:
		slog.DebugContext(ctx, fmt.Sprintf("squelching %q event", e.Type()), "event", nestedEvent.String())
		return nil
	}

	docs := firestoreClient.Collection(orderDocPath).Where("squarePaymentID", "==", paymentToProcess.ID)
	return firestoreClient.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		docSnaps, err := tx.Documents(docs).GetAll()
		if err != nil {
			return err
		}
		for _, docSnap := range docSnaps {
			// if we're here, we have updated payment information for a valid order
			order := types.Order{}
			if err := docSnap.DataTo(&order); err != nil {
				slog.ErrorContext(ctx, err.Error())
				continue
			}

			if _, ok := order.IdempotencyKeys[idempotencyKey]; ok {
				slog.DebugContext(ctx, "already processed update for this payment", "idempotencyKey", idempotencyKey, "orderID", order.ID, "paymentID", paymentToProcess.ID)
				continue
			}

			// add this change to the idempotencyKeys map
			if order.IdempotencyKeys == nil {
				order.IdempotencyKeys = make(map[string]bool)
			}
			order.IdempotencyKeys[idempotencyKey] = true
			updates := []firestore.Update{{
				Path:  "idempotencyKeys",
				Value: order.IdempotencyKeys,
			}}
			// default to a create event, where we'd want to try to update all relevant fields
			updateFields := []string{"feeAmount", "note", "squareCustomerID", "tipAmount", "totalAmount"}
			if len(fieldsInPaymentUpdate) > 0 {
				// this is an update event, so only touch the fields that have changed in the payment object
				updateFields = fieldsInPaymentUpdate
			}
			for _, field := range updateFields {
				switch field {
				case "feeAmount":
					updates = append(updates, firestore.Update{
						Path:  field,
						Value: paymentToProcess.FeeAmount,
					})
				case "note":
					if !strings.Contains(order.Note, paymentToProcess.Note) {
						var orderPrefix string
						if order.Note != "" {
							orderPrefix = order.Note + ", "
						}
						updates = append(updates, firestore.Update{
							Path:  field,
							Value: orderPrefix + paymentToProcess.Note,
						})
					}
				case "squareCustomerID":
					updates = append(updates, firestore.Update{
						Path:  field,
						Value: paymentToProcess.SquareCustomerID,
					})
				case "tipAmount":
					updates = append(updates, firestore.Update{
						Path:  field,
						Value: paymentToProcess.TipAmount,
					})
				case "totalAmount":
					updates = append(updates, firestore.Update{
						Path:  field,
						Value: paymentToProcess.TotalAmount,
					})
				}
			}

			// update order with Payment-sourced information
			if err := tx.Update(docSnap.Ref, updates); err != nil {
				slog.ErrorContext(ctx, "failed to update order with new customer info", "error", err)
				continue // we quietly continue here so as to not fail the entire txn
			}
			slog.DebugContext(ctx, "updated order with new customer info", "orderID", order.ID)
		}
		return nil
	})
}

// CustomerWatcher updates relevant order objects based on observed customer events
func CustomerWatcher(ctx context.Context, e event.Event) error {
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

	var idempotencyKey string
	var customerToProcess types.Customer
	var fieldsInCustomerUpdate []string

	switch nestedEvent.Type() {
	case eventschemas.CustomerCreatedType:
		rc := &eventschemas.CustomerCreated{}
		if err := nestedEvent.DataAs(rc); err != nil {
			return err
		}
		idempotencyKey = rc.IdempotencyKey
		customerToProcess = rc.Customer
	case eventschemas.CustomerUpdatedType:
		ru := &eventschemas.CustomerUpdated{}
		if err := nestedEvent.DataAs(ru); err != nil {
			return err
		}
		idempotencyKey = ru.IdempotencyKey
		customerToProcess = ru.Customer
		fieldsInCustomerUpdate = ru.UpdatedFields
	default:
		slog.DebugContext(ctx, fmt.Sprintf("squelching %q event", e.Type()), "event", nestedEvent.String())
		return nil
	}

	docs := firestoreClient.Collection(orderDocPath).Where("squareCustomerID", "==", customerToProcess.ID)
	return firestoreClient.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		orderSnaps, err := tx.Documents(docs).GetAll()
		if err != nil {
			return err
		}
		for _, orderSnap := range orderSnaps {
			// if we're here, we have updated customer information for a valid order
			order := types.Order{}
			if err := orderSnap.DataTo(&order); err != nil {
				slog.ErrorContext(ctx, err.Error())
				continue
			}

			if _, ok := order.IdempotencyKeys[idempotencyKey]; ok {
				slog.DebugContext(ctx, "already processed update for this order", "idempotencyKey", idempotencyKey)
				continue
			}
			if order.IdempotencyKeys == nil {
				order.IdempotencyKeys = make(map[string]bool)
			}
			order.IdempotencyKeys[idempotencyKey] = true

			// add this change to the idempotencyKeys map
			updates := []firestore.Update{{
				Path:  "idempotencyKeys",
				Value: order.IdempotencyKeys,
			}}
			// default to a create event, where we'd want to try to update all fields
			updateFields := []string{"emailAddress", "phoneNumber", "lastName", "firstName", "isKnight"}
			if len(fieldsInCustomerUpdate) > 0 {
				// this is an update event, so only touch the fields that have changed in the customer object
				updateFields = fieldsInCustomerUpdate
			}
			for _, field := range updateFields {
				switch field {
				case "emailAddress":
					updates = append(updates, firestore.Update{
						Path:  field,
						Value: customerToProcess.EmailAddress,
					})
				case "phoneNumber":
					updates = append(updates, firestore.Update{
						Path:  field,
						Value: customerToProcess.PhoneNumber,
					})
				case "lastName":
					updates = append(updates, firestore.Update{
						Path:  field,
						Value: customerToProcess.LastName,
					})
					if order.DisplayName == "" {
						updates = append(updates, firestore.Update{
							Path:  "displayName",
							Value: customerToProcess.FirstName + " " + customerToProcess.LastName,
						})
					}
				case "firstName":
					updates = append(updates, firestore.Update{
						Path:  field,
						Value: customerToProcess.FirstName,
					})
					if order.DisplayName == "" {
						updates = append(updates, firestore.Update{
							Path:  "displayName",
							Value: customerToProcess.FirstName + " " + customerToProcess.LastName,
						})
					}
				case "isKnight":
					updates = append(updates, firestore.Update{
						Path:  field,
						Value: customerToProcess.KnightOfColumbus,
					})
				}
			}

			// update order with Customer-sourced information
			if err := tx.Update(orderSnap.Ref, updates); err != nil {
				slog.ErrorContext(ctx, "failed to update order with new customer info", "error", err)
				continue // we quietly continue here so as to not fail the entire txn
			}
			slog.DebugContext(ctx, "updated order with new customer info", "orderID", order.ID, "updates", updates)
		}
		return nil
	})
}

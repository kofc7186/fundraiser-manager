package customercontroller

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/pubsub"
	"google.golang.org/api/iterator"
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
var customerDocPath string

var customerEventsTopic *pubsub.Topic
var squareCustomerRequestTopic *pubsub.Topic

var expirationTime time.Time

func init() {
	slog.SetDefault(logging.Logger)

	psClient, err := pubsub.NewClient(context.Background(), util.GetEnvOrPanic("GCP_PROJECT"))
	if err != nil {
		panic(err)
	}

	CUSTOMER_EVENTS_TOPIC := util.GetEnvOrPanic("CUSTOMER_EVENTS_TOPIC")
	customerEventsTopic = psClient.Topic(CUSTOMER_EVENTS_TOPIC)
	if ok, err := customerEventsTopic.Exists(context.Background()); !ok || err != nil {
		panic(fmt.Sprintf("existence check for %s failed: %v", CUSTOMER_EVENTS_TOPIC, err))
	}

	SQUARE_CUSTOMER_REQUEST_TOPIC := util.GetEnvOrPanic("SQUARE_CUSTOMER_REQUEST_TOPIC")
	squareCustomerRequestTopic = psClient.Topic(SQUARE_CUSTOMER_REQUEST_TOPIC)
	if ok, err := squareCustomerRequestTopic.Exists(context.Background()); !ok || err != nil {
		panic(fmt.Sprintf("existence check for %s failed: %v", SQUARE_CUSTOMER_REQUEST_TOPIC, err))
	}

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
	functions.CloudEvent("ProcessSquareCustomerWebhookEvent", ProcessSquareCustomerWebhookEvent)
	functions.CloudEvent("ProcessSquareCustomerResponse", ProcessSquareCustomerWebhookEvent) // Square API responses just get written like inbound webhooks
	functions.CloudEvent("ProcessCDCEvent", ProcessCDCEvent)
	functions.CloudEvent("OrderWatcher", OrderAndPaymentWatcher)
	functions.CloudEvent("PaymentWatcher", OrderAndPaymentWatcher)
}

// ProcessSquareCustomerWebhookEvent
func ProcessSquareCustomerWebhookEvent(ctx context.Context, e event.Event) error {
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
	var proposedCustomer types.Customer
	switch e.Type() {
	case eventschemas.CustomerCreatedFromSquareType:
		customerCreateRequest = true

		cr := &eventschemas.CustomerCreatedFromSquare{}
		if err := e.DataAs(cr); err != nil {
			return err
		}
		idempotencyKey = cr.BaseCustomer.IdempotencyKey
		proposedCustomer = cr.BaseCustomer.Customer
	case eventschemas.CustomerUpdatedFromSquareType:
		cu := &eventschemas.CustomerUpdatedFromSquare{}
		if err := e.DataAs(cu); err != nil {
			return err
		}
		idempotencyKey = cu.BaseCustomer.IdempotencyKey
		proposedCustomer = cu.BaseCustomer.Customer
	case eventschemas.SquareRetrieveCustomerResponseType:
		sgcc := &eventschemas.SquareRetrieveCustomerResponse{}
		if err := e.DataAs(sgcc); err != nil {
			return err
		}
		idempotencyKey = sgcc.BaseCustomer.IdempotencyKey
		proposedCustomer = sgcc.BaseCustomer.Customer
	default:
		return nil // TODO: is this really what we want to do?
	}

	// make sure to update the map to denote that we've processed this event already
	//
	// the boolean here is only to allow Firestore to map back to Go struct; the important
	// thing is that the key is put into the map
	proposedCustomer.IdempotencyKeys = make(map[string]bool, 1)
	if idempotencyKey != "" {
		proposedCustomer.IdempotencyKeys[idempotencyKey] = true
	}

	// ensure the firestore expiration timestamp is written in the appropriate field
	proposedCustomer.Expiration = expirationTime

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
		persistedCustomer := &types.Customer{}
		if err := docSnap.DataTo(persistedCustomer); err != nil {
			return err
		}
		// search the map to see if we've observed the idempotency key before
		if _, ok := persistedCustomer.IdempotencyKeys[idempotencyKey]; ok {
			// we've already processed this update from square, so ignore it
			slog.DebugContext(ctx, "skipped duplicate event seen", "idempotencyKey", idempotencyKey, "event", e)
			return nil
		}

		// check to see if this square event is out of order
		if persistedCustomer.SquareUpdatedTime.After(proposedCustomer.SquareUpdatedTime) {
			// we've already processed a newer update from square, so ignore it
			slog.DebugContext(ctx, "skipped out of order event seen from Square", "idempotencyKey", idempotencyKey, "event", e)
			return nil
		}

		// copy over idempotency keys from what we've seen before
		for key, val := range persistedCustomer.IdempotencyKeys {
			proposedCustomer.IdempotencyKeys[key] = val
		}

		// if we get here, we have a newer proposal for customer so let's write it
		return t.Set(docRef, proposedCustomer)
	}

	if err := firestoreClient.RunTransaction(ctx, transaction); err != nil {
		return err
	}

	slog.InfoContext(ctx, fmt.Sprintf("customer %v written at %v", docRef.ID, docRef.Path))
	return nil
}

// ProcessCDCEvent generates internal domain events from changes to firestore customers collection
func ProcessCDCEvent(ctx context.Context, e event.Event) error {
	var data firestoredata.DocumentEventData
	if err := proto.Unmarshal(e.Data(), &data); err != nil {
		return fmt.Errorf("proto.Unmarshal: %w", err)
	}

	var internalEvent *event.Event
	if data.GetValue() == nil {
		// the customer document was deleted
		b, err := json.Marshal(util.ParseFirebaseDocument(data.OldValue))
		if err != nil {
			return err
		}
		customer := &types.Customer{}
		if err := json.Unmarshal(b, customer); err != nil {
			return err
		}
		internalEvent, err = eventschemas.NewCustomerDeleted(customer)
		if err != nil {
			return err
		}
	} else if data.GetOldValue() == nil {
		// the customer document was created
		b, err := json.Marshal(util.ParseFirebaseDocument(data.Value))
		if err != nil {
			return err
		}
		customer := &types.Customer{}
		if err := json.Unmarshal(b, customer); err != nil {
			return err
		}
		internalEvent, err = eventschemas.NewCustomerCreated(customer)
		if err != nil {
			return err
		}
	} else {
		// the customer document was updated
		customerBytes, err := json.Marshal(util.ParseFirebaseDocument(data.Value))
		if err != nil {
			return err
		}
		customer := &types.Customer{}
		if err := json.Unmarshal(customerBytes, customer); err != nil {
			return err
		}
		oldCustomerBytes, err := json.Marshal(util.ParseFirebaseDocument(data.OldValue))
		if err != nil {
			return err
		}
		oldCustomer := &types.Customer{}
		if err := json.Unmarshal(oldCustomerBytes, oldCustomer); err != nil {
			return err
		}
		internalEvent, err = eventschemas.NewCustomerUpdated(oldCustomer, customer, data.UpdateMask.FieldPaths)
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

	publishResult := customerEventsTopic.Publish(timeoutContext, &pubsub.Message{Data: eventJSON})
	messageID, err := publishResult.Get(timeoutContext) // this call blocks until complete or timeout occurs
	if err != nil {
		return err
	}
	slog.InfoContext(ctx, fmt.Sprintf("published %s", internalEvent.Type()), "messageID", messageID, "customerID", internalEvent.Subject())
	return nil
}

// OrderAndPaymentWatcher listens to internal order/payment events; it should scan and see if they have a customer ID that we don't know about
// if we don't have it, we should fetch and create the event upon getting the value from Square
func OrderAndPaymentWatcher(ctx context.Context, e event.Event) error {
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

	var squareCustomerID string
	switch nestedEvent.Type() {
	case eventschemas.PaymentCreatedType:
		pc := &eventschemas.PaymentCreated{}
		if err := nestedEvent.DataAs(pc); err != nil {
			return err
		}
		squareCustomerID = pc.Payment.SquareCustomerID
	case eventschemas.PaymentUpdatedType:
		pu := &eventschemas.PaymentUpdated{}
		if err := nestedEvent.DataAs(pu); err != nil {
			return err
		}
		squareCustomerID = pu.Payment.SquareCustomerID
	case eventschemas.OrderCreatedType:
		oc := &eventschemas.OrderCreated{}
		if err := nestedEvent.DataAs(oc); err != nil {
			return err
		}
		squareCustomerID = oc.Order.SquareCustomerID
	case eventschemas.OrderUpdatedType:
		ou := &eventschemas.OrderUpdated{}
		if err := nestedEvent.DataAs(ou); err != nil {
			return err
		}
		squareCustomerID = ou.Order.SquareCustomerID
	default:
		slog.DebugContext(ctx, fmt.Sprintf("squelching %q event", e.Type()), "event", nestedEvent.String())
		return nil
	}

	if squareCustomerID == "" {
		slog.InfoContext(ctx, "cannot update customer when ID field is blank", "event", nestedEvent.String())
		return nil
	}

	customerRecordRef := firestoreClient.Collection(customerDocPath).Where("id", "==", squareCustomerID)
	return firestoreClient.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		customerIterator := tx.Documents(customerRecordRef)
		defer customerIterator.Stop()
		if _, err := customerIterator.Next(); err == iterator.Done {
			// if we're here, we don't have an entry in the customer table for the order we just observed
			// send a message to egress-square-gateway to fetch the customer object for us, and the order will update later
			getCustomerEvent := eventschemas.NewSquareRetrieveCustomerRequest(squareCustomerID)
			eventJSON, err := getCustomerEvent.MarshalJSON()
			if err != nil {
				return err
			}
			timeoutContext, cancel := context.WithTimeout(context.Background(), PUBLISH_TIMEOUT_SEC)
			defer cancel()

			publishResult := squareCustomerRequestTopic.Publish(timeoutContext, &pubsub.Message{Data: eventJSON})
			messageID, err := publishResult.Get(timeoutContext) // this call blocks until complete or timeout occurs
			if err != nil {
				return err
			}
			slog.InfoContext(ctx, "published RetrieveCustomerRequest", "messageID", messageID, "customerID", squareCustomerID)
		}
		// if we're here, the customer object already exists in our collection, so there is nothing for us to do
		return nil
	})
}

package eventlakecontroller

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"cloud.google.com/go/firestore"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/google/uuid"
	eventschemas "github.com/kofc7186/fundraiser-manager/pkg/event/schemas"
	"github.com/kofc7186/fundraiser-manager/pkg/logging"
	"github.com/kofc7186/fundraiser-manager/pkg/util"

	_ "github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
)

const FUNCTION_NAME = "event-lake-controller"

var firestoreClient *firestore.Client

var eventLakePath string

func init() {
	slog.SetDefault(logging.FunctionLogger(FUNCTION_NAME))

	var err error
	firestoreClient, err = firestore.NewClient(context.Background(), util.GetEnvOrPanic("GCP_PROJECT"))
	if err != nil {
		panic(err)
	}

	eventLakePath = fmt.Sprintf("fundraisers/%s/events", util.GetEnvOrPanic("FUNDRAISER_ID"))

	// do this last so we are ensured to have all the required clients established above
	functions.CloudEvent("EventLakeCapture", EventLakeCapture)
}

// EventLakeCapture puts an entry in the backing Firestore DB for each observed event
func EventLakeCapture(ctx context.Context, e event.Event) error {
	// there are two CloudEvents - one for the pubsub message "event", and then the data within
	var msg eventschemas.MessagePublishedData
	if err := e.DataAs(&msg); err != nil {
		slog.Error(err.Error(), "event", e)
		return err
	}

	// extract nested CloudEvent contents as arbitrary JSON to store in event lake
	eventMap := make(map[string]interface{})
	if err := json.Unmarshal(msg.Message.Data, &eventMap); err != nil {
		slog.Error(err.Error(), "data", msg.Message.Data)
		return err
	}

	var idString string
	if id, ok := eventMap["id"]; !ok {
		// if we were to error out here, we'd just get caught in a never-ending loop; generate an ephemeral UUID
		idString = uuid.Must(uuid.NewV7()).String()
		slog.ErrorContext(ctx, "id not found inside nested CloudEvent", "eventMap", eventMap, "uuid", idString)
	} else {
		idString = id.(string)
	}

	// set the document ID to be the event UUID
	docRef := firestoreClient.Doc(fmt.Sprintf("%s/%s", eventLakePath, idString))
	wr, err := docRef.Set(ctx, eventMap)
	if err != nil {
		return err
	}

	slog.DebugContext(ctx, fmt.Sprintf("%v written to event lake at %v", docRef.ID, docRef.Path), "written_at", wr.UpdateTime)
	return nil
}

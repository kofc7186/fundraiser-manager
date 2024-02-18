package schemas

import (
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	"github.com/kofc7186/fundraiser-manager/pkg/util"
)

const applicationJSON = "application/json"

func newEvent(eventType string) *cloudevents.Event {
	event := cloudevents.NewEvent()

	event.SetSource(util.GetEnvOrPanic("K_SERVICE"))
	event.SetType(eventType)

	// UUIDv7 are used because they contain timestamps and thus can be easily sorted in chronological order
	event.SetID(uuid.Must(uuid.NewV7()).String())

	// This will be recorded as the field which triggers deletion upon a Firestore TTL policy sweep
	event.SetExtension("expiration", util.GetEnvOrPanic("EXPIRATION_TIME"))

	return &event
}

package event

import (
	"errors"

	"github.com/cloudevents/sdk-go/v2/event"
)

func Filter(validTypes []string, e *event.Event) (string, interface{}, error) {
	for _, validType := range validTypes {
		if validType == e.Type() {
			return "", nil, nil
		}
	}
	return "", nil, errors.New("event did not match one of valid types")
}

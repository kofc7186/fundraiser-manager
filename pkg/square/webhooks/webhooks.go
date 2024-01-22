package webhooks

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/kofc7186/fundraiser-manager/pkg/square/types/webhooks"
)

// verifySquareWebhookSignature validates the event contained in the http.Request originated from a known
// issuer, as cryptographically verified through the provided signature
//
// Since this needs to read the entire request body to do the signature validation, this method returns a
// pointer to a bytes.Buffer which contains the request body if the signature was successfully validated,
// and otherwise returns an error
func verifySquareWebhookSignature(r *http.Request, signatureKey, notificationURL string) (*bytes.Buffer, error) {
	hash := hmac.New(sha256.New, []byte(signatureKey))
	hash.Write([]byte(notificationURL))

	defer r.Body.Close()
	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	payload := new(bytes.Buffer)
	if err := json.Compact(payload, requestBody); err != nil {
		return nil, err
	}
	if _, err := hash.Write(payload.Bytes()); err != nil {
		return nil, err
	}

	signature := r.Header.Get("x-square-hmacsha256-signature")
	if signature != base64.StdEncoding.EncodeToString(hash.Sum(nil)) {
		return nil, errors.New("square webhook signature could not be validated")
	}
	return payload, nil
}

func VerifySquareWebhook(r *http.Request, signatureKey, notificationURL string) (webhooks.SquareWebhookEvent, error) {
	payload, err := verifySquareWebhookSignature(r, signatureKey, notificationURL)
	if err != nil {
		return nil, err
	}

	baseWebhookEvent := &webhooks.WebhookBase{}
	if err := json.Unmarshal(payload.Bytes(), baseWebhookEvent); err != nil {
		return nil, err
	}

	var typedEventPointer webhooks.SquareWebhookEvent
	switch baseWebhookEvent.Type {
	case webhooks.SQUARE_WEBHOOK_PAYMENT_CREATED:
		typedEventPointer = &webhooks.PaymentCreated{}
	case webhooks.SQUARE_WEBHOOK_PAYMENT_UPDATED:
		typedEventPointer = &webhooks.PaymentUpdated{}
	case webhooks.SQUARE_WEBHOOK_REFUND_CREATED:
		typedEventPointer = &webhooks.RefundCreated{}
	case webhooks.SQUARE_WEBHOOK_REFUND_UPDATED:
		typedEventPointer = &webhooks.RefundUpdated{}
	}

	if err := json.Unmarshal(payload.Bytes(), typedEventPointer); err != nil {
		return nil, err
	}
	return typedEventPointer, nil
}

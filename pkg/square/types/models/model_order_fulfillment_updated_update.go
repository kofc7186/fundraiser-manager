/*
 * Square Connect API
 *
 * Client library for accessing the Square Connect APIs
 *
 * API version: 2.0
 * Contact: developers@squareup.com
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package models

// Information about fulfillment updates.
type OrderFulfillmentUpdatedUpdate struct {
	// A unique ID that identifies the fulfillment only within this order.
	FulfillmentUid string `json:"fulfillment_uid,omitempty"`
	// The state of the fulfillment before the change. The state is not populated if the fulfillment is created with this new `Order` version.
	OldState string `json:"old_state,omitempty"`
	// The state of the fulfillment after the change. The state might be equal to `old_state` if a non-state field was changed on the fulfillment (such as the tracking number).
	NewState string `json:"new_state,omitempty"`
}

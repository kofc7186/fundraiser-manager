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

// Contains details necessary to fulfill a pickup order.
type OrderFulfillmentPickupDetails struct {
	// Information about the person to pick up this fulfillment from a physical location.
	Recipient *OrderFulfillmentRecipient `json:"recipient,omitempty"`
	// The [timestamp](https://developer.squareup.com/docs/build-basics/working-with-dates) indicating when this fulfillment expires if it is not accepted. The timestamp must be in RFC 3339 format (for example, \"2016-09-04T23:59:33.123Z\"). The expiration time can only be set up to 7 days in the future. If `expires_at` is not set, this pickup fulfillment is automatically accepted when placed.
	ExpiresAt string `json:"expires_at,omitempty"`
	// The duration of time after which an open and accepted pickup fulfillment is automatically moved to the `COMPLETED` state. The duration must be in RFC 3339 format (for example, \"P1W3D\"). If not set, this pickup fulfillment remains accepted until it is canceled or completed.
	AutoCompleteDuration string `json:"auto_complete_duration,omitempty"`
	// The schedule type of the pickup fulfillment. Defaults to `SCHEDULED`.
	ScheduleType string `json:"schedule_type,omitempty"`
	// The [timestamp](https://developer.squareup.com/docs/build-basics/working-with-dates) that represents the start of the pickup window. Must be in RFC 3339 timestamp format, e.g., \"2016-09-04T23:59:33.123Z\". For fulfillments with the schedule type `ASAP`, this is automatically set to the current time plus the expected duration to prepare the fulfillment.
	PickupAt string `json:"pickup_at,omitempty"`
	// The window of time in which the order should be picked up after the `pickup_at` timestamp. Must be in RFC 3339 duration format, e.g., \"P1W3D\". Can be used as an informational guideline for merchants.
	PickupWindowDuration string `json:"pickup_window_duration,omitempty"`
	// The duration of time it takes to prepare this fulfillment. The duration must be in RFC 3339 format (for example, \"P1W3D\").
	PrepTimeDuration string `json:"prep_time_duration,omitempty"`
	// A note to provide additional instructions about the pickup fulfillment displayed in the Square Point of Sale application and set by the API.
	Note string `json:"note,omitempty"`
	// The [timestamp](https://developer.squareup.com/docs/build-basics/working-with-dates) indicating when the fulfillment was placed. The timestamp must be in RFC 3339 format (for example, \"2016-09-04T23:59:33.123Z\").
	PlacedAt string `json:"placed_at,omitempty"`
	// The [timestamp](https://developer.squareup.com/docs/build-basics/working-with-dates) indicating when the fulfillment was accepted. The timestamp must be in RFC 3339 format (for example, \"2016-09-04T23:59:33.123Z\").
	AcceptedAt string `json:"accepted_at,omitempty"`
	// The [timestamp](https://developer.squareup.com/docs/build-basics/working-with-dates) indicating when the fulfillment was rejected. The timestamp must be in RFC 3339 format (for example, \"2016-09-04T23:59:33.123Z\").
	RejectedAt string `json:"rejected_at,omitempty"`
	// The [timestamp](https://developer.squareup.com/docs/build-basics/working-with-dates) indicating when the fulfillment is marked as ready for pickup. The timestamp must be in RFC 3339 format (for example, \"2016-09-04T23:59:33.123Z\").
	ReadyAt string `json:"ready_at,omitempty"`
	// The [timestamp](https://developer.squareup.com/docs/build-basics/working-with-dates) indicating when the fulfillment expired. The timestamp must be in RFC 3339 format (for example, \"2016-09-04T23:59:33.123Z\").
	ExpiredAt string `json:"expired_at,omitempty"`
	// The [timestamp](https://developer.squareup.com/docs/build-basics/working-with-dates) indicating when the fulfillment was picked up by the recipient. The timestamp must be in RFC 3339 format (for example, \"2016-09-04T23:59:33.123Z\").
	PickedUpAt string `json:"picked_up_at,omitempty"`
	// The [timestamp](https://developer.squareup.com/docs/build-basics/working-with-dates) indicating when the fulfillment was canceled. The timestamp must be in RFC 3339 format (for example, \"2016-09-04T23:59:33.123Z\").
	CanceledAt string `json:"canceled_at,omitempty"`
	// A description of why the pickup was canceled. The maximum length: 100 characters.
	CancelReason string `json:"cancel_reason,omitempty"`
	// If set to `true`, indicates that this pickup order is for curbside pickup, not in-store pickup.
	IsCurbsidePickup bool `json:"is_curbside_pickup,omitempty"`
	// Specific details for curbside pickup. These details can only be populated if `is_curbside_pickup` is set to `true`.
	CurbsidePickupDetails *OrderFulfillmentPickupDetailsCurbsidePickupDetails `json:"curbside_pickup_details,omitempty"`
}

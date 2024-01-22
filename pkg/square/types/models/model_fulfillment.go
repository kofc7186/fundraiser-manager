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

// Contains details about how to fulfill this order. Orders can only be created with at most one fulfillment using the API. However, orders returned by the Orders API might contain multiple fulfillments because sellers can create multiple fulfillments using Square products such as Square Online.
type Fulfillment struct {
	// A unique ID that identifies the fulfillment only within this order.
	Uid string `json:"uid,omitempty"`
	// The type of the fulfillment.
	Type_ string `json:"type,omitempty"`
	// The state of the fulfillment.
	State string `json:"state,omitempty"`
	// Describes what order line items this fulfillment applies to. It can be `ALL` or `ENTRY_LIST` with a supplied list of fulfillment entries.
	LineItemApplication string `json:"line_item_application,omitempty"`
	// A list of entries pertaining to the fulfillment of an order. Each entry must reference a valid `uid` for an order line item in the `line_item_uid` field, as well as a `quantity` to fulfill.  Multiple entries can reference the same line item `uid`, as long as the total quantity among all fulfillment entries referencing a single line item does not exceed the quantity of the order's line item itself.  An order cannot be marked as `COMPLETED` before all fulfillments are `COMPLETED`, `CANCELED`, or `FAILED`. Fulfillments can be created and completed independently before order completion.
	Entries []FulfillmentFulfillmentEntry `json:"entries,omitempty"`
	// Application-defined data attached to this fulfillment. Metadata fields are intended to store descriptive references or associations with an entity in another system or store brief information about the object. Square does not process this field; it only stores and returns it in relevant API calls. Do not use metadata to store any sensitive information (such as personally identifiable information or card details).  Keys written by applications must be 60 characters or less and must be in the character set `[a-zA-Z0-9_-]`. Entries can also include metadata generated by Square. These keys are prefixed with a namespace, separated from the key with a ':' character.  Values have a maximum length of 255 characters.  An application can have up to 10 entries per metadata field.  Entries written by applications are private and can only be read or modified by the same application.  For more information, see [Metadata](https://developer.squareup.com/docs/build-basics/metadata).
	Metadata map[string]string `json:"metadata,omitempty"`
	// Contains details for a pickup fulfillment. These details are required when the fulfillment type is `PICKUP`.
	PickupDetails *FulfillmentPickupDetails `json:"pickup_details,omitempty"`
	// Contains details for a shipment fulfillment. These details are required when the fulfillment type is `SHIPMENT`.  A shipment fulfillment's relationship to fulfillment `state`: `PROPOSED`: A shipment is requested. `RESERVED`: Fulfillment accepted. Shipment processing. `PREPARED`: Shipment packaged. Shipping label created. `COMPLETED`: Package has been shipped. `CANCELED`: Shipment has been canceled. `FAILED`: Shipment has failed.
	ShipmentDetails *FulfillmentShipmentDetails `json:"shipment_details,omitempty"`
	// 
	DeliveryDetails *FulfillmentDeliveryDetails `json:"delivery_details,omitempty"`
}

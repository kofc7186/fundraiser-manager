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

// Information about the fulfillment recipient.
type OrderFulfillmentRecipient struct {
	// The ID of the customer associated with the fulfillment. If `customer_id` is provided, the fulfillment recipient's `display_name`, `email_address`, and `phone_number` are automatically populated from the targeted customer profile. If these fields are set in the request, the request values override the information from the customer profile. If the targeted customer profile does not contain the necessary information and these fields are left unset, the request results in an error.
	CustomerId string `json:"customer_id,omitempty"`
	// The display name of the fulfillment recipient. This field is required. If provided, the display name overrides the corresponding customer profile value indicated by `customer_id`.
	DisplayName string `json:"display_name,omitempty"`
	// The email address of the fulfillment recipient. If provided, the email address overrides the corresponding customer profile value indicated by `customer_id`.
	EmailAddress string `json:"email_address,omitempty"`
	// The phone number of the fulfillment recipient. This field is required. If provided, the phone number overrides the corresponding customer profile value indicated by `customer_id`.
	PhoneNumber string `json:"phone_number,omitempty"`
	// The address of the fulfillment recipient. This field is required. If provided, the address overrides the corresponding customer profile value indicated by `customer_id`.
	Address *Address `json:"address,omitempty"`
}

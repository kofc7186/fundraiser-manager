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

// Represents the payment details of a card to be used for payments. These details are determined by the payment token generated by Web Payments SDK.
type Card struct {
	// Unique ID for this card. Generated by Square.
	Id string `json:"id,omitempty"`
	// The card's brand.
	CardBrand string `json:"card_brand,omitempty"`
	// The last 4 digits of the card number.
	Last4 string `json:"last_4,omitempty"`
	// The expiration month of the associated card as an integer between 1 and 12.
	ExpMonth int64 `json:"exp_month,omitempty"`
	// The four-digit year of the card's expiration date.
	ExpYear int64 `json:"exp_year,omitempty"`
	// The name of the cardholder.
	CardholderName string `json:"cardholder_name,omitempty"`
	// The billing address for this card.
	BillingAddress *Address `json:"billing_address,omitempty"`
	// Intended as a Square-assigned identifier, based on the card number, to identify the card across multiple locations within a single application.
	Fingerprint string `json:"fingerprint,omitempty"`
	// **Required** The ID of a customer created using the Customers API to be associated with the card.
	CustomerId string `json:"customer_id,omitempty"`
	// The ID of the merchant associated with the card.
	MerchantId string `json:"merchant_id,omitempty"`
	// An optional user-defined reference ID that associates this card with another entity in an external system. For example, a customer ID from an external customer management system.
	ReferenceId string `json:"reference_id,omitempty"`
	// Indicates whether or not a card can be used for payments.
	Enabled bool `json:"enabled,omitempty"`
	// The type of the card. The Card object includes this field only in response to Payments API calls.
	CardType string `json:"card_type,omitempty"`
	// Indicates whether the Card is prepaid or not. The Card object includes this field only in response to Payments API calls.
	PrepaidType string `json:"prepaid_type,omitempty"`
	// The first six digits of the card number, known as the Bank Identification Number (BIN). Only the Payments API returns this field.
	Bin string `json:"bin,omitempty"`
	// Current version number of the card. Increments with each card update. Requests to update an existing Card object will be rejected unless the version in the request matches the current version for the Card.
	Version int64 `json:"version,omitempty"`
	// The card's co-brand if available. For example, an Afterpay virtual card would have a co-brand of AFTERPAY.
	CardCoBrand string `json:"card_co_brand,omitempty"`
}

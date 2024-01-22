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

// The customer address filter. This filter is used in a [CustomerCustomAttributeFilterValue](https://developer.squareup.com/reference/square_2023-12-13/objects/CustomerCustomAttributeFilterValue) filter when searching by an `Address`-type custom attribute.
type CustomerAddressFilter struct {
	// The postal code to search for. Only an `exact` match is supported.
	PostalCode *CustomerTextFilter `json:"postal_code,omitempty"`
	// The country code to search for.
	Country string `json:"country,omitempty"`
}

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

// Represents an [UpsertMerchantCustomAttribute](https://developer.squareup.com/reference/square_2023-12-13/merchant-custom-attributes-api/upsert-merchant-custom-attribute) response. Either `custom_attribute_definition` or `errors` is present in the response.
type UpsertMerchantCustomAttributeResponse struct {
	// The new or updated custom attribute.
	CustomAttribute *CustomAttribute `json:"custom_attribute,omitempty"`
	// Any errors that occurred during the request.
	Errors []ModelError `json:"errors,omitempty"`
}

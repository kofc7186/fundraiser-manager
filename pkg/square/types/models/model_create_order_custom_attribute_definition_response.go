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

// Represents a response from creating an order custom attribute definition.
type CreateOrderCustomAttributeDefinitionResponse struct {
	// The new custom attribute definition.
	CustomAttributeDefinition *CustomAttributeDefinition `json:"custom_attribute_definition,omitempty"`
	// Any errors that occurred during the request.
	Errors []ModelError `json:"errors,omitempty"`
}

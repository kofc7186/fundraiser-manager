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
// FulfillmentFulfillmentLineItemApplication : The `line_item_application` describes what order line items this fulfillment applies to. It can be `ALL` or `ENTRY_LIST` with a supplied list of fulfillment entries.
type FulfillmentFulfillmentLineItemApplication string

// List of FulfillmentFulfillmentLineItemApplication
const (
	ALL_FulfillmentFulfillmentLineItemApplication FulfillmentFulfillmentLineItemApplication = "ALL"
	ENTRY_LIST_FulfillmentFulfillmentLineItemApplication FulfillmentFulfillmentLineItemApplication = "ENTRY_LIST"
)

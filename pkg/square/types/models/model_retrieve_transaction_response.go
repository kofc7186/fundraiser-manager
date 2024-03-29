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

// Defines the fields that are included in the response body of a request to the [RetrieveTransaction](https://developer.squareup.com/reference/square_2023-12-13/transactions-api/retrieve-transaction) endpoint.  One of `errors` or `transaction` is present in a given response (never both).
type RetrieveTransactionResponse struct {
	// Any errors that occurred during the request.
	Errors []ModelError `json:"errors,omitempty"`
	// The requested transaction.
	Transaction *Transaction `json:"transaction,omitempty"`
}

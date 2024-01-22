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

// Represents the details of a tender with `type` `SQUARE_ACCOUNT`.
type TenderSquareAccountDetails struct {
	// The Square Account payment's current state (such as `AUTHORIZED` or `CAPTURED`). See [TenderSquareAccountDetailsStatus](https://developer.squareup.com/reference/square_2023-12-13/enums/TenderSquareAccountDetailsStatus) for possible values.
	Status string `json:"status,omitempty"`
}

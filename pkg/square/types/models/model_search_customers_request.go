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

// Defines the fields that are included in the request body of a request to the `SearchCustomers` endpoint.
type SearchCustomersRequest struct {
	// Include the pagination cursor in subsequent calls to this endpoint to retrieve the next set of results associated with the original query.  For more information, see [Pagination](https://developer.squareup.com/docs/build-basics/common-api-patterns/pagination).
	Cursor string `json:"cursor,omitempty"`
	// The maximum number of results to return in a single page. This limit is advisory. The response might contain more or fewer results. If the specified limit is invalid, Square returns a `400 VALUE_TOO_LOW` or `400 VALUE_TOO_HIGH` error. The default value is 100.  For more information, see [Pagination](https://developer.squareup.com/docs/build-basics/common-api-patterns/pagination).
	Limit int64 `json:"limit,omitempty"`
	// The filtering and sorting criteria for the search request. If a query is not specified, Square returns all customer profiles ordered alphabetically by `given_name` and `family_name`.
	Query *CustomerQuery `json:"query,omitempty"`
	// Indicates whether to return the total count of matching customers in the `count` field of the response.  The default value is `false`.
	Count bool `json:"count,omitempty"`
}

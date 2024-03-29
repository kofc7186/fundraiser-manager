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

// Represents the origination details of an order.
type OrderSource struct {
	// The name used to identify the place (physical or digital) that an order originates. If unset, the name defaults to the name of the application that created the order.
	Name string `json:"name,omitempty"`
}

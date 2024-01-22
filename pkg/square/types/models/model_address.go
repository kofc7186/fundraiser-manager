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

// Represents a postal address in a country.  For more information, see [Working with Addresses](https://developer.squareup.com/docs/build-basics/working-with-addresses).
type Address struct {
	// The first line of the address.  Fields that start with `address_line` provide the address's most specific details, like street number, street name, and building name. They do *not* provide less specific details like city, state/province, or country (these details are provided in other fields).
	AddressLine1 string `json:"address_line_1,omitempty"`
	// The second line of the address, if any.
	AddressLine2 string `json:"address_line_2,omitempty"`
	// The third line of the address, if any.
	AddressLine3 string `json:"address_line_3,omitempty"`
	// The city or town of the address. For a full list of field meanings by country, see [Working with Addresses](https://developer.squareup.com/docs/build-basics/working-with-addresses).
	Locality string `json:"locality,omitempty"`
	// A civil region within the address's `locality`, if any.
	Sublocality string `json:"sublocality,omitempty"`
	// A civil region within the address's `sublocality`, if any.
	Sublocality2 string `json:"sublocality_2,omitempty"`
	// A civil region within the address's `sublocality_2`, if any.
	Sublocality3 string `json:"sublocality_3,omitempty"`
	// A civil entity within the address's country. In the US, this is the state. For a full list of field meanings by country, see [Working with Addresses](https://developer.squareup.com/docs/build-basics/working-with-addresses).
	AdministrativeDistrictLevel1 string `json:"administrative_district_level_1,omitempty"`
	// A civil entity within the address's `administrative_district_level_1`. In the US, this is the county.
	AdministrativeDistrictLevel2 string `json:"administrative_district_level_2,omitempty"`
	// A civil entity within the address's `administrative_district_level_2`, if any.
	AdministrativeDistrictLevel3 string `json:"administrative_district_level_3,omitempty"`
	// The address's postal code. For a full list of field meanings by country, see [Working with Addresses](https://developer.squareup.com/docs/build-basics/working-with-addresses).
	PostalCode string `json:"postal_code,omitempty"`
	// The address's country, in the two-letter format of ISO 3166. For example, `US` or `FR`.
	Country string `json:"country,omitempty"`
	// Optional first name when it's representing recipient.
	FirstName string `json:"first_name,omitempty"`
	// Optional last name when it's representing recipient.
	LastName string `json:"last_name,omitempty"`
}

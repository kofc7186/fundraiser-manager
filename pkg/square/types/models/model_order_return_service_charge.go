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

// Represents the service charge applied to the original order.
type OrderReturnServiceCharge struct {
	// A unique ID that identifies the return service charge only within this order.
	Uid string `json:"uid,omitempty"`
	// The service charge `uid` from the order containing the original service charge. `source_service_charge_uid` is `null` for unlinked returns.
	SourceServiceChargeUid string `json:"source_service_charge_uid,omitempty"`
	// The name of the service charge.
	Name string `json:"name,omitempty"`
	// The catalog object ID of the associated [OrderServiceCharge](https://developer.squareup.com/reference/square_2023-12-13/objects/OrderServiceCharge).
	CatalogObjectId string `json:"catalog_object_id,omitempty"`
	// The version of the catalog object that this service charge references.
	CatalogVersion int64 `json:"catalog_version,omitempty"`
	// The percentage of the service charge, as a string representation of a decimal number. For example, a value of `\"7.25\"` corresponds to a percentage of 7.25%.  Either `percentage` or `amount_money` should be set, but not both.
	Percentage string `json:"percentage,omitempty"`
	// The amount of a non-percentage-based service charge.  Either `percentage` or `amount_money` should be set, but not both.
	AmountMoney *Money `json:"amount_money,omitempty"`
	// The amount of money applied to the order by the service charge, including any inclusive tax amounts, as calculated by Square.  - For fixed-amount service charges, `applied_money` is equal to `amount_money`. - For percentage-based service charges, `applied_money` is the money calculated using the percentage.
	AppliedMoney *Money `json:"applied_money,omitempty"`
	// The total amount of money to collect for the service charge.  __NOTE__: If an inclusive tax is applied to the service charge, `total_money` does not equal `applied_money` plus `total_tax_money` because the inclusive tax amount is already included in both `applied_money` and `total_tax_money`.
	TotalMoney *Money `json:"total_money,omitempty"`
	// The total amount of tax money to collect for the service charge.
	TotalTaxMoney *Money `json:"total_tax_money,omitempty"`
	// The calculation phase after which to apply the service charge.
	CalculationPhase string `json:"calculation_phase,omitempty"`
	// Indicates whether the surcharge can be taxed. Service charges calculated in the `TOTAL_PHASE` cannot be marked as taxable.
	Taxable bool `json:"taxable,omitempty"`
	// The list of references to `OrderReturnTax` entities applied to the `OrderReturnServiceCharge`. Each `OrderLineItemAppliedTax` has a `tax_uid` that references the `uid` of a top-level `OrderReturnTax` that is being applied to the `OrderReturnServiceCharge`. On reads, the applied amount is populated.
	AppliedTaxes []OrderLineItemAppliedTax `json:"applied_taxes,omitempty"`
	// The treatment type of the service charge.
	TreatmentType string `json:"treatment_type,omitempty"`
	// Indicates the level at which the apportioned service charge applies. For `ORDER` scoped service charges, Square generates references in `applied_service_charges` on all order line items that do not have them. For `LINE_ITEM` scoped service charges, the service charge only applies to line items with a service charge reference in their `applied_service_charges` field.  This field is immutable. To change the scope of an apportioned service charge, you must delete the apportioned service charge and re-add it as a new apportioned service charge.
	Scope string `json:"scope,omitempty"`
}

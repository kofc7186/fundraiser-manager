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

// Represents an applied portion of a discount to a line item in an order.  Order scoped discounts have automatically applied discounts present for each line item. Line-item scoped discounts must have applied discounts added manually for any applicable line items. The corresponding applied money is automatically computed based on participating line items.
type OrderLineItemAppliedDiscount struct {
	// A unique ID that identifies the applied discount only within this order.
	Uid string `json:"uid,omitempty"`
	// The `uid` of the discount that the applied discount represents. It must reference a discount present in the `order.discounts` field.  This field is immutable. To change which discounts apply to a line item, you must delete the discount and re-add it as a new `OrderLineItemAppliedDiscount`.
	DiscountUid string `json:"discount_uid"`
	// The amount of money applied by the discount to the line item.
	AppliedMoney *Money `json:"applied_money,omitempty"`
}

package order

import (
	"errors"
	"regexp"

	"github.com/kofc7186/fundraiser-manager/pkg/types/customer"
)

type OrderUpdateFunc func(*Order, *Order) error

// Customer Create needs all fields updated
// Customer Update should iterate through fieldmask, applying regex, and if matches call update function

func (o *Order) UpdateFromCustomerCreated(customer *customer.Customer) error {
	return nil
}

// updateMap maps eventTypes + fieldMasks (set with 'firestore' struct tags) to specific order update functions
var updateMap map[*regexp.Regexp]map[*regexp.Regexp][]OrderUpdateFunc = map[*regexp.Regexp]map[*regexp.Regexp][]OrderUpdateFunc{
	regexp.MustCompile("^org.kofc7186.fundraiserManager.customer.updated$"): {
		regexp.MustCompile("^emailAddress$"):         []OrderUpdateFunc{UpdateEmailAddress},
		regexp.MustCompile("^firstName$"):            []OrderUpdateFunc{UpdateFirstName},
		regexp.MustCompile("^lastName$"):             []OrderUpdateFunc{UpdateLastName},
		regexp.MustCompile("^(firstName|lastName)$"): []OrderUpdateFunc{UpdateDisplayName},
		regexp.MustCompile("^phoneNumber$"):          []OrderUpdateFunc{UpdatePhoneNumber},
		regexp.MustCompile("^isKnight$"):             []OrderUpdateFunc{UpdateIsKnight},
	},
	regexp.MustCompile("^org.kofc7186.fundraiserManager.payment.updated$"): {
		regexp.MustCompile("^feeAmount$"):        []OrderUpdateFunc{UpdateEmailAddress},
		regexp.MustCompile("^note$"):             []OrderUpdateFunc{UpdateEmailAddress},
		regexp.MustCompile("^receiptURL$"):       []OrderUpdateFunc{UpdateEmailAddress},
		regexp.MustCompile("^source$"):           []OrderUpdateFunc{UpdateEmailAddress},
		regexp.MustCompile("^squareCustomerID$"): []OrderUpdateFunc{UpdateEmailAddress},
		regexp.MustCompile("^squarePaymentID$"):  []OrderUpdateFunc{UpdateEmailAddress},
		regexp.MustCompile("^tipAmount$"):        []OrderUpdateFunc{UpdateEmailAddress},
		regexp.MustCompile("^totalAmount$"):      []OrderUpdateFunc{UpdateEmailAddress},
	},
	regexp.MustCompile("^org.kofc7186.fundraiserManager.square.retrieveOrder.response$"): {
		// CreatedAt, Items, SquareOrderState, Version
		// []Tender: SquarePaymentId, CustomerId
		// []Fulfillment/Pickup: CustomerId, DisplayName, EmailAddress, PhoneNumber, Note
		// Items!
		regexp.MustCompile(".?"): []OrderUpdateFunc{UpdateEmailAddress},
	},
}

func UpdateEmailAddress(orderToUpdate, updatedOrder *Order) error {
	if orderToUpdate == nil {
		return errors.New("orderToUpdate is nil")
	} else if updatedOrder == nil {
		return errors.New("updatedOrder is nil")
	}
	if orderToUpdate.EmailAddress == "" {
		orderToUpdate.EmailAddress = updatedOrder.EmailAddress
	}
	return nil
}

func UpdateFirstName(orderToUpdate, updatedOrder *Order) error {
	if orderToUpdate == nil {
		return errors.New("orderToUpdate is nil")
	} else if updatedOrder == nil {
		return errors.New("updatedOrder is nil")
	}
	if orderToUpdate.FirstName == "" {
		orderToUpdate.FirstName = updatedOrder.FirstName
	}
	return nil
}

func UpdateLastName(orderToUpdate, updatedOrder *Order) error {
	if orderToUpdate == nil {
		return errors.New("orderToUpdate is nil")
	} else if updatedOrder == nil {
		return errors.New("updatedOrder is nil")
	}
	if orderToUpdate.LastName == "" {
		orderToUpdate.LastName = updatedOrder.LastName
	}
	return nil
}

func UpdateDisplayName(orderToUpdate, updatedOrder *Order) error {
	if orderToUpdate == nil {
		return errors.New("orderToUpdate is nil")
	} else if updatedOrder == nil {
		return errors.New("updatedOrder is nil")
	}
	if orderToUpdate.DisplayName == "" {
		orderToUpdate.DisplayName = updatedOrder.FirstName + " " + updatedOrder.LastName
	}
	return nil
}

func UpdatePhoneNumber(orderToUpdate, updatedOrder *Order) error {
	if orderToUpdate == nil {
		return errors.New("orderToUpdate is nil")
	} else if updatedOrder == nil {
		return errors.New("updatedOrder is nil")
	}
	if orderToUpdate.PhoneNumber == "" {
		orderToUpdate.PhoneNumber = updatedOrder.PhoneNumber
	}
	return nil
}

func UpdateIsKnight(orderToUpdate, updatedOrder *Order) error {
	if orderToUpdate == nil {
		return errors.New("orderToUpdate is nil")
	} else if updatedOrder == nil {
		return errors.New("updatedOrder is nil")
	}
	orderToUpdate.KnightOfColumbus = updatedOrder.KnightOfColumbus
	return nil
}

package database

import "strings"

// Order represents sorgint order
type Order string

const (
	// OrderAsc is the ascending sort order
	OrderAsc = Order("asc")

	// OrderDesc is the descending sort order
	OrderDesc = Order("desc")
)

// OrderFromString creates a Status from a string
func OrderFromString(str string) Order {
	str = strings.ToLower(str)

	o := OrderAsc

	if str == "desc" {
		o = OrderDesc
	}

	return o
}

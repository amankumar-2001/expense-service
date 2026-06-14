// Package expensecategory enumerates the expense categories the tracker buckets
// spending into. Values match the strings the web client and PRD use verbatim.
package expensecategory

// Category is a discretionary-spend bucket.
type Category string

const (
	Groceries     Category = "Groceries"
	Transport     Category = "Transport"
	FoodDining    Category = "Food/Dining"
	Utilities     Category = "Utilities"
	Medical       Category = "Medical"
	Entertainment Category = "Entertainment"
	Shopping      Category = "Shopping"
	Other         Category = "Other"
)

func (c Category) String() string { return string(c) }

// Valid reports whether c is a recognised category.
func (c Category) Valid() bool {
	switch c {
	case Groceries, Transport, FoodDining, Utilities, Medical, Entertainment, Shopping, Other:
		return true
	default:
		return false
	}
}

// All lists every category, used for validation and stable summary ordering.
func All() []Category {
	return []Category{Groceries, Transport, FoodDining, Utilities, Medical, Entertainment, Shopping, Other}
}

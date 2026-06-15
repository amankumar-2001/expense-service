package parser

import (
	"strings"

	"github.com/kharchibook/expense-service/enums/expensecategory"
)

// categoryKeywords maps a category to its English trigger words (PRD §3.4,
// English entries only — multilingual support is out of scope for v1).
var categoryKeywords = map[expensecategory.Category][]string{
	expensecategory.Groceries:     {"sabzi", "vegetables", "kirana", "grocery", "groceries", "market", "milk", "atta", "dal", "rice"},
	expensecategory.Transport:     {"petrol", "diesel", "uber", "ola", "auto", "rickshaw", "bus", "metro", "cab", "fuel", "taxi"},
	expensecategory.FoodDining:    {"zomato", "swiggy", "restaurant", "hotel", "chai", "coffee", "lunch", "dinner", "breakfast", "food", "snacks"},
	expensecategory.Utilities:     {"electricity", "light bill", "water bill", "gas", "internet", "wifi", "broadband", "recharge", "bill"},
	expensecategory.Medical:       {"medicine", "doctor", "hospital", "pharmacy", "chemist", "clinic", "medical"},
	expensecategory.Entertainment: {"movie", "netflix", "prime", "hotstar", "spotify", "concert", "game", "cinema"},
	expensecategory.Shopping:      {"amazon", "flipkart", "clothes", "shoes", "mall", "shopping", "myntra"},
}

// categorize matches the descriptive text against the keyword dictionary,
// returning the first category whose keyword appears. Falls back to Other when
// nothing matches (the spot where a Claude categorization call would later slot
// in for the ~5% ambiguous tail — PRD §3.4).
func categorize(note string) expensecategory.Category {
	n := strings.ToLower(note)
	for _, cat := range expensecategory.All() {
		for _, kw := range categoryKeywords[cat] {
			if strings.Contains(n, kw) {
				return cat
			}
		}
	}
	return expensecategory.Other
}

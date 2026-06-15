package parser

import (
	"testing"

	"github.com/kharchibook/expense-service/enums/autopaytype"
	"github.com/kharchibook/expense-service/enums/expensecategory"
)

func TestParseExpense(t *testing.T) {
	cases := []struct {
		in     string
		amount float64
		cat    expensecategory.Category
	}{
		{"200 sabzi", 200, expensecategory.Groceries},
		{"500 petrol", 500, expensecategory.Transport},
		{"800 zomato", 800, expensecategory.FoodDining},
		{"1200 electricity bill", 1200, expensecategory.Utilities},
		{"99 something random", 99, expensecategory.Other},
	}
	for _, c := range cases {
		got := Parse(c.in)
		if got.Kind != KindExpense {
			t.Errorf("%q: kind = %v, want expense", c.in, got.Kind)
			continue
		}
		if got.Amount != c.amount {
			t.Errorf("%q: amount = %v, want %v", c.in, got.Amount, c.amount)
		}
		if got.Category != c.cat {
			t.Errorf("%q: category = %v, want %v", c.in, got.Category, c.cat)
		}
	}
}

func TestParseCommands(t *testing.T) {
	if got := Parse("hi"); got.Kind != KindGreeting {
		t.Errorf("hi: kind = %v, want greeting", got.Kind)
	}
	if got := Parse("summary"); got.Kind != KindSummary || got.Month != "" {
		t.Errorf("summary: %+v", got)
	}
	if got := Parse("summary june"); got.Kind != KindSummary || got.Month != "june" {
		t.Errorf("summary june: month = %q", got.Month)
	}
	if got := Parse("autopay list"); got.Kind != KindAutoPayList {
		t.Errorf("autopay list: kind = %v", got.Kind)
	}
	if got := Parse("upcoming"); got.Kind != KindUpcoming {
		t.Errorf("upcoming: kind = %v", got.Kind)
	}
	if got := Parse("committed"); got.Kind != KindCommitted {
		t.Errorf("committed: kind = %v", got.Kind)
	}
	if got := Parse("remove netflix"); got.Kind != KindRemove || got.Name != "netflix" {
		t.Errorf("remove netflix: %+v", got)
	}
}

func TestParseAddAutoPay(t *testing.T) {
	emi := Parse("add emi home loan 22000 on 5th")
	if emi.Kind != KindAddAutoPay {
		t.Fatalf("kind = %v, want add-autopay", emi.Kind)
	}
	if emi.Type != autopaytype.EMI {
		t.Errorf("type = %v, want emi", emi.Type)
	}
	if emi.Amount != 22000 {
		t.Errorf("amount = %v, want 22000", emi.Amount)
	}
	if emi.DeductDay != 5 {
		t.Errorf("deductDay = %v, want 5", emi.DeductDay)
	}
	if emi.Name != "home loan" {
		t.Errorf("name = %q, want %q", emi.Name, "home loan")
	}

	sub := Parse("add autopay netflix 649 on 12th")
	if sub.Type != autopaytype.Subscription {
		t.Errorf("netflix type = %v, want subscription", sub.Type)
	}
	if sub.Amount != 649 || sub.DeductDay != 12 || sub.Name != "netflix" {
		t.Errorf("netflix: %+v", sub)
	}

	ins := Parse("add autopay lic premium 4200 on 28th")
	if ins.Type != autopaytype.Insurance {
		t.Errorf("lic type = %v, want insurance", ins.Type)
	}
}

func TestParseSalary(t *testing.T) {
	s := Parse("salary 65000 on 1st")
	if s.Kind != KindSalary {
		t.Fatalf("kind = %v, want salary", s.Kind)
	}
	if s.Amount != 65000 {
		t.Errorf("amount = %v, want 65000", s.Amount)
	}
	if s.SalaryDay != 1 {
		t.Errorf("salaryDay = %v, want 1", s.SalaryDay)
	}
}

func TestParseUnknown(t *testing.T) {
	if got := Parse("blah blah"); got.Kind != KindUnknown {
		t.Errorf("kind = %v, want unknown", got.Kind)
	}
	if got := Parse(""); got.Kind != KindUnknown {
		t.Errorf("empty: kind = %v, want unknown", got.Kind)
	}
}

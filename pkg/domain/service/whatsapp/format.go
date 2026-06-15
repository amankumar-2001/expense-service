package whatsapp

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kharchibook/expense-service/enums/autopaytype"
	"github.com/kharchibook/expense-service/enums/expensecategory"
	"github.com/kharchibook/expense-service/pkg/domain/dto/response"
)

// inr formats an amount the Indian way: ₹1,23,456 (no decimals).
func inr(amount float64) string {
	n := int64(amount + 0.5)
	neg := n < 0
	if neg {
		n = -n
	}
	s := strconv.FormatInt(n, 10)
	var grouped string
	if len(s) <= 3 {
		grouped = s
	} else {
		last3 := s[len(s)-3:]
		rest := s[:len(s)-3]
		var parts []string
		for len(rest) > 2 {
			parts = append([]string{rest[len(rest)-2:]}, parts...)
			rest = rest[:len(rest)-2]
		}
		if rest != "" {
			parts = append([]string{rest}, parts...)
		}
		grouped = strings.Join(parts, ",") + "," + last3
	}
	if neg {
		grouped = "-" + grouped
	}
	return "₹" + grouped
}

var categoryEmoji = map[expensecategory.Category]string{
	expensecategory.Groceries:     "🛒",
	expensecategory.Transport:     "🚗",
	expensecategory.FoodDining:    "🍽️",
	expensecategory.Utilities:     "💡",
	expensecategory.Medical:       "🏥",
	expensecategory.Entertainment: "🎬",
	expensecategory.Shopping:      "🛍️",
	expensecategory.Other:         "📦",
}

func ordinalDay(day int) string {
	suffix := "th"
	if day < 11 || day > 13 {
		switch day % 10 {
		case 1:
			suffix = "st"
		case 2:
			suffix = "nd"
		case 3:
			suffix = "rd"
		}
	}
	return strconv.Itoa(day) + suffix
}

// formatExpenseConfirmation acknowledges a logged expense.
func formatExpenseConfirmation(e *response.ExpenseResponse) string {
	emoji := categoryEmoji[expensecategory.Category(e.Category)]
	return fmt.Sprintf("✅ Logged %s under %s %s", inr(e.Amount), emoji, e.Category)
}

// formatSummary renders the monthly spend breakdown (PRD §3.5).
func formatSummary(s *response.ExpenseSummaryResponse) string {
	var b strings.Builder
	fmt.Fprintf(&b, "📊 *%s — Expense Summary*\n\n", s.Month)
	if len(s.ByCategory) == 0 {
		b.WriteString("No expenses logged yet this month.\n")
	}
	for _, c := range s.ByCategory {
		emoji := categoryEmoji[expensecategory.Category(c.Category)]
		fmt.Fprintf(&b, "%s %-12s — %s\n", emoji, c.Category, inr(c.Total))
	}
	fmt.Fprintf(&b, "\n*Total spent: %s*\n\nType *autopay list* to see your fixed monthly commitments.", inr(s.Total))
	return b.String()
}

var typeHeader = []struct {
	typ    autopaytype.Type
	header string
}{
	{autopaytype.EMI, "🏠 *EMIs*"},
	{autopaytype.Subscription, "📱 *Subscriptions*"},
	{autopaytype.Insurance, "🛡️ *Insurance*"},
	{autopaytype.SIP, "📈 *SIPs*"},
	{autopaytype.Other, "📦 *Other*"},
}

// formatCommitted renders the monthly commitments + free-money (PRD §4.4). Used
// for both `autopay list` and `committed`.
func formatCommitted(c *response.CommittedSummaryResponse) string {
	var b strings.Builder
	b.WriteString("📋 *Your Monthly Commitments*\n")

	byType := map[autopaytype.Type][]response.AutoPayResponse{}
	for _, a := range c.Autopays {
		t := autopaytype.Type(a.Type)
		byType[t] = append(byType[t], a)
	}
	for _, th := range typeHeader {
		items := byType[th.typ]
		if len(items) == 0 {
			continue
		}
		fmt.Fprintf(&b, "\n%s\n", th.header)
		for _, a := range items {
			fmt.Fprintf(&b, "• %-13s — %s  (%s)\n", a.Name, inr(a.Amount), ordinalDay(a.DeductDay))
		}
	}

	fmt.Fprintf(&b, "\n💰 Total locked: %s/month", inr(c.TotalCommitted))
	if c.MonthlySalary > 0 {
		pct := int((c.TotalCommitted/c.MonthlySalary)*100 + 0.5)
		fmt.Fprintf(&b, "\nThat's *%d%%* of your %s salary.\nFree to spend: *%s*",
			pct, inr(c.MonthlySalary), inr(c.FreeMoney))
	} else {
		b.WriteString("\nSet your salary with *salary <amount> on <day>* to see your free money.")
	}
	return b.String()
}

// formatUpcoming renders deductions due in the next 7 days (PRD §4.3 `upcoming`).
func formatUpcoming(items []response.UpcomingDeductionResponse) string {
	if len(items) == 0 {
		return "🎉 Nothing due in the next 7 days."
	}
	var b strings.Builder
	b.WriteString("⏰ *Upcoming deductions (next 7 days)*\n\n")
	for _, d := range items {
		day := "today"
		if d.InDays == 1 {
			day = "tomorrow"
		} else if d.InDays > 1 {
			day = fmt.Sprintf("in %d days", d.InDays)
		}
		fmt.Fprintf(&b, "• %s — %s (%s, %s)\n", d.Name, inr(d.Amount), ordinalDay(d.DeductDay), day)
	}
	return b.String()
}

func formatAutoPayAdded(a *response.AutoPayResponse) string {
	return fmt.Sprintf("✅ Added %s — %s on the %s (%s).", a.Name, inr(a.Amount), ordinalDay(a.DeductDay), a.Type)
}

func formatSalarySet(amount float64, day int) string {
	return fmt.Sprintf("✅ Salary set to %s, credited on the %s. Type *committed* to see your free money.", inr(amount), ordinalDay(day))
}

func formatHelp() string {
	return strings.Join([]string{
		"👋 *KharchiBook* — track money on WhatsApp.",
		"",
		"Try:",
		"• `200 sabzi` — log an expense",
		"• `summary` — this month's spend",
		"• `add emi home loan 22000 on 5th`",
		"• `add autopay netflix 649 on 12th`",
		"• `autopay list` — your commitments",
		"• `upcoming` — deductions due soon",
		"• `salary 65000 on 1st`",
		"• `committed` — your free money",
		"• `remove netflix` — delete an autopay",
	}, "\n")
}

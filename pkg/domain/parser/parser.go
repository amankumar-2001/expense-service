// Package parser turns a raw WhatsApp message into a structured Command. It is
// deterministic and rule-based (no AI) per PRD §3.4, English-only. A Claude
// fallback for ambiguous categories is a deliberate future seam (see categorize).
package parser

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/kharchibook/expense-service/enums/autopaytype"
	"github.com/kharchibook/expense-service/enums/expensecategory"
)

// Kind discriminates the parsed command.
type Kind int

const (
	KindUnknown Kind = iota
	KindGreeting
	KindHelp
	KindExpense
	KindSummary
	KindAddAutoPay
	KindAutoPayList
	KindUpcoming
	KindRemove
	KindSalary
	KindCommitted
)

// Command is the structured result of parsing a message. Only the fields
// relevant to Kind are populated.
type Command struct {
	Kind Kind

	// Expense
	Amount   float64
	Category expensecategory.Category
	Note     string

	// Summary
	Month string

	// AutoPay (add / remove)
	Name      string
	Type      autopaytype.Type
	DeductDay int

	// Salary
	SalaryDay int
}

var (
	numberRe  = regexp.MustCompile(`\d+(?:\.\d+)?`)
	onDayRe   = regexp.MustCompile(`(?i)\bon\s+(?:the\s+)?(\d{1,2})(?:st|nd|rd|th)?\b`)
	bareDayRe = regexp.MustCompile(`(?i)\b(\d{1,2})(?:st|nd|rd|th)\b`)
)

// Parse classifies text into a Command. Commands are matched before the generic
// expense path so e.g. "summary" is never read as an expense.
func Parse(text string) Command {
	raw := strings.TrimSpace(text)
	lower := strings.ToLower(raw)

	switch {
	case lower == "":
		return Command{Kind: KindUnknown}
	case isGreeting(lower):
		return Command{Kind: KindGreeting}
	case lower == "help":
		return Command{Kind: KindHelp}
	case lower == "committed":
		return Command{Kind: KindCommitted}
	case lower == "upcoming":
		return Command{Kind: KindUpcoming}
	case lower == "autopay list" || lower == "list" || lower == "autopays":
		return Command{Kind: KindAutoPayList}
	case strings.HasPrefix(lower, "summary"):
		month := strings.TrimSpace(raw[len("summary"):])
		return Command{Kind: KindSummary, Month: month}
	case strings.HasPrefix(lower, "salary"):
		return parseSalary(raw)
	case strings.HasPrefix(lower, "remove"):
		name := strings.TrimSpace(raw[len("remove"):])
		return Command{Kind: KindRemove, Name: name}
	case strings.HasPrefix(lower, "add "):
		return parseAddAutoPay(raw)
	case numberRe.MatchString(lower):
		return parseExpense(raw)
	default:
		return Command{Kind: KindUnknown}
	}
}

func isGreeting(lower string) bool {
	switch lower {
	case "hi", "hello", "hey", "start", "hi!", "hello!":
		return true
	}
	return false
}

// parseExpense pulls the first number as the amount and categorizes the rest.
func parseExpense(raw string) Command {
	loc := numberRe.FindStringIndex(raw)
	if loc == nil {
		return Command{Kind: KindUnknown}
	}
	amount, err := strconv.ParseFloat(raw[loc[0]:loc[1]], 64)
	if err != nil || amount <= 0 {
		return Command{Kind: KindUnknown}
	}
	// The descriptive text is everything except the amount token.
	note := strings.TrimSpace(raw[:loc[0]] + " " + raw[loc[1]:])
	note = strings.Join(strings.Fields(note), " ")
	return Command{
		Kind:     KindExpense,
		Amount:   amount,
		Category: categorize(note),
		Note:     note,
	}
}

// parseSalary handles "salary 65000 on 1st".
func parseSalary(raw string) Command {
	cmd := Command{Kind: KindSalary}
	if m := onDayRe.FindStringSubmatch(raw); m != nil {
		cmd.SalaryDay, _ = strconv.Atoi(m[1])
	} else if m := bareDayRe.FindStringSubmatch(raw); m != nil {
		cmd.SalaryDay, _ = strconv.Atoi(m[1])
	}
	// Amount is the first number that isn't the deduction day.
	rest := onDayRe.ReplaceAllString(raw, " ")
	if m := numberRe.FindString(rest); m != "" {
		cmd.Amount, _ = strconv.ParseFloat(m, 64)
	}
	return cmd
}

// parseAddAutoPay handles "add emi home loan 22000 on 5th" and
// "add autopay netflix 649 on 12th".
func parseAddAutoPay(raw string) Command {
	cmd := Command{Kind: KindAddAutoPay}
	// Drop the leading "add".
	rest := strings.TrimSpace(raw[len("add"):])
	lowerRest := strings.ToLower(rest)

	// Determine the type from the leading keyword and strip it.
	switch {
	case strings.HasPrefix(lowerRest, "emi"):
		cmd.Type = autopaytype.EMI
		rest = rest[len("emi"):]
	case strings.HasPrefix(lowerRest, "autopay"):
		rest = rest[len("autopay"):]
		cmd.Type = "" // inferred from the name below
	case strings.HasPrefix(lowerRest, "subscription"):
		cmd.Type = autopaytype.Subscription
		rest = rest[len("subscription"):]
	case strings.HasPrefix(lowerRest, "insurance"):
		cmd.Type = autopaytype.Insurance
		rest = rest[len("insurance"):]
	case strings.HasPrefix(lowerRest, "sip"):
		cmd.Type = autopaytype.SIP
		rest = rest[len("sip"):]
	default:
		cmd.Type = "" // inferred below
	}
	rest = strings.TrimSpace(rest)

	// Deduction day from "on <Nth>".
	if m := onDayRe.FindStringSubmatch(rest); m != nil {
		cmd.DeductDay, _ = strconv.Atoi(m[1])
		rest = onDayRe.ReplaceAllString(rest, " ")
	}
	// Amount is the first remaining number.
	if loc := numberRe.FindStringIndex(rest); loc != nil {
		cmd.Amount, _ = strconv.ParseFloat(rest[loc[0]:loc[1]], 64)
		rest = rest[:loc[0]] + " " + rest[loc[1]:]
	}
	cmd.Name = strings.Join(strings.Fields(rest), " ")

	if cmd.Type == "" {
		cmd.Type = inferAutoPayType(cmd.Name)
	}
	return cmd
}

// inferAutoPayType guesses a subtype for a bare "add autopay <name>" from the
// name. Defaults to subscription, the common case for WhatsApp-added autopays.
func inferAutoPayType(name string) autopaytype.Type {
	n := strings.ToLower(name)
	switch {
	case containsAny(n, "lic", "insurance", "premium", "policy"):
		return autopaytype.Insurance
	case containsAny(n, "sip", "mutual fund", "mf"):
		return autopaytype.SIP
	case containsAny(n, "loan", "emi"):
		return autopaytype.EMI
	default:
		return autopaytype.Subscription
	}
}

func containsAny(haystack string, needles ...string) bool {
	for _, n := range needles {
		if strings.Contains(haystack, n) {
			return true
		}
	}
	return false
}

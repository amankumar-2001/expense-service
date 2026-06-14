// Package dao holds GORM-mapped database entities (data access objects).
package dao

import "time"

// Expense is a single logged discretionary spend. user_id references the
// auth-service user (learned from the JWT) — there is no cross-database FK.
type Expense struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement"`
	UserID      int64     `gorm:"column:user_id;not null;index:idx_expenses_user_date,priority:1"`
	Amount      float64   `gorm:"column:amount;type:numeric(12,2);not null"`
	Category    string    `gorm:"column:category;not null"`
	Note        string    `gorm:"column:note;not null;default:''"`
	RawText     string    `gorm:"column:raw_text;not null;default:''"`
	ExpenseDate time.Time `gorm:"column:expense_date;type:date;not null;index:idx_expenses_user_date,priority:2"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
}

// TableName overrides the default GORM table name.
func (Expense) TableName() string { return "expenses" }

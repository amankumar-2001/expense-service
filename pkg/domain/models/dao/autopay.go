package dao

import "time"

// AutoPay is a fixed recurring monthly commitment (EMI, subscription, insurance,
// SIP). Only rows with status "active" count toward committed money.
type AutoPay struct {
	ID              int64     `gorm:"column:id;primaryKey;autoIncrement"`
	UserID          int64     `gorm:"column:user_id;not null;index:idx_autopays_user_status,priority:1"`
	Name            string    `gorm:"column:name;not null"`
	Type            string    `gorm:"column:type;not null"`
	Amount          float64   `gorm:"column:amount;type:numeric(12,2);not null"`
	DeductDay       int       `gorm:"column:deduct_day;not null"`
	Source          string    `gorm:"column:source;not null;default:'manual'"`
	Status          string    `gorm:"column:status;not null;default:'active';index:idx_autopays_user_status,priority:2"`
	ConfidenceScore *float64  `gorm:"column:confidence_score"`
	Notes           string    `gorm:"column:notes;not null;default:''"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

// TableName overrides the default GORM table name.
func (AutoPay) TableName() string { return "autopays" }

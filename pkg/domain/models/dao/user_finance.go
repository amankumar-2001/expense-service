package dao

import "time"

// UserFinance stores the per-user salary inputs that drive the free-money
// calculation. Keyed by the auth-service user_id (one row per user).
type UserFinance struct {
	UserID        int64     `gorm:"column:user_id;primaryKey"`
	MonthlySalary float64   `gorm:"column:monthly_salary;type:numeric(12,2);not null;default:0"`
	SalaryDay     int       `gorm:"column:salary_day;not null;default:1"`
	UpdatedAt     time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

// TableName overrides the default GORM table name.
func (UserFinance) TableName() string { return "user_finance" }

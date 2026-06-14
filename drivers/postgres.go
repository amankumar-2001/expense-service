// Package drivers wires up external infrastructure clients (Postgres, Redis).
package drivers

import (
	"fmt"

	"github.com/kharchibook/expense-service/config"
	"github.com/kharchibook/expense-service/pkg/domain/models/dao"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewPostgres opens a GORM Postgres connection, configures the pool, and
// optionally runs AutoMigrate (dev/local convenience only).
func NewPostgres(cfg config.Store) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	if cfg.AutoMigrate {
		if err := autoMigrate(db); err != nil {
			return nil, err
		}
	}

	return db, nil
}

// autoMigrate creates/updates tables to match the DAO models. Production uses the
// versioned DDL under ddl/postgresql instead (autoMigrate=false).
func autoMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&dao.Expense{},
		&dao.AutoPay{},
		&dao.UserFinance{},
	)
	if err != nil {
		return fmt.Errorf("auto-migrate: %w", err)
	}
	return nil
}

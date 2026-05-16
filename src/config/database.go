package config

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewDBConnection() (*gorm.DB, error) {
	driver := GetEnvOr("DB_CONNECTION", "sqlite")

	switch driver {
	case "sqlite":
		return newSQLite()
	case "pgsql", "postgres", "postgresql":
		return newPostgres()
	default:
		return nil, fmt.Errorf("unsupported DB_CONNECTION %q — supported: sqlite, pgsql", driver)
	}
}

func newSQLite() (*gorm.DB, error) {
	dsn := GetEnvOr("DB_DATABASE", "maoi.db")
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, err
	}
	// SQLite supports only one writer at a time.
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	return db, nil
}

func newPostgres() (*gorm.DB, error) {
	host := GetEnvOr("DB_HOST", "127.0.0.1")
	port := GetEnvOr("DB_PORT", "5432")
	name := GetEnvOr("DB_DATABASE", "maoi")
	user := GetEnvOr("DB_USERNAME", "postgres")
	pass := GetEnv("DB_PASSWORD")
	tz := GetEnvOr("DB_TIMEZONE", "Asia/Jakarta")

	dsn := fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s TimeZone=%s sslmode=disable",
		host, port, name, user, pass, tz,
	)
	return gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
}

package gorm

import (
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DB = gorm.DB

type Config struct {
	DSN    string
	Logger Logger
}

func NewMySQL(c Config) (*gorm.DB, error) {
	l := logger.Default.LogMode(logger.Info)
	if c.Logger != nil {
		l = c.Logger
	}

	return gorm.Open(mysql.Open(c.DSN), &gorm.Config{
		Logger:      l,
		PrepareStmt: true,
	})
}

func NewPostgreSQL(c Config) (*gorm.DB, error) {
	l := logger.Default.LogMode(logger.Info)
	if c.Logger != nil {
		l = c.Logger
	}

	return gorm.Open(postgres.Open(c.DSN), &gorm.Config{
		Logger:      l,
		PrepareStmt: true,
	})
}

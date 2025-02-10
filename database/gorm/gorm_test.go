package gorm

import (
	"context"
	"testing"

	"github.com/sailwith/x/logger"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

const dsn = "root:root@tcp(127.0.0.1:3306)/db?charset=utf8mb4&parseTime=True&loc=Local"

func TestNew(t *testing.T) {
	db, err := New(Config{
		DSN: dsn,
	})
	if assert.NoError(t, err) {
		t.Log(db.Name())
		printSQL(db)
	}
}

func TestNewWithLogger(t *testing.T) {
	cl, err := newCustomLogger()
	assert.NoError(t, err)
	db, err := New(Config{
		DSN:    dsn,
		Logger: cl,
	})
	if assert.NoError(t, err) {
		t.Log(db.Name())
		printSQL(db)
	}
}

func printSQL(db *gorm.DB) {
	type user struct {
		ID int
	}
	var u user
	ctx := context.Background()
	ctx = logger.SetTraceID(ctx, "123456")
	db.WithContext(ctx).Raw("select * from user where id = 1").Scan(&u)
}

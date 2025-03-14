package gorm

import (
	"context"
	"testing"

	"github.com/sailxy/x/logger"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

const mySQLDSN = "root:root@tcp(127.0.0.1:3306)/mysql?charset=utf8mb4&parseTime=True&loc=Local"
const postgreSQLDSN = "host=localhost user=postgres password=postgres dbname=user_service port=5432 sslmode=disable TimeZone=Asia/Shanghai"

func TestNewMySQL(t *testing.T) {
	db, err := NewMySQL(Config{
		DSN: mySQLDSN,
	})
	if assert.NoError(t, err) {
		t.Log(db.Name())
		printSQL(db)
	}
}

func TestNewMySQLWithLogger(t *testing.T) {
	cl, err := newCustomLogger()
	assert.NoError(t, err)
	db, err := NewMySQL(Config{
		DSN:    mySQLDSN,
		Logger: cl,
	})
	if assert.NoError(t, err) {
		t.Log(db.Name())
		printSQL(db)
	}
}

func TestNewPostgreSQL(t *testing.T) {
	db, err := NewPostgreSQL(Config{
		DSN: postgreSQLDSN,
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
	db.WithContext(ctx).Raw("select * from users where id = 1").Scan(&u)
}

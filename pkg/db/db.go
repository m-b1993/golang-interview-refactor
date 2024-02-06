package db

import (
	"gorm.io/gorm"

	"interview/pkg/entity"
	"interview/pkg/log"
)

// DB represents a DB connection that can be used to run SQL queries.
type DB struct {
	db     *gorm.DB
	logger log.Logger
}

// New returns a new DB connection that wraps the given dbx.DB instance.
func New(db *gorm.DB, logger log.Logger) *DB {
	// connect to the database
	return &DB{db, logger}
}

// DB returns the db.DB wrapped by this object.
func (db *DB) DB() *gorm.DB {
	return db.db
}

func (db *DB) MigrateDatabase() error {
	return db.db.AutoMigrate(&entity.CartEntity{}, &entity.CartItem{})
}

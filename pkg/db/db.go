package db

import (
	"context"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"interview/pkg/entity"
	"interview/pkg/log"
)

type contextKey int

const (
	txKey contextKey = iota
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

// With returns a Builder that can be used to build and execute SQL queries.
// With will return the transaction if it is found in the given context.
// Otherwise it will return a DB connection associated with the context.
func (db *DB) With(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey).(*gorm.DB); ok {
		return tx
	}
	return db.db.WithContext(ctx)
}

// Transactional starts a transaction and calls the given function with a context storing the transaction.
// The transaction associated with the context can be accesse via With().
func (db *DB) Transactional(ctx context.Context, f func(ctx context.Context) error) error {
	return db.db.Transaction(func(tx *gorm.DB) error {
		return f(context.WithValue(ctx, txKey, tx))
	})
}

// TransactionHandler returns a middleware that starts a transaction.
// The transaction started is kept in the context and can be accessed via With().
func (db *DB) TransactionHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		db.db.Transaction(func(tx *gorm.DB) error {
			ctx := context.WithValue(c.Request.Context(), txKey, tx)
			c.Request = c.Request.WithContext(ctx)
			c.Next()
			return nil
		})
	}
}

func (db *DB) MigrateDatabase() error {
	return db.db.AutoMigrate(&entity.CartEntity{}, &entity.CartItem{})
}

package cart

import (
	"interview/pkg/db"
	"interview/pkg/log"
)

type Repository interface {
}

type repository struct {
	db     *db.DB
	logger log.Logger
}

// NewRepository creates a new password repository
func NewRepository(db *db.DB, logger log.Logger) Repository {
	return repository{db, logger}
}

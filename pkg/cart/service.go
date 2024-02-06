package cart

import (
	"interview/pkg/db"
	"interview/pkg/log"
)

type Service interface {
}

type service struct {
	db     *db.DB
	repo   Repository
	logger log.Logger
}

func NewRepository(db *dbcontext.DB, logger log.Logger) Repository {
	return repository{db, logger}
}

package router

import (
	"interview/pkg/db"

	"interview/pkg/cart"
	"interview/pkg/log"

	"github.com/gin-gonic/gin"
)

type routes struct {
	router *gin.Engine
}

func New(router *gin.Engine) *routes {
	return &routes{
		router: router,
	}
}

func (r *routes) RegisterHandlers(logger log.Logger, db *db.DB) {
	cartService := cart.NewService(db, logger)
	cart.RegisterHandlers(r.router.Group("/cart"), cartService, logger)
}

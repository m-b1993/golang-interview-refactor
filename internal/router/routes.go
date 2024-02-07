package router

import (
	"interview/pkg/db"

	"interview/internal/middlewares"
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
	r.router.Use(middlewares.SessionMiddleware(logger))
	cartRepo := cart.NewRepository(db, logger)
	cartService := cart.NewService(cartRepo, logger)
	cart.RegisterHandlers(r.router.Group(cart.CartPath), cartService, logger)
}

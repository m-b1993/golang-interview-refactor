package router

import (
	"interview/pkg/db"

	"github.com/gin-gonic/gin"
	"interview/pkg/controllers"
	"interview/pkg/log"
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
	var taxController controllers.TaxController
	r.router.GET("/", taxController.ShowAddItemForm)
	r.router.POST("/add-item", taxController.AddItem)
	r.router.GET("/remove-cart-item", taxController.DeleteCartItem)
}

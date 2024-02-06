package cart

import (
	"errors"
	"interview/pkg/log"
	"net/http"

	"time"

	"github.com/gin-gonic/gin"
)

func RegisterHandlers(r *gin.RouterGroup, service Service, logger log.Logger) {
	res := resource{service, logger}

	r.GET("/", res.showAddItemForm())
	r.POST("/add", res.addItem())
	r.GET("/remove", res.deleteItem())
}

type resource struct {
	service Service
	logger  log.Logger
}

func (r *resource) showAddItemForm() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, err := ctx.Request.Cookie("ice_session_id")
		if errors.Is(err, http.ErrNoCookie) {
			ctx.SetCookie("ice_session_id", time.Now().String(), 3600, "/", "localhost", false, true)
		}
		r.service.GetCartData(ctx)
	}
}

func (r *resource) addItem() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		cookie, err := ctx.Request.Cookie("ice_session_id")

		if err != nil || errors.Is(err, http.ErrNoCookie) || (cookie != nil && cookie.Value == "") {
			ctx.Redirect(302, "/")
			return
		}

		r.service.AddItemToCart(ctx)
	}
}

func (r *resource) deleteItem() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		cookie, err := ctx.Request.Cookie("ice_session_id")

		if err != nil || errors.Is(err, http.ErrNoCookie) || (cookie != nil && cookie.Value == "") {
			ctx.Redirect(302, "/")
			return
		}

		r.service.DeleteCartItem(ctx)
	}
}

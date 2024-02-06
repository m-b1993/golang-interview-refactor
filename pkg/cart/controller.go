package cart

import (
	"errors"
	"fmt"
	"interview/pkg/log"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
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
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		data := map[string]interface{}{
			"Error":     c.Query("error"),
			"CartItems": r.service.GetCartItems(ctx),
		}
		html, err := r.service.RenderTemplate(data)
		if err != nil {
			r.logger.Errorf("Failed to render cart template: %s", err)
			c.AbortWithStatus(500)
			return
		}
		c.Header("Content-Type", "text/html")
		c.String(200, html)
	}
}

type cartItemForm struct {
	Product  string `form:"product"   binding:"required"`
	Quantity string `form:"quantity"  binding:"required"`
}

func (r *resource) addItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		addItemForm, err := r.getCartItemForm(c)
		if err != nil {
			c.Redirect(302, CartPath+"?error="+err.Error())
			return
		}
		quantity, err := strconv.ParseInt(addItemForm.Quantity, 10, 0)
		if err != nil {
			c.Redirect(302, CartPath+"?error="+errors.New("quantity must be a number").Error())
			return
		}
		err = r.service.AddItemToCart(ctx, addItemForm.Product, int(quantity))
		if err != nil {
			c.Redirect(302, CartPath+"?error="+err.Error())
			return
		}
		c.Redirect(302, CartPath)
	}
}

func (r *resource) deleteItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		cartItemIDString := c.Query("cart_item_id")
		err := r.service.DeleteCartItem(ctx, cartItemIDString)
		if err != nil {
			c.Redirect(302, CartPath+"?error="+err.Error())
			return
		}
		c.Redirect(302, CartPath)
	}
}

func (r *resource) getCartItemForm(c *gin.Context) (*cartItemForm, error) {
	if c.Request.Body == nil {
		return nil, fmt.Errorf("body cannot be nil")
	}

	form := &cartItemForm{}

	if err := binding.FormPost.Bind(c.Request, form); err != nil {
		r.logger.Errorf("Error in binding processing cart form data: %s", err)
		return nil, err
	}

	return form, nil
}

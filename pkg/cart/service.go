package cart

import (
	"errors"
	"fmt"
	"html/template"
	"interview/internal/utils"
	"interview/pkg/db"
	"interview/pkg/entity"
	"interview/pkg/log"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gorm.io/gorm"
)

type Service interface {
	GetCartData(c *gin.Context)
	AddItemToCart(c *gin.Context)
	DeleteCartItem(c *gin.Context)
	getCartItemData(sessionID string) (items []map[string]interface{})
	renderTemplate(pageData interface{}) (string, error)
	getCartItemForm(c *gin.Context) (*cartItemForm, error)
}

type service struct {
	db     *db.DB
	repo   Repository
	logger log.Logger
}

func NewService(db *db.DB, logger log.Logger) Service {
	repo := NewRepository(db, logger)
	return service{db, repo, logger}
}

const CartPath = "/cart"

var itemPriceMapping = map[string]float64{
	"shoe":  100,
	"purse": 200,
	"bag":   300,
	"watch": 300,
}

type cartItemForm struct {
	Product  string `form:"product"   binding:"required"`
	Quantity string `form:"quantity"  binding:"required"`
}

func (s service) GetCartData(c *gin.Context) {
	data := map[string]interface{}{
		"Error": c.Query("error"),
		//"cartItems": cartItems,
	}

	cookie, err := c.Request.Cookie("ice_session_id")
	if err == nil {
		data["CartItems"] = s.getCartItemData(cookie.Value)
	}

	html, err := s.renderTemplate(data)
	if err != nil {
		s.logger.Errorf("Failed to render cart template: %s", err)
		c.AbortWithStatus(500)
		return
	}

	c.Header("Content-Type", "text/html")
	c.String(200, html)
}

func (s service) AddItemToCart(c *gin.Context) {
	cookie, _ := c.Request.Cookie("ice_session_id")

	db := s.db.DB()
	var isCartNew bool
	var cartEntity entity.CartEntity
	result := db.Where(fmt.Sprintf("status = '%s' AND session_id = '%s'", entity.CartOpen, cookie.Value)).First(&cartEntity)

	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.Redirect(302, CartPath+"?error=internal error")
			return
		}
		isCartNew = true
		cartEntity = entity.CartEntity{
			SessionID: cookie.Value,
			Status:    entity.CartOpen,
		}
		db.Create(&cartEntity)
	}

	addItemForm, err := s.getCartItemForm(c)
	if err != nil {
		c.Redirect(302, CartPath+"?error="+err.Error())
		return
	}

	item, ok := itemPriceMapping[addItemForm.Product]
	if !ok {
		c.Redirect(302, CartPath+"?error=invalid item name")
		return
	}

	quantity, err := strconv.ParseInt(addItemForm.Quantity, 10, 0)
	if err != nil {
		c.Redirect(302, CartPath+"?error=invalid quantity")
		return
	}

	var cartItemEntity entity.CartItem
	if isCartNew {
		cartItemEntity = entity.CartItem{
			CartID:      cartEntity.ID,
			ProductName: addItemForm.Product,
			Quantity:    int(quantity),
			Price:       item * float64(quantity),
		}
		db.Create(&cartItemEntity)
	} else {
		result = db.Where(" cart_id = ? and product_name  = ?", cartEntity.ID, addItemForm.Product).First(&cartItemEntity)

		if result.Error != nil {
			if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
				c.Redirect(302, CartPath+"?error=internal error")
				return
			}
			cartItemEntity = entity.CartItem{
				CartID:      cartEntity.ID,
				ProductName: addItemForm.Product,
				Quantity:    int(quantity),
				Price:       item * float64(quantity),
			}
			db.Create(&cartItemEntity)

		} else {
			cartItemEntity.Quantity += int(quantity)
			cartItemEntity.Price += item * float64(quantity)
			db.Save(&cartItemEntity)
		}
	}

	c.Redirect(302, CartPath)
}

func (s service) DeleteCartItem(c *gin.Context) {
	cartItemIDString := c.Query("cart_item_id")
	if cartItemIDString == "" {
		c.Redirect(302, CartPath)
		return
	}

	cookie, _ := c.Request.Cookie("ice_session_id")

	db := s.db.DB()

	var cartEntity entity.CartEntity
	result := db.Where(fmt.Sprintf("status = '%s' AND session_id = '%s'", entity.CartOpen, cookie.Value)).First(&cartEntity)
	if result.Error != nil {
		c.Redirect(302, CartPath+"?error=internal error")
		return
	}

	if cartEntity.Status == entity.CartClosed {
		c.Redirect(302, CartPath)
		return
	}

	cartItemID, err := strconv.Atoi(cartItemIDString)
	if err != nil {
		c.Redirect(302, CartPath+"?error=invalid cart item id")
		return
	}

	var cartItemEntity entity.CartItem

	result = db.Where(" ID  = ?", cartItemID).First(&cartItemEntity)
	if result.Error != nil {
		c.Redirect(302, CartPath+"?error=internal error")
		return
	}

	db.Delete(&cartItemEntity)
	c.Redirect(302, CartPath)
}

func (s service) getCartItemData(sessionID string) (items []map[string]interface{}) {
	db := s.db.DB()
	var cartEntity entity.CartEntity
	result := db.Where(fmt.Sprintf("status = '%s' AND session_id = '%s'", entity.CartOpen, sessionID)).First(&cartEntity)

	if result.Error != nil {
		return
	}

	var cartItems []entity.CartItem
	result = db.Where(fmt.Sprintf("cart_id = %d", cartEntity.ID)).Find(&cartItems)
	if result.Error != nil {
		return
	}

	for _, cartItem := range cartItems {
		item := map[string]interface{}{
			"ID":       cartItem.ID,
			"Quantity": cartItem.Quantity,
			"Price":    cartItem.Price,
			"Product":  cartItem.ProductName,
		}

		items = append(items, item)
	}
	return items
}

func (s service) renderTemplate(pageData interface{}) (string, error) {
	// Read and parse the HTML template file
	templatesDir := utils.GetTemplatesDir()
	templatePath := filepath.Join(templatesDir, "add_item_form.html")
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return "", fmt.Errorf("Error parsing template: %v ", err)
	}

	// Create a strings.Builder to store the rendered template
	var renderedTemplate strings.Builder

	err = tmpl.Execute(&renderedTemplate, pageData)
	if err != nil {
		return "", fmt.Errorf("Error parsing template: %v ", err)
	}

	// Convert the rendered template to a string
	resultString := renderedTemplate.String()

	return resultString, nil
}

func (s service) getCartItemForm(c *gin.Context) (*cartItemForm, error) {
	if c.Request.Body == nil {
		return nil, fmt.Errorf("body cannot be nil")
	}

	form := &cartItemForm{}

	if err := binding.FormPost.Bind(c.Request, form); err != nil {
		s.logger.Errorf("Error in binding processing cart form data: %s", err)
		return nil, err
	}

	return form, nil
}

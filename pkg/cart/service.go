package cart

import (
	"context"
	"errors"
	"fmt"
	"interview/pkg/db"
	"interview/pkg/entity"
	"interview/pkg/log"
	"strconv"

	"gorm.io/gorm"
)

type Service interface {
	AddItemToCart(ctx context.Context, product string, qty int) error
	DeleteCartItem(ctx context.Context, cartItemIDString string) error
	GetCartItems(ctx context.Context) []map[string]interface{}
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

func (s service) GetCartItems(ctx context.Context) (items []map[string]interface{}) {
	db := s.db.DB()
	var cartEntity entity.CartEntity
	sessionID := ctx.Value("SessionId").(string)
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

func (s service) AddItemToCart(ctx context.Context, product string, qty int) error {
	sessionID := ctx.Value("SessionId").(string)

	db := s.db.DB()
	var isCartNew bool
	var cartEntity entity.CartEntity
	result := db.Where(fmt.Sprintf("status = '%s' AND session_id = '%s'", entity.CartOpen, sessionID)).First(&cartEntity)

	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return errors.New("internal error")
		}
		isCartNew = true
		cartEntity = entity.CartEntity{
			SessionID: sessionID,
			Status:    entity.CartOpen,
		}
		db.Create(&cartEntity)
	}

	item, ok := itemPriceMapping[product]
	if !ok {
		return errors.New("invalid item name")
	}

	var cartItemEntity entity.CartItem
	if isCartNew {
		cartItemEntity = entity.CartItem{
			CartID:      cartEntity.ID,
			ProductName: product,
			Quantity:    qty,
			Price:       item * float64(qty),
		}
		db.Create(&cartItemEntity)
	} else {
		result = db.Where(" cart_id = ? and product_name  = ?", cartEntity.ID, product).First(&cartItemEntity)

		if result.Error != nil {
			if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return errors.New("internal error")
			}
			cartItemEntity = entity.CartItem{
				CartID:      cartEntity.ID,
				ProductName: product,
				Quantity:    qty,
				Price:       item * float64(qty),
			}
			db.Create(&cartItemEntity)

		} else {
			cartItemEntity.Quantity += int(qty)
			cartItemEntity.Price += item * float64(qty)
			db.Save(&cartItemEntity)
		}
	}

	return nil
}

func (s service) DeleteCartItem(ctx context.Context, cartItemIDString string) error {
	if cartItemIDString == "" {
		return nil
	}

	sessionID := ctx.Value("SessionId").(string)

	db := s.db.DB()

	var cartEntity entity.CartEntity
	result := db.Where(fmt.Sprintf("status = '%s' AND session_id = '%s'", entity.CartOpen, sessionID)).First(&cartEntity)
	if result.Error != nil {
		return errors.New("internal error")
	}

	if cartEntity.Status == entity.CartClosed {
		return nil
	}

	cartItemID, err := strconv.Atoi(cartItemIDString)
	if err != nil {
		return errors.New("invalid cart item id")
	}

	var cartItemEntity entity.CartItem

	result = db.Where(" ID  = ?", cartItemID).First(&cartItemEntity)
	if result.Error != nil {
		return errors.New("internal error")
	}

	db.Delete(&cartItemEntity)
	return nil
}

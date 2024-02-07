package cart

import (
	"context"
	"errors"
	"interview/pkg/entity"
	"interview/pkg/log"

	"gorm.io/gorm"
)

type Service interface {
	AddItemToCart(ctx context.Context, product string, qty int) error
	DeleteCartItem(ctx context.Context, cartItemID uint) error
	GetCartItems(ctx context.Context) []map[string]interface{}
}

type service struct {
	repo   Repository
	logger log.Logger
}

func NewService(repo Repository, logger log.Logger) Service {
	return service{repo, logger}
}

const CartPath = "/cart"

var itemPriceMapping = map[string]float64{
	"shoe":  100,
	"purse": 200,
	"bag":   300,
	"watch": 300,
}

func (s service) GetCartItems(ctx context.Context) (items []map[string]interface{}) {
	sessionID := ctx.Value("SessionId").(string)
	conditions := map[string]interface{}{
		"status":     entity.CartOpen,
		"session_id": sessionID,
	}
	cartEntities, err := s.repo.QueryCart(ctx, conditions, "id desc", 1, 0)
	if err != nil || len(cartEntities) == 0 {
		return
	}
	cartEntity := cartEntities[0]

	conditions = map[string]interface{}{
		"cart_id": cartEntity.ID,
	}
	cartItems, err := s.repo.QueryCartItem(ctx, conditions, "id desc", 100, 0)
	if err != nil {
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

	var isCartNew bool
	var cartEntity entity.CartEntity

	conditions := map[string]interface{}{
		"status":     entity.CartOpen,
		"session_id": sessionID,
	}
	cartEntities, err := s.repo.QueryCart(ctx, conditions, "id desc", 1, 0)

	if err != nil || len(cartEntities) == 0 {
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Errorf("error querying cart: %v", err)
			return errors.New("internal error")
		}
		isCartNew = true
		cartEntity = entity.CartEntity{
			SessionID: sessionID,
			Status:    entity.CartOpen,
		}
		s.repo.CreateCart(ctx, &cartEntity)
	} else {
		cartEntity = cartEntities[0]
	}

	item, ok := itemPriceMapping[product]
	if !ok {
		return errors.New("invalid item name")
	}
	subTotal := item * float64(qty)

	var cartItemEntity entity.CartItem
	if isCartNew {
		cartItemEntity = entity.CartItem{
			CartID:      cartEntity.ID,
			ProductName: product,
			Quantity:    qty,
			Price:       subTotal,
		}
		s.repo.CreateCartItem(ctx, &cartItemEntity)
	} else {
		conditions = map[string]interface{}{
			"cart_id":      cartEntity.ID,
			"product_name": product,
		}
		cartItems, err := s.repo.QueryCartItem(ctx, conditions, "id desc", 1, 0)
		if err != nil || len(cartItems) == 0 {
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				s.logger.Errorf("error querying cart item: %v", err)
				return errors.New("internal error")
			}
			cartItemEntity = entity.CartItem{
				CartID:      cartEntity.ID,
				ProductName: product,
				Quantity:    qty,
				Price:       subTotal,
			}
			s.repo.CreateCartItem(ctx, &cartItemEntity)
		} else {
			cartItemEntity = cartItems[0]
			cartItemEntity.Quantity += int(qty)
			cartItemEntity.Price += subTotal
			s.repo.UpdateCartItem(ctx, &cartItemEntity)
		}
	}
	cartEntity.Total += subTotal
	s.repo.UpdateCart(ctx, &cartEntity)

	return nil
}

func (s service) DeleteCartItem(ctx context.Context, cartItemID uint) error {
	sessionID := ctx.Value("SessionId").(string)

	var cartEntity entity.CartEntity
	conditions := map[string]interface{}{
		"status":     entity.CartOpen,
		"session_id": sessionID,
	}
	cartEntities, err := s.repo.QueryCart(ctx, conditions, "id desc", 1, 0)
	if err != nil || len(cartEntities) == 0 {
		s.logger.Errorf("error querying cart: %v", err)
		return errors.New("internal error")
	}
	cartEntity = cartEntities[0]

	if cartEntity.Status == entity.CartClosed {
		return nil
	}

	var cartItemEntity entity.CartItem

	conditions = map[string]interface{}{
		"ID": cartItemID,
	}
	cartItems, err := s.repo.QueryCartItem(ctx, conditions, "id desc", 1, 0)
	if err != nil || len(cartItems) == 0 {
		s.logger.Errorf("error querying cart item: %v", err)
		return errors.New("internal error")
	}

	cartItemEntity = cartItems[0]
	if cartItemEntity.CartID != cartEntity.ID {
		return errors.New("invalid cart item id")
	}

	s.repo.DeleteCartItem(ctx, &cartItemEntity)

	return nil
}

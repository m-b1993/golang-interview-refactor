package cart

import (
	"context"
	"errors"
	"interview/pkg/entity"
	"interview/pkg/log"
)

type Service interface {
	AddItemToCart(ctx context.Context, product string, qty int) error
	DeleteCartItem(ctx context.Context, cartItemID uint) error
	GetCartItems(ctx context.Context) []map[string]interface{}
	GetProducts() []string
	getCart(ctx context.Context) (entity.CartEntity, error)
	getOrCreateCart(ctx context.Context) (entity.CartEntity, bool, error)
}

type service struct {
	repo   Repository
	logger log.Logger
}

var CartNotFoundError = errors.New("cart not found")
var InternalError = errors.New("internal error")

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
	cartEntity, err := s.getCart(ctx)
	if err != nil {
		s.logger.Errorf("error getting cart: %v", err)
		return
	}
	conditions := map[string]interface{}{
		"cart_id": cartEntity.ID,
	}
	cartItems, err := s.repo.QueryCartItem(ctx, conditions, "id desc", 100, 0)
	if err != nil {
		s.logger.Errorf("error querying cart items: %v", err)
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
	cartEntity, isCartNew, err := s.getOrCreateCart(ctx)
	if err != nil {
		return err
	}
	item, ok := itemPriceMapping[product]
	if !ok {
		return errors.New("invalid item name")
	}
	subTotal := item * float64(qty)

	createItem := func() error {
		cartItemEntity := entity.CartItem{
			CartID:      cartEntity.ID,
			ProductName: product,
			Quantity:    qty,
			Price:       subTotal,
		}
		return s.repo.CreateCartItem(ctx, &cartItemEntity)
	}

	err = nil
	if isCartNew {
		err = createItem()
	} else {
		conditions := map[string]interface{}{
			"cart_id":      cartEntity.ID,
			"product_name": product,
		}
		cartItems, err := s.repo.QueryCartItem(ctx, conditions, "id desc", 1, 0)
		if err != nil {
			s.logger.Errorf("error querying cart item: %v", err)
			return InternalError
		}
		if len(cartItems) == 0 {
			err = createItem()
		} else {
			cartItemEntity := cartItems[0]
			cartItemEntity.Quantity += qty
			cartItemEntity.Price += subTotal
			err = s.repo.UpdateCartItem(ctx, &cartItemEntity)
		}
	}
	if err != nil {
		s.logger.Errorf("error adding item to cart: %v", err)
		return InternalError
	}

	cartEntity.Total += subTotal
	err = s.repo.UpdateCart(ctx, &cartEntity)
	if err != nil {
		s.logger.Errorf("error updating cart: %v", err)
		return InternalError
	}

	return nil
}

func (s service) DeleteCartItem(ctx context.Context, cartItemID uint) error {
	cartEntity, err := s.getCart(ctx)
	if err != nil {
		s.logger.Errorf("error getting cart: %v", err)
		return InternalError
	}

	conditions := map[string]interface{}{
		"ID":      cartItemID,
		"cart_id": cartEntity.ID,
	}
	err = s.repo.DeleteCartItem(ctx, conditions)
	if err != nil {
		s.logger.Errorf("error deleting cart item: %v", err)
		return InternalError
	}

	return nil
}

func (s service) GetProducts() []string {
	var products []string
	for product := range itemPriceMapping {
		products = append(products, product)
	}
	return products
}

func (s service) getCart(ctx context.Context) (entity.CartEntity, error) {
	sessionID := ctx.Value("SessionId").(string)
	conditions := map[string]interface{}{
		"status":     entity.CartOpen,
		"session_id": sessionID,
	}
	cartEntities, err := s.repo.QueryCart(ctx, conditions, "id desc", 1, 0)
	if err != nil {
		return entity.CartEntity{}, err
	}
	if len(cartEntities) == 0 {
		return entity.CartEntity{}, CartNotFoundError
	}
	return cartEntities[0], nil
}

func (s service) getOrCreateCart(ctx context.Context) (entity.CartEntity, bool, error) {
	sessionID := ctx.Value("SessionId").(string)
	created := false
	cartEntity, err := s.getCart(ctx)
	if err != nil && !errors.Is(err, CartNotFoundError) {
		return entity.CartEntity{}, false, err
	}
	if errors.Is(err, CartNotFoundError) {
		cartEntity = entity.CartEntity{
			SessionID: sessionID,
			Status:    entity.CartOpen,
		}
		s.repo.CreateCart(ctx, &cartEntity)
		created = true
	}
	return cartEntity, created, nil
}

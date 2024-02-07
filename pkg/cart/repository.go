package cart

import (
	"context"
	"interview/pkg/db"
	"interview/pkg/entity"
	"interview/pkg/log"
)

type Repository interface {
	QueryCart(ctx context.Context, conditions map[string]interface{}, order string, limit int, offset int) ([]entity.CartEntity, error)
	QueryCartItem(ctx context.Context, conditions map[string]interface{}, order string, limit int, offset int) ([]entity.CartItem, error)
	CreateCart(ctx context.Context, cartEntity *entity.CartEntity) error
	CreateCartItem(ctx context.Context, cartItem *entity.CartItem) error
	UpdateCart(ctx context.Context, cartEntity *entity.CartEntity) error
	UpdateCartItem(ctx context.Context, cartItem *entity.CartItem) error
	DeleteCart(ctx context.Context, cartEntity *entity.CartEntity) error
	DeleteCartItem(ctx context.Context, cartItem *entity.CartItem) error
}

type repository struct {
	db     *db.DB
	logger log.Logger
}

func NewRepository(db *db.DB, logger log.Logger) Repository {
	return repository{db, logger}
}

func (r repository) QueryCart(ctx context.Context, conditions map[string]interface{}, order string, limit int, offset int) ([]entity.CartEntity, error) {
	var cartEntities []entity.CartEntity
	db := r.db.With(ctx)
	result := db.Where(conditions).
		Order(order).
		Limit(limit).
		Offset(offset).
		Find(&cartEntities)
	if result.Error != nil {
		return nil, result.Error
	}
	return cartEntities, nil
}

func (r repository) QueryCartItem(ctx context.Context, conditions map[string]interface{}, order string, limit int, offset int) ([]entity.CartItem, error) {
	var cartItems []entity.CartItem
	db := r.db.With(ctx)
	result := db.Where(conditions).
		Order(order).
		Limit(limit).
		Offset(offset).
		Find(&cartItems)
	if result.Error != nil {
		return nil, result.Error
	}
	return cartItems, nil
}

func (r repository) CreateCart(ctx context.Context, cartEntity *entity.CartEntity) error {
	db := r.db.With(ctx)
	result := db.Create(cartEntity)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r repository) CreateCartItem(ctx context.Context, cartItem *entity.CartItem) error {
	db := r.db.With(ctx)
	result := db.Create(cartItem)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r repository) UpdateCart(ctx context.Context, cartEntity *entity.CartEntity) error {
	db := r.db.With(ctx)
	result := db.Save(cartEntity)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r repository) UpdateCartItem(ctx context.Context, cartItem *entity.CartItem) error {
	db := r.db.With(ctx)
	result := db.Save(cartItem)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r repository) DeleteCart(ctx context.Context, cartEntity *entity.CartEntity) error {
	db := r.db.With(ctx)
	result := db.Delete(cartEntity)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r repository) DeleteCartItem(ctx context.Context, cartItem *entity.CartItem) error {
	db := r.db.With(ctx)
	result := db.Delete(cartItem)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

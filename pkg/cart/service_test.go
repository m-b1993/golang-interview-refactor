package cart

import (
	"context"
	"interview/pkg/entity"
	"interview/pkg/log"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

const sessionID = "123456789"

var expected = []map[string]interface{}{
	{
		"ID":       uint(1),
		"Quantity": 3,
		"Price":    float64(300),
		"Product":  "shoe",
	},
	{
		"ID":       uint(2),
		"Quantity": 1,
		"Price":    float64(200),
		"Product":  "purse",
	},
}

type mockCartRepo struct {
	cards []entity.CartEntity
	items []entity.CartItem
}

func Test_service_GetCartItems(t *testing.T) {
	logger, _ := log.NewForTest()
	repo := getMockedRepo()
	service := NewService(&repo, logger)
	ctx := context.WithValue(context.Background(), "SessionId", sessionID)
	got := service.GetCartItems(ctx)
	assert.Equal(t, expected, got)
}

func Test_service_AddItemToCart(t *testing.T) {
	logger, _ := log.NewForTest()
	repo := getMockedRepo()
	service := NewService(&repo, logger)
	ctx := context.WithValue(context.Background(), "SessionId", sessionID)

	qty := 2
	product := "watch"
	err := service.AddItemToCart(ctx, product, qty)
	assert.Nil(t, err)
	assert.Equal(t, 4, len(repo.items))
	expected := append(expected, map[string]interface{}{
		"ID":       uint(4),
		"Quantity": qty,
		"Price":    itemPriceMapping[product] * float64(qty),
		"Product":  product,
	})
	got := service.GetCartItems(ctx)
	assert.Equal(t, expected, got)

	assert.Equal(t, float64(1100), repo.cards[0].Total)
}

func Test_service_DeleteCartItem(t *testing.T) {
	logger, _ := log.NewForTest()
	repo := getMockedRepo()
	service := NewService(&repo, logger)
	ctx := context.WithValue(context.Background(), "SessionId", sessionID)
	err := service.DeleteCartItem(ctx, 1)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(repo.items))
	expected := expected[1:]
	got := service.GetCartItems(ctx)
	assert.Equal(t, expected, got)
}

func getMockedRepo() mockCartRepo {
	carts := []entity.CartEntity{
		{
			Model:     gorm.Model{ID: 1},
			SessionID: sessionID,
			Status:    entity.CartOpen,
			Total:     500,
		},
		{
			Model:     gorm.Model{ID: 2},
			SessionID: "987654321",
			Status:    entity.CartOpen,
			Total:     300,
		},
	}
	items := []entity.CartItem{
		{
			Model:       gorm.Model{ID: 1},
			CartID:      1,
			ProductName: "shoe",
			Quantity:    3,
			Price:       itemPriceMapping["shoe"] * 3,
		},
		{
			Model:       gorm.Model{ID: 2},
			CartID:      1,
			ProductName: "purse",
			Quantity:    1,
			Price:       itemPriceMapping["purse"],
		},
		{
			Model:       gorm.Model{ID: 3},
			CartID:      2,
			ProductName: "bag",
			Quantity:    1,
			Price:       itemPriceMapping["bag"],
		},
	}
	repo := mockCartRepo{
		cards: carts,
		items: items,
	}
	return repo
}

func (m *mockCartRepo) QueryCart(ctx context.Context, conditions map[string]interface{}, order string, limit int, offset int) ([]entity.CartEntity, error) {
	var carts []entity.CartEntity
	for _, c := range m.cards {
		matched := true
		for k, v := range conditions {
			if k == "session_id" && c.SessionID == v.(string) {
				continue
			}
			if k == "status" && c.Status == v.(entity.Status) {
				continue
			}
			matched = false
		}
		if matched {
			carts = append(carts, c)
		}
	}
	return carts, nil
}

func (m *mockCartRepo) QueryCartItem(ctx context.Context, conditions map[string]interface{}, order string, limit int, offset int) ([]entity.CartItem, error) {
	var items []entity.CartItem
	for _, c := range m.items {
		matched := true
		for k, v := range conditions {
			if k == "ID" && c.CartID == v.(uint) {
				continue
			}
			if k == "cart_id" && c.CartID == v.(uint) {
				continue
			}
			matched = false
		}
		if matched {
			items = append(items, c)
		}
	}
	return items, nil
}

func (m *mockCartRepo) CreateCart(ctx context.Context, cartEntity *entity.CartEntity) error {
	cartEntity.ID = uint(len(m.cards) + 1)
	m.cards = append(m.cards, *cartEntity)
	return nil
}

func (m *mockCartRepo) CreateCartItem(ctx context.Context, cartItem *entity.CartItem) error {
	cartItem.ID = uint(len(m.items) + 1)
	m.items = append(m.items, *cartItem)
	return nil
}

func (m *mockCartRepo) UpdateCart(ctx context.Context, cartEntity *entity.CartEntity) error {
	for i, c := range m.cards {
		if c.ID == cartEntity.ID {
			m.cards[i] = *cartEntity
		}
	}
	return nil
}

func (m *mockCartRepo) UpdateCartItem(ctx context.Context, cartItem *entity.CartItem) error {
	for i, c := range m.items {
		if c.ID == cartItem.ID {
			m.items[i] = *cartItem
		}
	}
	return nil
}

func (m *mockCartRepo) DeleteCartById(ctx context.Context, id uint) error {
	for i, c := range m.cards {
		if c.ID == id {
			m.cards = append(m.cards[:i], m.cards[i+1:]...)
		}
	}
	return nil
}

func (m *mockCartRepo) DeleteCartItemById(ctx context.Context, id uint) error {
	for i, c := range m.items {
		if c.ID == id {
			m.items = append(m.items[:i], m.items[i+1:]...)
		}
	}
	return nil
}

func (m *mockCartRepo) DeleteCart(ctx context.Context, conditions map[string]interface{}) error {
	for k, v := range conditions {
		if k == "id" {
			for i, c := range m.cards {
				if c.ID == v.(uint) {
					m.cards = append(m.cards[:i], m.cards[i+1:]...)
				}
			}
		}
		if k == "status" {
			for i, c := range m.cards {
				if c.Status == v.(entity.Status) {
					m.cards = append(m.cards[:i], m.cards[i+1:]...)
				}
			}
		}
	}
	return nil
}

func (m *mockCartRepo) DeleteCartItem(ctx context.Context, conditions map[string]interface{}) error {
	for k, v := range conditions {
		if k == "id" {
			for i, c := range m.items {
				if c.ID == v.(uint) {
					m.items = append(m.items[:i], m.items[i+1:]...)
				}
			}
		}
		if k == "cart_id" {
			for i, c := range m.items {
				if c.CartID == v.(uint) {
					m.items = append(m.items[:i], m.items[i+1:]...)
				}
			}
		}
	}
	return nil
}

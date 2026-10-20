package repository

import (
	"ordersystem/model"
	"time"
)

type DatabaseHandler struct {
	// drinks represent all available drinks
	drinks []model.Drink
	// orders serves as order history
	orders []model.Order
}

// todo //
func NewDatabaseHandler() *DatabaseHandler {
	// Init the drinks slice with some test data
	drinks := []model.Drink{
		{1, "Beer", 2.00, "Hagenberg Gold"},
		{2, "Spritzer", 1.40, "Wine with soda"},
		{3, "Coffee", 0.00, "Mifare isn't that secure ;)"},
	}

	// Init orders slice with some test data
	orders := []model.Order{
		{1, time.Now(), 25},
		{2, time.Now(), 15},
		{3, time.Now(), 10},
	}

	return &DatabaseHandler{
		drinks: drinks,
		orders: orders,
	}
}

func (db *DatabaseHandler) GetDrinks() []model.Drink {
	return db.drinks
}

func (db *DatabaseHandler) GetOrders() []model.Order {
	return db.orders
}

// todo //
func (db *DatabaseHandler) GetTotalledOrders() map[uint64]uint64 {
	// calculate total orders
	// key = DrinkID, value = Amount of orders
	// totalledOrders map[uint64]uint64
	totalledOrders := make(map[uint64]uint64)
	for _, order := range db.orders {
		totalledOrders[order.DrinkID] += uint64(order.Amount)
	}
	return totalledOrders
}

func (db *DatabaseHandler) AddOrder(order *model.Order) {
	// todo //
	// add order to db.orders slice
	db.orders = append(db.orders, *order)
}

package rest

import (
	"encoding/json"
	"net/http"
	"ordersystem/model"
	"ordersystem/repository"

	"github.com/go-chi/render"
)

// GetMenu 			godoc
// @tags 			Menu
// @Description 	Returns the menu of all drinks
// @Produce  		json
// @Success 		200 {array} model.Drink
// @Router 			/api/menu [get]
func GetMenu(db *repository.DatabaseHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// todo //
		// get slice from db
		menu := db.GetDrinks()
		// render.Status(r, http.StatusOK)
		render.Status(r, http.StatusOK)
		// render.JSON(w, r, <your-slice>)
		render.JSON(w, r, menu)
	}
}

// todo create GetOrders /api/order/all  //
// GetOrders 			godoc
// @tags 			Orders
// @Description 	Returns all orders
// @Produce  		json
// @Success 		200 {array} model.Order
// @Router 			/api/order/all [get]
func GetOrders(db *repository.DatabaseHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orders := db.GetOrders()
		render.Status(r, http.StatusOK)
		render.JSON(w, r, orders)
	}
}

// todo create GetOrdersTotal /api/order/total  //
// GetOrdersTotal 			godoc
// @tags 			OrdersTotal
// @Description 	Returns the amount of each drink
// @Produce  		json
// @Success 		200 {object} map[uint64]uint64
// @Router 			/api/order/total [get]
func GetOrdersTotal(db *repository.DatabaseHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		totalledorders := db.GetTotalledOrders()
		render.Status(r, http.StatusOK)
		render.JSON(w, r, totalledorders)
	}
}

// PostOrder 		godoc
// @tags 			Order
// @Description 	Adds an order to the db
// @Accept 			json
// @Param 			b body model.Order true "Order"
// @Produce  		json
// @Success 		200
// @Failure     	400
// @Router 			/api/order [post]
func PostOrder(db *repository.DatabaseHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// todo
		// declare empty order struct
		newOrder := model.Order{}

		// err := json.NewDecoder(r.Body).Decode(&<your-order-struct>)
		err := json.NewDecoder(r.Body).Decode(&newOrder)

		// handle error and render Status 400
		if err != nil { // if there is an error
			render.Status(r, http.StatusBadRequest) //Status 400
			render.JSON(w, r, err)
			return
		}

		// add to db
		db.AddOrder(&newOrder)

		render.Status(r, http.StatusOK)
		render.JSON(w, r, "ok")

	}
}

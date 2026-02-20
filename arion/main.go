// Package main Recapo API server.
//
//	@title			Recapo API
//	@version		1.0
//	@description	Order management API for Jastipers
//	@description
//	@description	All responses use this envelope:
//	@description	{ "success": bool, "data": <payload>, "code": string, "message": string }
//	@description	Success: success=true, data holds the payload. Error: success=false, data={}, code and message describe the error.
//	@host			localhost:4000
//	@BasePath		/
//	@schemes		http
//
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"

	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/zeirash/recapo/arion/common/config"
	"github.com/zeirash/recapo/arion/common/database"
	"github.com/zeirash/recapo/arion/common/logger"
	"github.com/zeirash/recapo/arion/common/middleware"
	"github.com/zeirash/recapo/arion/handler"

	_ "github.com/zeirash/recapo/arion/docs" // swagger docs
)

func NewRouter() *mux.Router {
	handler.Init()
	r := mux.NewRouter()

	// Global middleware
	r.Use(middleware.Recovery)

	// Swagger UI (WrapHandler is http.HandlerFunc; doc is served from swag registry)
	r.PathPrefix("/swagger/").HandlerFunc(httpSwagger.WrapHandler)

	// Routes API No Auth
	r.HandleFunc("/health", handler.HealthHandler)
	r.HandleFunc("/public/shops/{share_token}/products", handler.GetShopProductsHandler).Methods("GET")
	r.HandleFunc("/public/shops/{share_token}/order", handler.CreateShopTempOrderHandler).Methods("POST")

	r.HandleFunc("/login", handler.LoginHandler).Methods("POST")
	r.HandleFunc("/register", handler.RegisterHandler).Methods("POST")
	r.HandleFunc("/refresh", handler.RefreshHandler).Methods("POST")

	r.Handle("/test/{id}", middleware.ChainMiddleware()(http.HandlerFunc(handler.HealthHandler)))
	// r.Handle("/test/{id}", middleware.MiddlewareWrapper(http.HandlerFunc(handler.HealthHandler), middleware.Authentication()))

	// For User
	r.Handle("/user", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.UpdateUserHandler))).Methods("PATCH")
	r.Handle("/user", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.GetUserHandler))).Methods("GET")
	r.Handle("/users", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.GetUsersHandler))).Methods("GET")

	// For Customer
	r.Handle("/customer", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.CreateCustomerHandler))).Methods("POST")
	r.Handle("/customers", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.GetCustomersHandler))).Methods("GET")
	r.Handle("/customers/{customer_id}", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.UpdateCustomerHandler))).Methods("PATCH")
	r.Handle("/customers/{customer_id}", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.DeleteCustomerHandler))).Methods("DELETE")
	r.Handle("/customers/{customer_id}", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.GetCustomerHandler))).Methods("GET")
	r.Handle("/customers/check_active_order", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.CustomerCheckActiveOrderHandler))).Methods("POST")

	// For Shop
	r.Handle("/shop/share_token", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.GetShopShareTokenHandler))).Methods("GET")

	// For Product (register literal paths before /products/{product_id} so they match first)
	r.Handle("/product", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.CreateProductHandler))).Methods("POST")
	r.Handle("/products", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.GetProductsHandler))).Methods("GET")
	r.Handle("/products/purchase_list", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.PurchaseListProductHandler))).Methods("GET")
	r.Handle("/products/{product_id}", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.UpdateProductHandler))).Methods("PATCH")
	r.Handle("/products/{product_id}", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.DeleteProductHandler))).Methods("DELETE")
	r.Handle("/products/{product_id}", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.GetProductHandler))).Methods("GET")

	// For Order
	r.Handle("/order", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.CreateOrderHandler))).Methods("POST")
	r.Handle("/orders", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.GetOrdersHandler))).Methods("GET")
	r.Handle("/orders/{order_id}", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.UpdateOrderHandler))).Methods("PATCH")
	r.Handle("/orders/{order_id}", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.DeleteOrderHandler))).Methods("DELETE")
	r.Handle("/orders/{order_id}", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.GetOrderHandler))).Methods("GET")
	r.Handle("/orders/{order_id}/item", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.CreateOrderItemHandler))).Methods("POST")
	r.Handle("/orders/{order_id}/items", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.GetOrderItemsHandler))).Methods("GET")
	r.Handle("/orders/{order_id}/items/{item_id}", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.UpdateOrderItemHandler))).Methods("PATCH")
	r.Handle("/orders/{order_id}/items/{item_id}", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.DeleteOrderItemHandler))).Methods("DELETE")
	r.Handle("/orders/{order_id}/items/{item_id}", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.GetOrderItemHandler))).Methods("GET")

	// For Temp Order
	r.Handle("/temp_orders", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.GetTempOrdersHandler))).Methods("GET")
	r.Handle("/temp_orders/merge", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.MergeTempOrderHandler))).Methods("POST")
	r.Handle("/temp_orders/{temp_order_id}", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.GetTempOrderHandler))).Methods("GET")
	r.Handle("/temp_orders/{temp_order_id}/reject", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.RejectTempOrderHandler))).Methods("PATCH")

	// For System
	r.Handle("/system/user/{user_id}", middleware.ChainMiddleware(middleware.Authentication, middleware.CheckSystemMode)(http.HandlerFunc(handler.GetUserHandler))).Methods("GET")

	return r
}

func main() {
	// init logger
	logger.Init()

	// init config
	config.InitConfig()

	// init database
	database.InitDB()
	defer database.CloseDB()

	// init router
	r := NewRouter()
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:3001"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		Debug:            true,
	})
	handler := c.Handler(r)

	logger.Info("Server starting on port :4000")
	logger.Fatal(http.ListenAndServe(":4000", handler))
}

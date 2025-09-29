package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"

	"github.com/zeirash/recapo/arion/common/config"
	"github.com/zeirash/recapo/arion/common/middleware"
	"github.com/zeirash/recapo/arion/handler"
)

func NewRouter() *mux.Router {
	handler.Init()
	r := mux.NewRouter()

	// Routes API
	r.HandleFunc("/health", handler.HealthHandler)

	r.HandleFunc("/login", handler.LoginHandler).Methods("POST")
	r.HandleFunc("/register", handler.RegisterHandler).Methods("POST")

	r.Handle("/test/{id}", middleware.ChainMiddleware()(http.HandlerFunc(handler.HealthHandler)))
	// r.Handle("/test/{id}", middleware.MiddlewareWrapper(http.HandlerFunc(handler.HealthHandler), middleware.Authentication()))

	// For User
	r.Handle("/user", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.UpdateUserHandler))).Methods("PATCH")
	r.Handle("/user", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.GetUserHandler))).Methods("GET")
	r.Handle("/users", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.GetUsersHandler))).Methods("GET")

	// For Customer
	r.Handle("/customer", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.CreateCustomerHandler))).Methods("POST")
	r.Handle("/customer/{customer_id}", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.UpdateCustomerHandler))).Methods("PATCH")
	r.Handle("/customer/{customer_id}", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.DeleteCustomerHandler))).Methods("DELETE")
	r.Handle("/customer/{customer_id}", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.GetCustomerHandler))).Methods("GET")
	r.Handle("/customers", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.GetCustomersHandler))).Methods("GET")

	// For Product
	r.Handle("/product", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.CreateProductHandler))).Methods("POST")
	r.Handle("/product/{product_id}", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.UpdateProductHandler))).Methods("PATCH")
	r.Handle("/product/{product_id}", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.DeleteProductHandler))).Methods("DELETE")
	r.Handle("/product/{product_id}", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.GetProductHandler))).Methods("GET")
	r.Handle("/products", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.GetProductsHandler))).Methods("GET")

	r.Handle("/product/{product_id}/price", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.CreatePriceHandler))).Methods("POST")
	r.Handle("/product/{product_id}/price/{price_id}", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.UpdateProductPriceHandler))).Methods("PATCH")
	r.Handle("/product/{product_id}/price/{price_id}", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.DeleteProductPriceHandler))).Methods("DELETE")

	// For System
	r.Handle("/system/user/{user_id}", middleware.ChainMiddleware(middleware.Authentication, middleware.CheckSystemMode)(http.HandlerFunc(handler.GetUserHandler))).Methods("GET")

	return r
}

func main() {
	// init config
	config.InitConfig()

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

	fmt.Println("Server run on port :4000")
	log.Fatal(http.ListenAndServe(":4000", handler))
}

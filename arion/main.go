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
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
	})
	handler := c.Handler(r)

	fmt.Println("Server run on port :3000")
	log.Fatal(http.ListenAndServe(":3000", handler))
}

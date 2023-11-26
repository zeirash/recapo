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

	r.Handle("/test", middleware.MiddlewareWrapper(http.HandlerFunc(handler.HealthHandler), middleware.JwtMiddleware()))

	// For User
	r.Handle("/user", middleware.MiddlewareWrapper(http.HandlerFunc(handler.UpdateUserHandler), middleware.JwtMiddleware())).Methods("PATCH")

	// TODO: For System

	return r
}

func main() {
	// init config
	config.InitConfig()

	// init router
	r := NewRouter()
	c := cors.New(cors.Options{
			AllowedOrigins: []string{"*"},
			AllowCredentials: true,
	})
	handler := c.Handler(r)

	fmt.Println("Server run on port :3000")
	log.Fatal(http.ListenAndServe(":3000", handler))
}

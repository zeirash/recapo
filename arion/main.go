package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"

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

	return r
}

func main() {
	// init config
	config.InitConfig()

	// init router
	r := NewRouter()

	fmt.Println("Server run on port :3000")
	log.Fatal(http.ListenAndServe(":3000", r))
}

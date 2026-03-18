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
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"


	sentry "github.com/getsentry/sentry-go"
	"github.com/gorilla/mux"
	"github.com/rs/cors"

	promhttp "github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/zeirash/recapo/arion/common/config"
	"github.com/zeirash/recapo/arion/common/database"
	"github.com/zeirash/recapo/arion/common/logger"
	"github.com/zeirash/recapo/arion/common/middleware"
	"github.com/zeirash/recapo/arion/handler"

	_ "github.com/zeirash/recapo/arion/docs" // swagger docs
)

func initSentry() {
	cfg := config.GetConfig()
	if cfg.SentryDSN == "" {
		return
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              cfg.SentryDSN,
		Environment:      cfg.Env,
		Release:          cfg.Version,
		TracesSampleRate: 0.0,
	})
	if err != nil {
		logger.WithError(err).Warn("failed to initialize Sentry")
	}
}

func NewRouter() *mux.Router {
	handler.Init()
	r := mux.NewRouter()

	// Global middleware
	r.Use(middleware.Recovery)
	r.Use(middleware.RequestLogger)
	r.Use(middleware.Metrics)

	// Metrics endpoint (unauthenticated)
	r.Handle("/metrics", promhttp.Handler())

	// Swagger UI (WrapHandler is http.HandlerFunc; doc is served from swag registry)
	r.PathPrefix("/swagger/").HandlerFunc(httpSwagger.WrapHandler)

	// Routes API No Auth
	r.HandleFunc("/health", handler.HealthHandler)
	r.HandleFunc("/plans", handler.GetPlansHandler).Methods("GET")
	r.HandleFunc("/public/shops/{share_token}/products", handler.GetShopProductsHandler).Methods("GET")
	r.HandleFunc("/public/shops/{share_token}/order", handler.CreateShopTempOrderHandler).Methods("POST")

	r.HandleFunc("/login", handler.LoginHandler).Methods("POST")
	r.HandleFunc("/send_otp", handler.SendOTPHandler).Methods("POST")
	r.HandleFunc("/register", handler.RegisterHandler).Methods("POST")
	r.HandleFunc("/refresh", handler.RefreshHandler).Methods("POST")
	r.HandleFunc("/forgot_password", handler.ForgotPasswordHandler).Methods("POST")
	r.HandleFunc("/reset_password", handler.ResetPasswordHandler).Methods("POST")
	r.Handle("/logout", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.LogoutHandler))).Methods("POST")

	// Subscription
	r.HandleFunc("/webhook/midtrans", handler.MidtransWebhookHandler).Methods("POST")
	r.Handle("/subscription", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.GetSubscriptionHandler))).Methods("GET")
	r.Handle("/subscription/checkout", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.CheckoutHandler))).Methods("POST")
	r.Handle("/subscription/cancel", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.CancelSubscriptionHandler))).Methods("POST")

	// User
	r.Handle("/user", middleware.ChainMiddleware(middleware.Authentication, middleware.SubscriptionCheck)(http.HandlerFunc(handler.UpdateUserHandler))).Methods("PATCH")
	r.Handle("/user", middleware.ChainMiddleware(middleware.Authentication, middleware.SubscriptionCheck)(http.HandlerFunc(handler.GetUserHandler))).Methods("GET")
	r.Handle("/users", middleware.ChainMiddleware(middleware.Authentication, middleware.SubscriptionCheck)(http.HandlerFunc(handler.GetUsersHandler))).Methods("GET")

	// Customer
	r.Handle("/customer", middleware.ChainMiddleware(middleware.Authentication, middleware.SubscriptionCheck)(http.HandlerFunc(handler.CreateCustomerHandler))).Methods("POST")
	r.Handle("/customers", middleware.ChainMiddleware(middleware.Authentication, middleware.SubscriptionCheck)(http.HandlerFunc(handler.GetCustomersHandler))).Methods("GET")
	r.Handle("/customers/{customer_id}", middleware.ChainMiddleware(middleware.Authentication, middleware.SubscriptionCheck)(http.HandlerFunc(handler.UpdateCustomerHandler))).Methods("PATCH")
	r.Handle("/customers/{customer_id}", middleware.ChainMiddleware(middleware.Authentication, middleware.SubscriptionCheck)(http.HandlerFunc(handler.DeleteCustomerHandler))).Methods("DELETE")
	r.Handle("/customers/{customer_id}", middleware.ChainMiddleware(middleware.Authentication, middleware.SubscriptionCheck)(http.HandlerFunc(handler.GetCustomerHandler))).Methods("GET")
	r.Handle("/customers/check_active_order", middleware.ChainMiddleware(middleware.Authentication, middleware.SubscriptionCheck)(http.HandlerFunc(handler.CustomerCheckActiveOrderHandler))).Methods("POST")

	// Shop
	r.Handle("/shop/share_token", middleware.ChainMiddleware(middleware.Authentication, middleware.SubscriptionCheck)(http.HandlerFunc(handler.GetShopShareTokenHandler))).Methods("GET")

	// For Product (register literal paths before /products/{product_id} so they match first)
	r.Handle("/product", middleware.ChainMiddleware(middleware.Authentication, middleware.SubscriptionCheck)(http.HandlerFunc(handler.CreateProductHandler))).Methods("POST")
	r.Handle("/products", middleware.ChainMiddleware(middleware.Authentication, middleware.SubscriptionCheck)(http.HandlerFunc(handler.GetProductsHandler))).Methods("GET")
	r.Handle("/products/purchase_list", middleware.ChainMiddleware(middleware.Authentication, middleware.SubscriptionCheck)(http.HandlerFunc(handler.PurchaseListProductHandler))).Methods("GET")
	r.Handle("/products/image", middleware.ChainMiddleware(middleware.Authentication, middleware.SubscriptionCheck)(http.HandlerFunc(handler.UploadProductImageHandler))).Methods("POST")
	r.Handle("/products/image", middleware.ChainMiddleware(middleware.Authentication, middleware.SubscriptionCheck)(http.HandlerFunc(handler.DeleteProductImageHandler))).Methods("DELETE")
	r.Handle("/products/{product_id}", middleware.ChainMiddleware(middleware.Authentication, middleware.SubscriptionCheck)(http.HandlerFunc(handler.UpdateProductHandler))).Methods("PATCH")
	r.Handle("/products/{product_id}", middleware.ChainMiddleware(middleware.Authentication, middleware.SubscriptionCheck)(http.HandlerFunc(handler.DeleteProductHandler))).Methods("DELETE")
	r.Handle("/products/{product_id}", middleware.ChainMiddleware(middleware.Authentication, middleware.SubscriptionCheck)(http.HandlerFunc(handler.GetProductHandler))).Methods("GET")

	// Order
	r.Handle("/order", middleware.ChainMiddleware(middleware.Authentication, middleware.SubscriptionCheck)(http.HandlerFunc(handler.CreateOrderHandler))).Methods("POST")
	r.Handle("/orders", middleware.ChainMiddleware(middleware.Authentication, middleware.SubscriptionCheck)(http.HandlerFunc(handler.GetOrdersHandler))).Methods("GET")
	r.Handle("/orders/{order_id}", middleware.ChainMiddleware(middleware.Authentication, middleware.SubscriptionCheck)(http.HandlerFunc(handler.UpdateOrderHandler))).Methods("PATCH")
	r.Handle("/orders/{order_id}", middleware.ChainMiddleware(middleware.Authentication, middleware.SubscriptionCheck)(http.HandlerFunc(handler.DeleteOrderHandler))).Methods("DELETE")
	r.Handle("/orders/{order_id}", middleware.ChainMiddleware(middleware.Authentication, middleware.SubscriptionCheck)(http.HandlerFunc(handler.GetOrderHandler))).Methods("GET")
	r.Handle("/orders/{order_id}/export", middleware.ChainMiddleware(middleware.Authentication, middleware.SubscriptionCheck)(http.HandlerFunc(handler.ExportOrderHandler))).Methods("POST")
	r.Handle("/orders/{order_id}/item", middleware.ChainMiddleware(middleware.Authentication, middleware.SubscriptionCheck)(http.HandlerFunc(handler.CreateOrderItemHandler))).Methods("POST")
	r.Handle("/orders/{order_id}/items", middleware.ChainMiddleware(middleware.Authentication, middleware.SubscriptionCheck)(http.HandlerFunc(handler.GetOrderItemsHandler))).Methods("GET")
	r.Handle("/orders/{order_id}/items/{item_id}", middleware.ChainMiddleware(middleware.Authentication, middleware.SubscriptionCheck)(http.HandlerFunc(handler.UpdateOrderItemHandler))).Methods("PATCH")
	r.Handle("/orders/{order_id}/items/{item_id}", middleware.ChainMiddleware(middleware.Authentication, middleware.SubscriptionCheck)(http.HandlerFunc(handler.DeleteOrderItemHandler))).Methods("DELETE")
	r.Handle("/orders/{order_id}/items/{item_id}", middleware.ChainMiddleware(middleware.Authentication, middleware.SubscriptionCheck)(http.HandlerFunc(handler.GetOrderItemHandler))).Methods("GET")
	r.Handle("/orders/{order_id}/payment", middleware.ChainMiddleware(middleware.Authentication, middleware.SubscriptionCheck)(http.HandlerFunc(handler.CreateOrderPaymentHandler))).Methods("POST")
	r.Handle("/orders/{order_id}/payments", middleware.ChainMiddleware(middleware.Authentication, middleware.SubscriptionCheck)(http.HandlerFunc(handler.GetOrderPaymentsHandler))).Methods("GET")
	r.Handle("/orders/{order_id}/payments/{payment_id}", middleware.ChainMiddleware(middleware.Authentication, middleware.SubscriptionCheck)(http.HandlerFunc(handler.UpdateOrderPaymentAmountHandler))).Methods("PATCH")
	r.Handle("/orders/{order_id}/payments/{payment_id}", middleware.ChainMiddleware(middleware.Authentication, middleware.SubscriptionCheck)(http.HandlerFunc(handler.DeleteOrderPaymentHandler))).Methods("DELETE")

	// Temp Order
	r.Handle("/temp_orders", middleware.ChainMiddleware(middleware.Authentication, middleware.SubscriptionCheck)(http.HandlerFunc(handler.GetTempOrdersHandler))).Methods("GET")
	r.Handle("/temp_orders/merge", middleware.ChainMiddleware(middleware.Authentication, middleware.SubscriptionCheck)(http.HandlerFunc(handler.MergeTempOrderHandler))).Methods("POST")
	r.Handle("/temp_orders/{temp_order_id}", middleware.ChainMiddleware(middleware.Authentication, middleware.SubscriptionCheck)(http.HandlerFunc(handler.GetTempOrderHandler))).Methods("GET")
	r.Handle("/temp_orders/{temp_order_id}/reject", middleware.ChainMiddleware(middleware.Authentication, middleware.SubscriptionCheck)(http.HandlerFunc(handler.RejectTempOrderHandler))).Methods("PATCH")

	// Feedback
	r.Handle("/feedback", middleware.ChainMiddleware(middleware.Authentication)(http.HandlerFunc(handler.CreateFeedbackHandler))).Methods("POST")

	// System
	r.Handle("/system/user/{user_id}", middleware.ChainMiddleware(middleware.Authentication, middleware.CheckSystemMode)(http.HandlerFunc(handler.GetUserHandler))).Methods("GET")

	// Static file serving for uploaded product images
	uploadDir := config.GetConfig().UploadDir
	r.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadDir))))

	return r
}

func main() {
	// init logger
	logger.Init()

	// init config
	config.InitConfig()

	// init sentry
	initSentry()
	defer sentry.Flush(2 * time.Second)

	// init database
	database.InitDB()
	defer database.CloseDB()
	database.RegisterDBMetrics(database.GetDB())

	// ensure upload directory exists
	os.MkdirAll(config.GetConfig().UploadDir+"/products", 0755)

	// init router (also initializes handler services via handler.Init())
	r := NewRouter()

	// start cron jobs
	startCron()

	cfg := config.GetConfig()
	c := cors.New(cors.Options{
		AllowedOrigins:   strings.Split(cfg.CORSOrigins, ","),
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		Debug:            cfg.Env != "production",
	})
	h := c.Handler(r)

	port := fmt.Sprintf(":%d", config.GetConfig().ServicePort)
	logger.Info("Server starting on port ", port)
	logger.Fatal(http.ListenAndServe(port, h))
}

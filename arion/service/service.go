package service

import (
	"github.com/zeirash/recapo/arion/common/config"
	"github.com/zeirash/recapo/arion/common/database"
	"github.com/zeirash/recapo/arion/store"
)

var (
	cfg config.Config

	userStore      store.UserStore
	tokenStore     store.TokenStore
	shopStore      store.ShopStore
	customerStore  store.CustomerStore
	productStore   store.ProductStore
	orderStore     store.OrderStore
	orderItemStore store.OrderItemStore

	// dbGetter is a function that returns a database connection.
	// It can be overridden in tests to return a mock.
	dbGetter func() database.DB = func() database.DB { return database.GetDBWrapper() }
)

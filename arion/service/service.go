package service

import (
	"github.com/zeirash/recapo/arion/common/config"
	"github.com/zeirash/recapo/arion/store"
)

var (
	cfg config.Config

	userStore store.UserStore
	tokenStore store.TokenStore
)

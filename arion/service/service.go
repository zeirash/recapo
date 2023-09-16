package service

import (
	"github.com/zeirash/recapo/arion/common/config"
	"github.com/zeirash/recapo/arion/store"
)

var (
	cfg config.Config

	user store.UserStore
	token store.TokenStore
)

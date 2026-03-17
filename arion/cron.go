package main

import (
	"context"
	"time"

	"github.com/zeirash/recapo/arion/common/logger"
	"github.com/zeirash/recapo/arion/service"
)

func startCron() {
	go runDailyCron()
}

func runDailyCron() {
	svc := service.NewSubscriptionService()
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	// run once on startup
	runExpireSubscriptions(svc)

	for range ticker.C {
		runExpireSubscriptions(svc)
	}
}

func runExpireSubscriptions(svc service.SubscriptionService) {
	if err := svc.ExpireSubscriptions(context.Background()); err != nil {
		logger.WithError(err).Error("expire_subscriptions_cron_error")
	}
}

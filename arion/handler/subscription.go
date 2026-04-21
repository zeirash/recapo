package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/zeirash/recapo/arion/common"
	"github.com/zeirash/recapo/arion/common/apierr"
	"github.com/zeirash/recapo/arion/common/logger"
	"github.com/zeirash/recapo/arion/service"
)

// GetPlansHandler godoc
//
//	@Summary		List active plans
//	@Description	Returns all active subscription plans. No authentication required.
//	@Tags			subscription
//	@Produce		json
//	@Success		200	{array}		response.PlanData
//	@Failure		500	{object}	ErrorApiResponse
//	@Router			/plans [get]
func GetPlansHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	plans, err := subscriptionService.GetActivePlans(ctx)
	if err != nil {
		logger.WithError(err).Error("get_plans_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "get_plans")
		return
	}
	WriteJson(w, http.StatusOK, plans)
}

// GetSubscriptionHandler godoc
//
//	@Summary		Get shop subscription
//	@Description	Returns the current subscription and plan for the authenticated shop.
//	@Tags			subscription
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	response.SubscriptionData
//	@Failure		404	{object}	ErrorApiResponse
//	@Failure		500	{object}	ErrorApiResponse
//	@Router			/subscription [get]
func GetSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shopID := ctx.Value(common.ShopIDKey).(int)

	sub, err := subscriptionService.GetSubscriptionByShopID(ctx, shopID)
	if err != nil {
		if err.Error() == apierr.ErrSubscriptionNotFound {
			WriteErrorJson(w, r, http.StatusNotFound, err, "not_found")
			return
		}
		logger.WithError(err).Error("get_subscription_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "get_subscription")
		return
	}

	WriteJson(w, http.StatusOK, sub)
}

type CheckoutRequest struct {
	PlanID int `json:"plan_id"`
}

// CheckoutHandler godoc
//
//	@Summary		Initiate checkout
//	@Description	Creates a Midtrans SNAP payment session for the given plan. Returns a redirect URL.
//	@Tags			subscription
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		CheckoutRequest	true	"plan_id"
//	@Success		200		{object}	response.CheckoutData
//	@Failure		400		{object}	ErrorApiResponse
//	@Failure		404		{object}	ErrorApiResponse
//	@Failure		500		{object}	ErrorApiResponse
//	@Router			/subscription/checkout [post]
func CheckoutHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shopID := ctx.Value(common.ShopIDKey).(int)

	inp := CheckoutRequest{}
	if err := ParseJson(r.Body, &inp); err != nil {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "parse_json")
		return
	}

	if inp.PlanID <= 0 {
		WriteErrorJson(w, r, http.StatusBadRequest, errors.New(apierr.ErrPlanIDRequired), "validation")
		return
	}

	data, err := subscriptionService.Checkout(ctx, shopID, inp.PlanID)
	if err != nil {
		if err.Error() == apierr.ErrPlanNotFound || err.Error() == apierr.ErrSubscriptionNotFound {
			WriteErrorJson(w, r, http.StatusNotFound, err, "not_found")
			return
		}
		logger.WithError(err).Error("checkout_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "checkout")
		return
	}

	WriteJson(w, http.StatusOK, data)
}

// CancelSubscriptionHandler godoc
//
//	@Summary		Cancel subscription
//	@Description	Cancels the active subscription for the authenticated shop.
//	@Tags			subscription
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	object{}
//	@Failure		400	{object}	ErrorApiResponse
//	@Failure		404	{object}	ErrorApiResponse
//	@Failure		500	{object}	ErrorApiResponse
//	@Router			/subscription/cancel [post]
func CancelSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shopID := ctx.Value(common.ShopIDKey).(int)

	if err := subscriptionService.CancelSubscription(ctx, shopID); err != nil {
		if err.Error() == apierr.ErrSubscriptionNotFound {
			WriteErrorJson(w, r, http.StatusNotFound, err, "not_found")
			return
		}
		if err.Error() == apierr.ErrSubscriptionNotActive {
			WriteErrorJson(w, r, http.StatusBadRequest, err, "not_active")
			return
		}
		logger.WithError(err).Error("cancel_subscription_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "cancel_subscription")
		return
	}

	WriteJson(w, http.StatusOK, struct{}{})
}

// MidtransWebhookHandler godoc
//
//	@Summary		Midtrans payment webhook
//	@Description	Receives Midtrans payment notifications and activates subscriptions on settlement.
//	@Tags			subscription
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	object{}
//	@Failure		400	{object}	ErrorApiResponse
//	@Failure		500	{object}	ErrorApiResponse
//	@Router			/webhook/midtrans [post]
func MidtransWebhookHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	payload := service.MidtransWebhookPayload{}
	if err := ParseJson(r.Body, &payload); err != nil {
		WriteErrorJson(w, r, http.StatusBadRequest, err, "parse_json")
		return
	}

	if strings.HasPrefix(payload.OrderID, "payment_notif_test_") {
		WriteJson(w, http.StatusOK, struct{}{})
		return
	}

	if err := subscriptionService.HandleMidtransWebhook(ctx, payload); err != nil {
		if err.Error() == apierr.ErrInvalidSignature {
			WriteErrorJson(w, r, http.StatusBadRequest, err, "invalid_signature")
			return
		}
		if err.Error() == apierr.ErrPaymentNotFound {
			WriteErrorJson(w, r, http.StatusNotFound, err, "not_found")
			return
		}
		logger.WithError(err).Error("midtrans_webhook_error")
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "webhook_error")
		return
	}

	WriteJson(w, http.StatusOK, struct{}{})
}

package middleware

import (
	"errors"
	"net/http"

	"github.com/zeirash/recapo/arion/common"
	"github.com/zeirash/recapo/arion/common/apierr"
	"github.com/zeirash/recapo/arion/handler"
	"github.com/zeirash/recapo/arion/service"
)

// SubscriptionCheck middleware verifies the shop has an active subscription.
// Must be chained after Authentication (requires ShopIDKey in context).
// Returns HTTP 402 with code "subscription_required" if inactive.
func SubscriptionCheck(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		shopID, ok := ctx.Value(common.ShopIDKey).(int)
		if !ok || shopID == 0 {
			handler.WriteErrorJson(w, r, http.StatusUnauthorized, errors.New(apierr.ErrMissingShopContext), "unauthorized")
			return
		}

		svc := handler.GetSubscriptionService()
		if svc == nil {
			// Fallback: use a new service instance if handler not initialized yet
			svc = service.NewSubscriptionService()
		}

		active, err := svc.IsSubscriptionActive(shopID)
		if err != nil {
			handler.WriteErrorJson(w, r, http.StatusInternalServerError, err, "subscription_check")
			return
		}

		if !active {
			handler.WriteErrorJson(w, r, http.StatusPaymentRequired, errors.New(apierr.ErrSubscriptionRequired), "subscription_required")
			return
		}

		next.ServeHTTP(w, r)
	})
}

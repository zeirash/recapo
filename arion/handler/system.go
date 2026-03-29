package handler

import (
	"net/http"
)

// GetSystemStatsHandler godoc
//
//	@Summary		System stats
//	@Description	Returns platform-wide stats: total shops, users, subscriptions by status, MRR.
//	@Tags			system
//	@Produce		json
//	@Success		200	{object}	ApiResponse
//	@Failure		401	{object}	ErrorApiResponse
//	@Router			/system/stats [get]
//	@Security		BearerAuth
func GetSystemStatsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	stats, err := systemService.GetSystemStats(ctx)
	if err != nil {
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "get_system_stats")
		return
	}
	WriteJson(w, http.StatusOK, stats)
}

// GetSystemShopsHandler godoc
//
//	@Summary		System shops list
//	@Description	Returns all shops with owner info and subscription details.
//	@Tags			system
//	@Produce		json
//	@Success		200	{object}	ApiResponse
//	@Failure		401	{object}	ErrorApiResponse
//	@Router			/system/shops [get]
//	@Security		BearerAuth
func GetSystemShopsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shops, err := systemService.GetSystemShops(ctx)
	if err != nil {
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "get_system_shops")
		return
	}
	WriteJson(w, http.StatusOK, shops)
}

// GetSystemPaymentsHandler godoc
//
//	@Summary		System payment history
//	@Description	Returns all subscription payments across all shops.
//	@Tags			system
//	@Produce		json
//	@Success		200	{object}	ApiResponse
//	@Failure		401	{object}	ErrorApiResponse
//	@Router			/system/payments [get]
//	@Security		BearerAuth
func GetSystemPaymentsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	payments, err := systemService.GetSystemPayments(ctx)
	if err != nil {
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "get_system_payments")
		return
	}
	WriteJson(w, http.StatusOK, payments)
}

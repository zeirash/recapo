package handler

import (
	"net/http"

	"github.com/zeirash/recapo/arion/model"
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
//	@Param			date_from	query		string	false	"Filter from date (YYYY-MM-DD)"
//	@Param			date_to		query		string	false	"Filter to date (YYYY-MM-DD)"
//	@Param			status		query		string	false	"Filter by status (e.g. pending,paid,failed)"
//	@Param			sort		  query		string	false	"Sort by column and order (e.g. created_at,desc)"
//	@Success		200	{object}	ApiResponse
//	@Failure		401	{object}	ErrorApiResponse
//	@Router			/system/payments [get]
//	@Security		BearerAuth
func GetSystemPaymentsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	opts := model.SystemPaymentFilterOptions{}
	if df := r.URL.Query().Get("date_from"); df != "" {
		if t, err := parseDate(df); err == nil {
			opts.DateFrom = &t
		}
	}
	if dt := r.URL.Query().Get("date_to"); dt != "" {
		if t, err := parseDate(dt); err == nil {
			opts.DateTo = &t
		}
	}
	if sort := r.URL.Query().Get("sort"); sort != "" {
		opts.Sort = &sort
	}
	if status := r.URL.Query().Get("status"); status != "" {
		opts.Status = &status
	}
	payments, err := systemService.GetSystemPayments(ctx, opts)
	if err != nil {
		WriteErrorJson(w, r, http.StatusInternalServerError, err, "get_system_payments")
		return
	}
	WriteJson(w, http.StatusOK, payments)
}

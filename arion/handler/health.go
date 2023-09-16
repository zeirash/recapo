package handler

import (
	"net/http"

	"github.com/zeirash/recapo/arion/common/response"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	body := response.HealthCheck{
		Status: "OK",
	}

	WriteJson(w, http.StatusOK, body)
}

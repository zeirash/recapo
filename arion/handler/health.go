package handler

import (
	"net/http"

	"github.com/zeirash/recapo/arion/common/response"
)

type Body struct {
	Status string `json:"status"`
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	body := Body{
		Status: "OK",
	}

	response.SuccessResponse(w, http.StatusOK, body)
}

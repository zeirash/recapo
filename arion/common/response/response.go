package response

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ErrorBodyResponse struct {
	Code  int    `json:"code"`
	Cause string `json:"cause"`
}

func SuccessResponse(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(status)

	jsonResp, err := json.Marshal(body)
	if err != nil {
		fmt.Println("error marshall body: ", err.Error())
		return
	}
	w.Write(jsonResp)
}

func ErrorResponse(w http.ResponseWriter, status int, code int, err error) {
	w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(status)

	body := ErrorBodyResponse{
		Code:  code,
		Cause: err.Error(),
	}

	jsonResp, err := json.Marshal(body)
	if err != nil {
		fmt.Println("error marshall body: ", err.Error())
		return
	}
	w.Write(jsonResp)
}

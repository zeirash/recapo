package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
	"github.com/zeirash/recapo/arion/service"
)

type ApiResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Code    string      `json:"code"`
	Message string      `json:"message"`
}

var (
	userService service.UserService
)

func Init() {
	if userService == nil {
		userService = service.NewUserService()
	}
}

func WriteJson(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(status)

	res := ApiResponse{
		Success: true,
		Data:    body,
	}

	jsonResp, err := json.Marshal(res)
	if err != nil {
		fmt.Println("error marshall body: ", err.Error())
		return
	}
	w.Write(jsonResp)
}

func WriteErrorJson(w http.ResponseWriter, status int, err error, code string) {
	w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(status)

	res := ApiResponse{
		Success: false,
		Code: code,
		Message: err.Error(),
	}

	jsonResp, err := json.Marshal(res)
	if err != nil {
		fmt.Println("error marshall body: ", err.Error())
		return
	}
	w.Write(jsonResp)
}

func ParseJson(input io.ReadCloser, result interface{}) error {
	err := json.NewDecoder(input).Decode(result)
	return errors.Wrap(err, "Failed parsing json")
}

package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
	"github.com/zeirash/recapo/arion/service"
)

type ErrorBodyResponse struct {
	Code  string `json:"code"`
	Error string `json:"error"`
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

	jsonResp, err := json.Marshal(body)
	if err != nil {
		fmt.Println("error marshall body: ", err.Error())
		return
	}
	w.Write(jsonResp)
}

func WriteErrorJson(w http.ResponseWriter, status int, err error, code string) {
	w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(status)

	body := ErrorBodyResponse{
		Code: code,
		Error: err.Error(),
	}

	jsonResp, err := json.Marshal(body)
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

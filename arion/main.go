package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/zeirash/recapo/arion/handler"
	// "sort"
	// "github.com/joho/godotenv"
	// "github.com/kelseyhightower/envconfig"
	// "golang.org/x/net/http2"
	// "golang.org/x/net/http2/h2c"
	// "golang-program-structure/common/logging"
	// _ "golang-program-structure/docs"
	// "golang-program-structure/service"
)

func NewRouter() *mux.Router {
	r := mux.NewRouter()

	return r
}

func main() {
	r := NewRouter()

	// Routes API
	r.HandleFunc("/health", handler.HealthHandler)

	fmt.Println("Server run on port :3000")
	log.Fatal(http.ListenAndServe(":3000", r))

	// // Load the environment vars from a .env file if present
	// err := godotenv.Load()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// // Output the environment, sorted by var, for debug purposes
	// var envVars map[string]string
	// envVars, _ = godotenv.Read()
	// keys := make([]string, 0, len(envVars))
	// for key := range envVars {
	// 	keys = append(keys, key)
	// }
	// sort.Strings(keys)

	// // Load the config struct with values from the environment without any prefix (i.e. "")
	// var config service.Config
	// err = envconfig.Process("", &config)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// // Set up the logger and attach to the config struct
	// logging.SetupLogger(config.Env)
	// logger := logging.GetLogger()
	// config.Logger = logger

	// // Output the config for debugging
	// logger.Infof("%+v\n", config)

	// // Set the router/handler environment level and initialize
	// mode := gin.ReleaseMode
	// if config.Env == "local" {
	// 	mode = gin.DebugMode
	// }
	// gin.SetMode(mode)

	// // Service HTTP/2 requests using unencrypted method, H2C
	// address := fmt.Sprintf(":%v", config.ServicePort)
	// h := NewHandler(&config)

	// server := &http.Server{
	// 	Addr:    address,
	// 	Handler: h2c.NewHandler(h, &http2.Server{}),
	// }

	// logger.Infof("Listening on %s", address)
	// logger.Fatal(server.ListenAndServe())
}

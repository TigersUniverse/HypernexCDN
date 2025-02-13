package main

import (
	"HypernexCDN/cdn"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"os"
)

func main() {
	if !loadConfig() {
		fmt.Println("Config generated! Please fill out all info and return when done.")
		os.Exit(1)
		return
	}
	cdn.CreateSession(config.AWS_key, config.AWS_secret, config.AWS_endpoint, config.AWS_region, config.AWS_bucket)
	router := mux.NewRouter()
	cdn.CreateRoutes(router)
	err := http.ListenAndServe(":3333", router)

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}

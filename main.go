package main

import (
	"log"

	"github.com/coapcloud/coap-hooks-router/coap"
	hooksAPI "github.com/coapcloud/coap-hooks-router/hooks"
)

const (
	hooksAPIPort = 8081
	coapPort     = 5683
)

func main() {
	hooksRepo, err := hooksAPI.NewHooksRepository("dbdata/hooks.db")
	if err != nil {
		log.Fatalf("Error trying to create HooksRepository: %v\n", err)
	}
	defer hooksRepo.Cleanup()

	go hooksAPI.ListenAndServe(hooksRepo, hooksAPIPort)

	coap.ListenAndServe(hooksRepo, coapPort)
}

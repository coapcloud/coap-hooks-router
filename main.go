package main

import (
	"log"

	"github.com/coapcloud/coap-hooks-router/config"
	"github.com/coapcloud/coap-hooks-router/hooks"
)

func main() {
	hooksRepo, err := hooks.NewHooksRepository(config.DBFilename)
	if err != nil {
		log.Fatalf("Error trying to create HooksRepository: %v\n", err)
	}
	defer hooksRepo.Cleanup()

	hooks.ListenAndServe(hooksRepo, 8081)
}

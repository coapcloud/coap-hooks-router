package main

import (
	"fmt"
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

	go func() {
		for e := range hooksRepo.Events() {
			switch e.EventType {
			case hooks.EventTypePut:
				fmt.Printf("PUT EVENT for hook: %s owned by %s\n", e.Hook.Name, e.Hook.Owner)
			case hooks.EventTypeDelete:
				fmt.Printf("DELETE EVENT for hook: %s owned by %s\n", e.Hook.Name, e.Hook.Owner)
			case hooks.EventTypeDeleteForOwner:
				fmt.Println("DELETE FOR OWNER EVENT for owner")
			}
		}
	}()

	hooks.ListenAndServe(hooksRepo, 8081)
}

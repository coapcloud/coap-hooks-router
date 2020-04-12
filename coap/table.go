package coap

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/coapcloud/coap-hooks-router/hooks"
	"github.com/coapcloud/go-coap/codes"
	"github.com/derekparker/trie"
)

type routeTable struct {
	*trie.Trie
	*sync.RWMutex
}

func newRouteTable() routeTable {
	return routeTable{
		trie.New(),
		&sync.RWMutex{},
	}
}

func registerRoutes(r *routeTable, hooksRepo hooks.Repository) {
	allHooks, err := hooksRepo.ListAllHooks()
	if err != nil {
		log.Fatal("couldn't list all hooks")
	}

	log.Println("Adding persisted routes to the route table...")
	log.Println()

	for _, v := range allHooks {
		err = r.registerRouteForHook(*v)
		if err != nil {
			log.Fatal("error registering route: *v")
		}
	}
}

func (r *routeTable) HotAddRoute(h hooks.Hook) error {
	return r.registerRouteForHook(h)
}

func (r *routeTable) HotUpdateRouteDest(h hooks.Hook) error {
	err := r.deregisterRouteForHook(h)
	if err != nil {
		return fmt.Errorf("error encountered while deregistering route: %w", err)
	}

	return r.registerRouteForHook(h)
}

func (r *routeTable) HotRemoveRoute(h hooks.Hook) error {
	return r.deregisterRouteForHook(h)
}

func routeKey(verb codes.Code, owner, hookName string) string {
	key := fmt.Sprintf("%d-%s/%s", verb, owner, hookName)

	fmt.Printf("route key generated: %s\n", key)
	return key
}

func (r *routeTable) registerRouteForHook(h hooks.Hook) error {
	r.Lock()
	defer r.Unlock()

	verb := codes.POST // this part can later be extended to allow the registration of different verbs

	key := routeKey(verb, h.Owner, h.Name)

	if r.Add(key, h) == nil {
		return errors.New("did not register route, corrupt routing structure")
	}

	log.Printf("added route: %s -> %s\n", key, h.Destination)

	return nil
}

func (r *routeTable) deregisterRouteForHook(h hooks.Hook) error {
	r.Lock()
	defer r.Unlock()

	verb := codes.POST // this part can later be extended to allow the registration of different verbs

	key := routeKey(verb, h.Owner, h.Name)

	_, ok := r.Find(key)
	if !ok {
		return errors.New("could not find route for deletion")
	}

	r.Remove(key)

	log.Printf("deleted route: %s -> %s\n", key, h.Destination)

	return nil
}

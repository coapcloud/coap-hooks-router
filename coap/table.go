package coap

import (
	"errors"
	"log"
	"sync"

	"github.com/coapcloud/coap-hooks-router/hooks"
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

	for _, v := range allHooks {
		r.registerRoute(*v)
	}
}

func (r *routeTable) registerRoute(h hooks.Hook) {
	key := routeKey(h.Owner, h.Name)

	// log.Printf("registering route: %v %v -> openfaas func: %s\n", verb.String(), path, openfaasFuncID)

	node := r.Add(key, h)
	if node != nil {
		// log.Printf("registered route: %s /%s to func %q\n", verb.String(), path, openfaasFuncID)
	}
}

func (r *routeTable) HotRegisterRoute(h hooks.Hook) {
	r.Lock()
	defer r.Unlock()

	key := routeKey(h.Owner, h.Name)

	// log.Printf("registering route: %v %v -> openfaas func: %s\n", verb.String(), path, openfaasFuncID)

	node := r.Add(key, h)
	if node != nil {
		// log.Printf("registered route: %s /%s to func %q\n", verb.String(), path, openfaasFuncID)
	}
}

func (r *routeTable) HotModifyRoute(h hooks.Hook) error {
	r.Lock()
	defer r.Unlock()

	key := routeKey(h.Owner, h.Name)

	_, ok := r.Find(key)
	if !ok {
		return errors.New("could not find route")
	}

	// log.Printf("deregistering route: %v %v -> openfaas func: %s\n", verb.String(), path, openfaasFuncID)

	r.Remove(key)

	// log.Printf("registering route: %v %v -> openfaas func: %s\n", verb.String(), path, openfaasFuncID)

	node := r.Add(key, h)
	if node != nil {
		// log.Printf("registered route: %s /%s to func %q\n", verb.String(), path, openfaasFuncID)
	}

	return nil
}

func (r *routeTable) HotDeRegisterRoute(h hooks.Hook) error {
	r.Lock()
	defer r.Unlock()

	key := routeKey(h.Owner, h.Name)

	_, ok := r.Find(key)
	if !ok {
		return errors.New("could not find route")
	}

	// log.Printf("deregistering route: %v %v -> openfaas func: %s\n", verb.String(), path)

	r.Remove(key)

	return nil
}

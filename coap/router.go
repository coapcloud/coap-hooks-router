package coap

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/coapcloud/coap-hooks-router/hooks"
	"github.com/coapcloud/go-coap"
	"github.com/coapcloud/go-coap/codes"
)

func ListenAndServe(hooksRepo hooks.Repository, port int) {
	r := newRouteTable()
	registerRoutes(&r, hooksRepo)

	mux := coap.NewServeMux()
	mux.Handle("*", r)

	// start listening for hook API events to register and deregister routes
	go func() {
		for e := range hooksRepo.Events() {
			changedHooks := e.Hooks

			switch e.EventType {
			case hooks.EventTypeCreate:
				r.HotAddRoute(*changedHooks[0])
			case hooks.EventTypeDelete:
				r.HotRemoveRoute(*changedHooks[0])
			case hooks.EventTypeDeleteForOwner:
				for _, v := range changedHooks {
					err := r.HotRemoveRoute(*v)
					if err != nil {
						log.Printf("couldn't remove hook: %v with error: %v", *v, err)
					}
				}
			}
		}
	}()

	fmt.Printf("serving CoAP requests on %d\n", port)
	log.Fatal(coap.ListenAndServe("udp", fmt.Sprintf(":%d", port), mux))
}

// ServeCOAP - implementation of coap.Handler
func (r routeTable) ServeCOAP(w coap.ResponseWriter, req *coap.Request) {
	var (
		respBdy []byte
		err     error
	)

	log.Printf("Got message: %#v path=%q: from %v\n", req.Msg.PathString(), req.Msg, req.Client.RemoteAddr())

	log.Printf("req: %+v\n", req)

	dest, ok := r.match(req.Msg.Code(), req.Msg.PathString())
	if ok {
		buf := new(bytes.Buffer)
		buf.Write(req.Msg.Payload())

		respBdy, err = dest.Fire(buf)
		if err != nil {
			log.Printf("Error while trying to invoke webhook %v\n", err)
			w.SetCode(codes.InternalServerError)
			respBdy = []byte("could not run callback for request")
		}
	} else {
		log.Println("could not match route")
		w.SetCode(codes.NotFound)
		respBdy = []byte("not found")
	}

	ctx, cancel := context.WithTimeout(req.Ctx, 3*time.Second)
	defer cancel()

	log.Printf("Writing response to %v\n\n", req.Client.RemoteAddr())
	w.SetContentFormat(coap.TextPlain)
	if _, err := w.WriteWithContext(ctx, respBdy); err != nil {
		log.Printf("Cannot send response: %v", err)
	}
}

func (r *routeTable) match(verb codes.Code, path string) (hooks.Hook, bool) {
	r.RLock()
	defer r.RUnlock()

	path = strings.TrimLeft(path, "/")

	toks := strings.Split(path, "/")

	if len(toks) < 2 {
		log.Printf("couldn't find route from path: %v\n", path)
		return hooks.Hook{}, false
	}

	key := routeKey(verb, toks[0], toks[1])

	node, ok := r.Find(key)
	if !ok {
		log.Printf("couldn't find route with key: %s\n", key)
		return hooks.Hook{}, false
	}

	v, ok := node.Meta().(hooks.Hook)
	if !ok {
		log.Printf("couldn't find route with key: %s (type assertion failed)\n", key)
		return hooks.Hook{}, false
	}

	return v, true
}

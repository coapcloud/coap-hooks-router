package coap

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/coapcloud/coap-hooks-router/hooks"
	"github.com/coapcloud/go-coap"
)

func ListenAndServe(hooksRepo hooks.Repository, port int) {
	r := newRouteTable()
	registerRoutes(&r, hooksRepo)

	mux := coap.NewServeMux()
	mux.Handle("*", r)

	fmt.Printf("serving CoAP requests on %d\n", port)
	log.Fatal(coap.ListenAndServe("udp", fmt.Sprintf(":%d", port), mux))
}

// ServeCOAP - implementation of coap.Handler
func (r routeTable) ServeCOAP(w coap.ResponseWriter, req *coap.Request) {
	var (
		respBdy string
		// err     error
	)

	log.Printf("Got message: %#v path=%q: from %v\n", req.Msg.PathString(), req.Msg, req.Client.RemoteAddr())

	log.Printf("req: %+v\n", req)

	// funcID, ok := r.match(req.Msg.Code(), req.Msg.PathString())
	// if !ok {
	// 	log.Println("could not match route")
	// 	w.SetCode(coap.NotFound)
	// 	respBdy = fmt.Sprintf("not found")
	// }

	// fire hook function for route + verb
	// respBdy, err = openfaasCall(funcID, req.Msg.Payload())
	// if err != nil {
	// 	log.Printf("Error while trying to invoke openfaas function %v\n", err)
	// 	w.SetCode(coap.InternalServerError)
	// 	respBdy = fmt.Sprint("could not run callback for request")
	// }

	ctx, cancel := context.WithTimeout(req.Ctx, 3*time.Second)
	defer cancel()

	log.Printf("Writing response to %v\n\n", req.Client.RemoteAddr())
	w.SetContentFormat(coap.TextPlain)
	if _, err := w.WriteWithContext(ctx, []byte(respBdy)); err != nil {
		log.Printf("Cannot send response: %v", err)
	}
}

func (r *routeTable) match(owner, hookName string) (string, bool) {
	r.RLock()
	defer r.RUnlock()

	key := routeKey(owner, hookName)

	fmt.Println(key)

	node, ok := r.Find(key)
	if !ok {
		log.Printf("couldn't find openfaas function id for route: %s\n", key)
		return "", false
	}

	meta := node.Meta()

	v, ok := meta.(string)
	if !ok {
		log.Printf("couldn't find string-ey openfaas function id for route: %s\n", key)
		return "", false
	}

	return v, true
}

func routeKey(owner, hookName string) string {
	return fmt.Sprintf("%d.%s", owner, hookName)
}

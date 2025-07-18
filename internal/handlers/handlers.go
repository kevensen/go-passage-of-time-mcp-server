package handlers

import (
	"log/slog"
	"net/http"
)

var (
	Routers map[string]Router
	mux     *http.ServeMux
)

type Router interface {
	Ready() bool
}

type Handler func(w http.ResponseWriter, r *http.Request)

func init() {
	mux = http.NewServeMux()
}

func Add(method, route string, router Router, handler Handler) {
	if Routers == nil {
		Routers = make(map[string]Router, 1)
	}

	if _, ok := Routers[route]; !ok {
		Routers[route] = router
	}
	mux.HandleFunc(method+" "+route, handler)
	slog.Info("Adding route", "method", method, "route", route)

}

func AddAll(route string, router Router, handler Handler) {
	if Routers == nil {
		Routers = make(map[string]Router, 1)
	}

	if _, ok := Routers[route]; !ok {
		Routers[route] = router
	}

	mux.HandleFunc(route, handler)
	slog.Info("Adding routes", "route", route)

}

func RouterByName(name string) Router {
	if r, ok := Routers[name]; ok {
		return r
	}
	slog.Error("Router not found", "name", name)
	return nil
}

func Start(addr string) error {
	s := &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	return s.ListenAndServe()
}

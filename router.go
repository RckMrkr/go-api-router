package router

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// Route type is used to specify a specific endpoint.
type Route struct {
	Name    string
	Headers []string
	Host    string
	Methods []string
	Path    string
	Queries []string
	Schemes []string
	Handler http.HandlerFunc
	Before  Before
	After   After
}

// Before is a faux class for functions that wrap the handlers being run before the actual handler
type Before []func(http.HandlerFunc) http.HandlerFunc

// After is a faux class for functions that wrap the handlers being run before the actual handler
type After []http.HandlerFunc

// Routes is used to bundle Route's
type Routes []Route

func createRoute(router *mux.Router, route Route) {
	r := router.
		Path(route.Path).
		Name(route.Name)

	if route.Headers != nil {
		r.Headers(route.Headers...)
	}

	if route.Queries != nil {
		r.Queries(route.Queries...)
	}

	if route.Schemes != nil {
		r.Schemes(route.Schemes...)
	}

	if route.Methods != nil {
		r.Methods(route.Methods...)
	}

	if route.Host != "" {
		r.Host(route.Host)
	}

	handler := route.Handler
	beforeLength := len(route.Before) - 1
	for i := range route.Before {
		handler = route.Before[beforeLength-i](handler)
	}

	r.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
		for _, after := range route.After {
			after(w, r)
		}
	})

}

// CreateRouter Creates a router you can use to listen and serve
func CreateRouter(routes Routes) *mux.Router {
	var i int
	var router *mux.Router
	var routerPath string
	var ok bool

	routers := make(map[string]*mux.Router, len(routes))
	baseRouter := mux.NewRouter()
	routers[""] = baseRouter

	for _, route := range routes {
		parts := strings.Split(route.Path, "/")
		parts = parts[1:]

		// Finding the closest preexisting router
		for i = range parts {
			routeSlices := parts[1 : len(parts)-i]
			routerPath = strings.Join(routeSlices, "/")
			router, ok = routers[routerPath]
			if ok {
				break
			}
		}

		i = len(parts) - i

		// Create the subrouters we need
		for ; i < len(parts); i++ {
			prefix := fmt.Sprintf("/%s", parts[i-1])
			route.Path = route.Path[len(prefix):]
			router = router.PathPrefix(prefix).Subrouter()
			routerPath = strings.Join(parts[:i], "/")
			routers[routerPath] = router
		}

		createRoute(router, route)
	}

	return baseRouter
}

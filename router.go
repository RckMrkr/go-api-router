package router

import (
	"github.com/gorilla/mux"
	"net/http"
)

// Route type is used to specify a specific endpoint.
type Route struct {
	Name       string
	Method     []string
	Pattern    string
	Handler    http.HandlerFunc
	Middleware []Middleware
}

type Middleware func(http.HandlerFunc) http.HandlerFunc

// Routes is used to bundle Route's
type Routes []Route

// Resource is the "parent" of routes, meaning that the pattern of the Routes will always be prefixed by the Path of the preceding resources.
type Resource struct {
	Path       string
	Resources  Resources
	Routes     Routes
	Middleware []Middleware
}

func (r Resource) resources() Resources {
	if r.Middleware == nil {
		return r.Resources
	}

	for i := range r.Resources {
		r.Resources[i].Middleware = append(r.Middleware, r.Resources[i].Middleware...)
	}

	return r.Resources
}

// Resources is used to bundle Resource's
type Resources []Resource

// CreateRouter Creates a router you can use to listen and serve
func CreateRouter(resources Resources) *mux.Router {
	r := mux.NewRouter().StrictSlash(true)
	createResources(resources, r)
	return r
}

func createResources(resources Resources, r *mux.Router) {
	if resources == nil {
		return
	}

	c := make(chan bool, len(resources))
	for _, resource := range resources {
		go func() {
			createSubRouter(resource, r)
			c <- true
		}()
	}

	for _ = range resources {
		<-c
	}
}

func createSubRouter(res Resource, r *mux.Router) {
	s := r.Path(res.Path).Subrouter()
	createResources(res.resources(), s)

	if res.Routes == nil {
		return
	}
	for _, route := range res.Routes {
		handler := createHandler(route.Handler, route.Middleware, res.Middleware)
		s.
			Methods(route.Method...).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

}

func createHandler(handler http.HandlerFunc, wrappers ...[]Middleware) http.HandlerFunc {
	for _, middlewares := range wrappers {
		for _, middleware := range middlewares {
			handler = middleware(handler)
		}
	}
	return handler
}

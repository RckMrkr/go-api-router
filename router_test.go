package router

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func middlewareWrapper(str string) func(http.HandlerFunc) http.HandlerFunc {
	return func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, str)
			h.ServeHTTP(w, r)
		}
	}
}

func handlerWrapper(str string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, str)
		return
	}
}

func testRouter() *mux.Router {
	res := Resources{
		Resource{
			Path: "/users/",
			Middleware: []Middleware{
				middlewareWrapper("1"),
				middlewareWrapper("2"),
				middlewareWrapper("3"),
			},
			Routes: Routes{
				Route{
					Pattern: "/",
					Name:    "UserIndex",
					Method:  []string{"GET"},
					Handler: handlerWrapper("6"),
					Middleware: []Middleware{
						middlewareWrapper("4"),
						middlewareWrapper("5"),
					},
				},
			},
			Resources: Resources{
				Resource{
					Path: "/admins/",
					Routes: Routes{
						Route{
							Pattern: "/",
							Name:    "AdminCreate",
							Method:  []string{"GET"},
							Handler: handlerWrapper("7"),
						},
					},
				},
			},
		},
		Resource{
			Path: "/admins/",
			Resources: Resources{
				Resource{
					Path: "/super/",
					Routes: Routes{
						Route{
							Pattern: "/",
							Name:    "AdminCreate",
							Method:  []string{"GET"},
							Handler: handlerWrapper("7"),
						},
					},
				},
			},
		},
	}

	return CreateRouter(res)
}

func TestAttachMiddleware(t *testing.T) {
	assert := assert.New(t)
	router := testRouter()
	req, _ := http.NewRequest("GET", "http://example.com:43256/users/", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal("123456", w.Body.String())

}

func TestSubResources(t *testing.T) {
	assert := assert.New(t)
	router := testRouter()
	req, _ := http.NewRequest("GET", "http://example.com:43256/admins/super/", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal("7", w.Body.String())
}

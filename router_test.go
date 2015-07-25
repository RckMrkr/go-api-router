package router

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	"github.com/stretchr/testify/assert"
)

func before(str string) func(http.HandlerFunc) http.HandlerFunc {
	return func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, str)
			h.ServeHTTP(w, r)
		}
	}
}

func after(str string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, str)
	}
}

func handler(str string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, str)
		return
	}
}

func testRouter() *mux.Router {
	res := Routes{
		Route{
			Path:    "/before",
			Name:    "Before",
			Methods: []string{"GET"},
			Before:  Before{before("Before")},
			Handler: handler("Handler"),
		},
		Route{
			Path:    "/after",
			Name:    "After",
			Methods: []string{"GET"},
			After:   After{after("After")},
			Handler: handler("Handler"),
		},
		Route{
			Path:    "/scheme",
			Name:    "UserIndex",
			Methods: []string{"GET"},
			Schemes: []string{"https"},
			Handler: handler("Scheme"),
		},
		Route{
			Path:    "/host",
			Name:    "Host",
			Methods: []string{"GET"},
			Host:    "correct.example.org",
			Handler: handler("Host"),
		},
		Route{
			Path:    "/r1/r2/",
			Name:    "FirstSub",
			Methods: []string{"GET"},
			Handler: handler("FirstSub"),
		},
		Route{
			Path:    "/r1/{r2}/r3",
			Name:    "SecondSub",
			Methods: []string{"GET"},
			Handler: handler("SecondSub"),
		},
		Route{
			Path:    "/queries",
			Name:    "Queries",
			Methods: []string{"GET"},
			Handler: handler("Queries"),
			Queries: []string{"key", "correct"},
		},
		Route{
			Path:    "/headers",
			Name:    "Headers",
			Methods: []string{"GET"},
			Handler: handler("Headers"),
			Headers: []string{"X-Test-Header", "Is correct"},
		},
	}

	return CreateRouter(res)
}

func TestShouldAttacheBefore(t *testing.T) {
	assert := assert.New(t)
	router := testRouter()
	req, _ := http.NewRequest("GET", "/before", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal("BeforeHandler", w.Body.String())
}

func TestShouldAttacheAfter(t *testing.T) {
	assert := assert.New(t)
	router := testRouter()
	req, _ := http.NewRequest("GET", "/after", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal("HandlerAfter", w.Body.String())
}

func TestIncorrectSchemeIsNotWorking(t *testing.T) {
	assert := assert.New(t)
	router := testRouter()
	req, _ := http.NewRequest("GET", "http://example.org/scheme", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(404, w.Code)
}

func TestCorrectSchemeIsWorking(t *testing.T) {
	assert := assert.New(t)
	router := testRouter()
	req, _ := http.NewRequest("GET", "https://example.org/scheme", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal("Scheme", w.Body.String())
}

func TestIncorrectHeadersIsNotWorking(t *testing.T) {
	assert := assert.New(t)
	router := testRouter()
	req, _ := http.NewRequest("GET", "/headers", nil)
	req.Header.Set("X-Test-Header", "Is not correct")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(404, w.Code)
}

func TestCorrectHeaderIsWorking(t *testing.T) {
	assert := assert.New(t)
	router := testRouter()
	req, _ := http.NewRequest("GET", "/headers", nil)
	req.Header.Set("X-Test-Header", "Is correct")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal("Headers", w.Body.String())
}

func TestIncorrectQueriesIsNotWorking(t *testing.T) {
	assert := assert.New(t)
	router := testRouter()
	req, _ := http.NewRequest("GET", "/queries?key=incorrect", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(404, w.Code)
}

func TestCorrectQueriesIsWorking(t *testing.T) {
	assert := assert.New(t)
	router := testRouter()
	req, _ := http.NewRequest("GET", "/queries?key=correct", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal("Queries", w.Body.String())
}

func TestCorrectHostIsWorking(t *testing.T) {
	assert := assert.New(t)
	router := testRouter()
	req, _ := http.NewRequest("GET", "http://correct.example.org/host", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal("Host", w.Body.String())
	assert.Equal(200, w.Code)
}

func TestIncorrectHostIsNotWorking(t *testing.T) {
	assert := assert.New(t)
	router := testRouter()
	req, _ := http.NewRequest("GET", "https://incorrect.example.org/host", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(404, w.Code)
}

func TestSubrouters(t *testing.T) {
	assert := assert.New(t)
	router := testRouter()
	req, _ := http.NewRequest("GET", "/r1/r2/", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal("FirstSub", w.Body.String())
	req, _ = http.NewRequest("GET", "/r1/r2/r3", nil)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal("SecondSub", w.Body.String())
}

package gress

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

var (
	handlerOK = func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "hello world")
		w.WriteHeader(http.StatusOK)
	}

	mwUser = func(w http.ResponseWriter, r *http.Request) {
		if r.URL.User == nil || r.URL.User.Username() != "admin" {
			http.Error(w, "", http.StatusUnauthorized)
		}
	}

	mwID = func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get(":id")
		if id == "admin" {
			http.Error(w, "", http.StatusUnauthorized)
		}
	}
)

func TestListen(t *testing.T) {
	app := New()
	go app.Listen(":8080")
}

func TestGressOk(t *testing.T) {

	r, _ := http.NewRequest("GET", "/person/lastname/firstname?learn=golang", nil)
	w := httptest.NewRecorder()

	handler := new(Gress)
	handler.Get("/person/:last/:first", handlerOK)
	handler.ServeHTTP(w, r)

	lastNameParam := r.URL.Query().Get(":last")
	firstNameParam := r.URL.Query().Get(":first")
	learnParam := r.URL.Query().Get("learn")

	if lastNameParam != "lastname" {
		t.Errorf("url param set to [%s]; want [%s]", lastNameParam, "lastname")
	}
	if firstNameParam != "firstname" {
		t.Errorf("url param set to [%s]; want [%s]", firstNameParam, "firstname")
	}
	if learnParam != "golang" {
		t.Errorf("url param set to [%s]; want [%s]", learnParam, "golang")
	}
}

// TestNotFound tests that a 404 code is returned in the
// response if no route matches the request url.
func TestNotFound(t *testing.T) {

	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler := new(Gress)
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("Code set to [%v]; want [%v]", w.Code, http.StatusNotFound)
	}
}

// TestStatic tests the ability to serve static
// content from the filesystem
func TestStatic(t *testing.T) {

	r, _ := http.NewRequest("GET", "/gress_test.go", nil)
	w := httptest.NewRecorder()
	pwd, _ := os.Getwd()

	handler := new(Gress)
	handler.Static("/", pwd)
	handler.ServeHTTP(w, r)

	testFile, _ := ioutil.ReadFile(pwd + "/gress_test.go")
	if w.Body.String() != string(testFile) {
		t.Errorf("handler.Static failed to serve file")
	}
}

// TestUse tests the ability to apply middleware function
func TestUse(t *testing.T) {

	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler := new(Gress)
	handler.Get("/", handlerOK)
	handler.Use(mwUser)
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Did not apply mw. Code set to [%v]; want [%v]", w.Code, http.StatusUnauthorized)
	}

	r, _ = http.NewRequest("GET", "/", nil)
	r.URL.User = url.User("admin")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Code set to [%v]; want [%v]", w.Code, http.StatusOK)
	}
}

// TestUseParam tests the ability to apply middleware
func TestUseParam(t *testing.T) {

	r, _ := http.NewRequest("GET", "/:id", nil)
	w := httptest.NewRecorder()

	// first test that the param filter does not trigger
	handler := new(Gress)
	handler.Get("/", handlerOK)
	handler.Get("/:id", handlerOK)
	handler.UseParam("id", mwID)
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Code set to [%v]; want [%v]", w.Code, http.StatusOK)
	}

	// now test the param filter does trigger
	r, _ = http.NewRequest("GET", "/admin", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Did not apply Param mw. Code set to [%v]; want [%v]", w.Code, http.StatusUnauthorized)
	}

}

// Benchmark_GressdHandler runs a benchmark against
// the Gress using the default settings.
func Benchmark_GressdHandler(b *testing.B) {
	handler := new(Gress)
	handler.Get("/", handlerOK)

	for i := 0; i < b.N; i++ {
		r, _ := http.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
	}
}

// Benchmark_GressdHandler runs a benchmark against
// the Gress using the default settings with REST
// URL params.
func Benchmark_GressdHandlerParams(b *testing.B) {

	app := new(Gress)
	app.Get("/:user", handlerOK)

	for i := 0; i < b.N; i++ {
		r, _ := http.NewRequest("GET", "/admin", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
	}
}

// Benchmark_ServeMux runs a benchmark against
// the ServeMux Go function. We use this to determine
// performance impact of our library, when compared
// to the out-of-the-box Mux provided by Go.
func Benchmark_ServeMux(b *testing.B) {

	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	mux := http.NewServeMux()
	mux.HandleFunc("/", handlerOK)

	for i := 0; i < b.N; i++ {
		mux.ServeHTTP(w, r)
	}
}

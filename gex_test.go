package gex

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

var handlerOk = func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello world")
	w.WriteHeader(http.StatusOK)
}

func TestListen(t *testing.T) {
	mux := New()
	mux.HandleFunc("/api/:id", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL.Query().Get("id"))
		io.WriteString(w, r.URL.Path)
	})
	go mux.Listen(":8080")
}

func TestMux(t *testing.T) {
	r, _ := http.NewRequest("GET", "/person/lastname/firstname?learn=golang", nil)
	w := httptest.NewRecorder()

	handler := New()
	handler.HandleFunc("/person/:last/:first", handlerOk)
	handler.ServeHTTP(w, r)

	lastNameParam := r.URL.Query().Get("last")
	firstNameParam := r.URL.Query().Get("first")
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

func TestNotFound(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler := New()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("Code set to [%v]; want [%v]", w.Code, http.StatusNotFound)
	}
}

func TestStatic(t *testing.T) {

	r, _ := http.NewRequest("GET", "/gex.go", nil)
	w := httptest.NewRecorder()
	pwd, _ := os.Getwd()

	mux := New()
	mux.Static("/", pwd)
	mux.ServeHTTP(w, r)

	b, _ := ioutil.ReadFile(pwd + "/gex.go")
	if w.Body.String() != string(b) {
		t.Errorf("handler.Static failed to serve file")
	}
}

func TestFilter(t *testing.T) {
	w := httptest.NewRecorder()
	mux := New()

	mux.FilterFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "route ")
	})
	mux.FilterFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "ok ")
	})
	mux.HandleFunc("/ok", handlerOk)

	r, _ := http.NewRequest("GET", "/ok", nil)
	mux.ServeHTTP(w, r)
	if w.Body.String() != "route ok hello world" {
		t.Errorf("filter failed")
	}
}

func TestValidFilter(t *testing.T) {
	w := httptest.NewRecorder()
	mux := New()
	mux.FilterFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.User == nil || r.URL.User.Username() != "admin" {
			http.Error(w, "", http.StatusUnauthorized)
		}
	})

	mux.HandleFunc("/ok", handlerOk)

	r, _ := http.NewRequest("GET", "/ok", nil)
	mux.ServeHTTP(w, r)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Did not apply mw. Code set to [%v]; want [%v]", w.Code, http.StatusUnauthorized)
	}
	r.URL.User = url.User("admin")
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Code set to [%v]; want [%v]", w.Code, http.StatusOK)
	}
}

func TestFilterWithRegex(t *testing.T) {
	mux := New()
	mux.FilterFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "route ")
	})
	mux.FilterFunc("/:id", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s ", r.URL.Query().Get("id"))
	})
	mux.HandleFunc("/:id", handlerOk)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/myid", nil)
	mux.ServeHTTP(w, r)
	if w.Body.String() != "route myid hello world" {
		t.Errorf("validfilter failed")
	}
}

func TestSubRouter(t *testing.T) { //TODO
	mux := New()
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "hello world")
	})
	http.Handle("/api/", mux)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/api/hello", nil)
	mux.ServeHTTP(w, r)
	if w.Body.String() != "hello world" {
		t.Errorf("subrouter failed")
	}
}

func BenchmarkGexMux(b *testing.B) {
	mux := New()
	mux.HandleFunc("/", handlerOk)
	mux.HandleFunc("/api/:id", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL.Query().Get("id"))
		io.WriteString(w, r.URL.Path)
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r, _ := http.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
	}
}

func BenchmarkServeMux(b *testing.B) {
	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	mux := http.NewServeMux()
	mux.HandleFunc("/", handlerOk)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mux.ServeHTTP(w, r)
	}
}

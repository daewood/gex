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
	app := New()
	go app.Listen(":8080")
}

func TestGexOk(t *testing.T) {

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
		fmt.Fprintf(w, "middleware route\n")
	})
	mux.FilterFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "middleware ok\n")
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/ok", handlerOk)

	r, _ := http.NewRequest("GET", "/ok", nil)
	mux.ServeHTTP(w, r)
	fmt.Println(w.Body.String())
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
	w := httptest.NewRecorder()
	mux := New()
	mux.FilterFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("middleware route\n")
	})
	mux.FilterFunc("/:id", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("mw:" + r.URL.Query().Get("id"))
		fmt.Fprintf(w, "middleware id\n")
	})
	mux.HandleFunc("/:id", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL.Query().Get("id"))
		io.WriteString(w, r.URL.Path)
	})
	{
		r, _ := http.NewRequest("GET", "/myid", nil)
		mux.ServeHTTP(w, r)
		fmt.Println(w.Body.String())
	}
}

func TestGexMux(t *testing.T) {
	w := httptest.NewRecorder()
	mux := New()
	mux.HandleFunc("/api/:id", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL.Query().Get("id"))
		io.WriteString(w, r.URL.Path)
	})
	r, _ := http.NewRequest("GET", "/api/myid", nil)
	mux.ServeHTTP(w, r)
	fmt.Println(w.Body.String())
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

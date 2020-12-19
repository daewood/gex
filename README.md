# gex
extend http.ServeMux to support parameterized routes and filters

## Getting Started

    package main

    import (
        "fmt"
        "github.com/daewood/gex"
        "net/http"
    )

    func Whoami(w http.ResponseWriter, r *http.Request) {
        params := r.URL.Query()
        lastName := params.Get("last")
        firstName := params.Get("first")
        fmt.Fprintf(w, "you are %s %s", firstName, lastName)
    }

    func main() {
        app := gex.New()
        app.HandleFunc("/:last/:first", Whoami)

        app.Listen(":8080")
    }

### Work with standard http
    package main

    import (
        "fmt"
        "github.com/daewood/gex"
        "net/http"
    )
    func main() {
        mux := gex.New()
        mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
            fmt.Fprintf(w, "hello world")
        })

        http.ListenAndServe(":8080", mux)
    }

### Static Examples

    pwd, _ := os.Getwd()
    app.Static("/static", pwd)

this will serve any files in `/static`, including files in subdirectories. For example `/static/logo.gif` or `/static/style/main.css`.

## Middleware
You can apply middleware to gex, which is useful for enforcing security,
redirects, etc.

You can, for example, filter all request to enforce some type of security:

    var mwUser = func(w http.ResponseWriter, r *http.Request) {
    	if r.URL.User == nil || r.URL.User.Username() != "admin" {
    		http.Error(w, "", http.StatusUnauthorized)
    	}
    }

    app.FilterFunc("/", mwUser)
    app.Handle("/:id", handler)

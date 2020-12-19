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

### As a mux

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
        http.Handle("/", mux)
        http.ListenAndServe(":8080", nil)
    }

### Static Examples

    package main

    import (
        "fmt"
        "github.com/daewood/gex"
        "net/http"
    )
    func main() {
        app := gex.New()
        api.Filter("/", handlerLog)           //a http.Handler, implement by yourself
        api.FilterFunc("/api/", validatFunc)  //a func(http.ResponseWriter, *http.Request), implement by yourself
        api.Handle("/api/", handlerApi)       //a http.Handler, implement by yourself
        
        pwd, _ := os.Getwd()
        app.Static("/static", pwd)
        app.Listen(":8080")
    }

this will serve any files in `/static`, including files in subdirectories. For example `/static/logo.gif` or `/static/style/main.css`.

## Middleware
You can apply middleware to gex, which is useful for enforcing security,
redirects, etc.

You can, for example, filter all request to enforce some type of security:

    package main

    import (
        "fmt"
        "github.com/daewood/gex"
        "net/http"
    )
    func main() {
        app := gex.New()
        var mwUser = func(w http.ResponseWriter, r *http.Request) {
            if r.URL.User == nil || r.URL.User.Username() != "admin" {
                http.Error(w, "", http.StatusUnauthorized)
            }
        }

        app.FilterFunc("/", mwUser)
        app.Filter("/hello", filterHandler) // filterHandler is a http.Handler, implement by yourself
        mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
            fmt.Fprintf(w, "hello world")
        })
        app.Listen(":8080")
    }

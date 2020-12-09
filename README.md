# Gress
express like go lightweight web app framework

## Getting Started

    package main

    import (
        "fmt"
        "github.com/daewood/gress"
        "net/http"
    )

    func Whoami(w http.ResponseWriter, r *http.Request) {
        params := r.URL.Query()
        lastName := params.Get(":last")
        firstName := params.Get(":first")
        fmt.Fprintf(w, "you are %s %s", firstName, lastName)
    }

    func main() {
        app := gress.New()
        app.Get("/:last/:first", Whoami)

        app.Listen(":8080")
    }

### Route Examples
You can create routes for all http methods:

    app.Get("/:param", handler)
    app.Put("/:param", handler)
    app.Post("/:param", handler)
    app.Patch("/:param", handler)
    app.Delete("/:param", handler)

You can specify custom regular expressions for routes:

    app.Get("/files/:param(.+)", handler)

You can also create routes for static files:

    pwd, _ := os.Getwd()
    app.Static("/static", pwd)

this will serve any files in `/static`, including files in subdirectories. For example `/static/logo.gif` or `/static/style/main.css`.

## Middleware
You can apply middleware to gress, which is useful for enforcing security,
redirects, etc.

You can, for example, filter all request to enforce some type of security:

    var mwUser = func(w http.ResponseWriter, r *http.Request) {
    	if r.URL.User == nil || r.URL.User.Username() != "admin" {
    		http.Error(w, "", http.StatusUnauthorized)
    	}
    }

    r.Use(mwUser)

You can also apply mw only when certain REST URL Parameters exist:

    r.Get("/:id", handler)
    r.UseParam("id", func(rw http.ResponseWriter, r *http.Request) {
		...
	})
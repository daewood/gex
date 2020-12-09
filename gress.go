package gress

import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type route struct {
	method  string
	regex   *regexp.Regexp
	params  map[int]string
	handler http.HandlerFunc
}

// Gress ...
type Gress struct {
	routes []*route
	mws    []http.HandlerFunc
}

// New ...
func New() *Gress {
	return &Gress{}
}

// Listen ...
func (m *Gress) Listen(addr string) {
	http.ListenAndServe(addr, m)
}

// Get adds a new Route for GET requests.
func (m *Gress) Get(pattern string, handler http.HandlerFunc) *Gress {
	return m.addRoute("GET", pattern, handler)
}

// Put adds a new Route for PUT requests.
func (m *Gress) Put(pattern string, handler http.HandlerFunc) *Gress {
	return m.addRoute("PUT", pattern, handler)
}

// Delete adds a new Route for DELETE requests.
func (m *Gress) Delete(pattern string, handler http.HandlerFunc) *Gress {
	return m.addRoute("DELETE", pattern, handler)
}

// Patch adds a new Route for PATCH requests.
func (m *Gress) Patch(pattern string, handler http.HandlerFunc) *Gress {
	return m.addRoute("PATCH", pattern, handler)
}

// Post adds a new Route for POST requests.
func (m *Gress) Post(pattern string, handler http.HandlerFunc) *Gress {
	return m.addRoute("POST", pattern, handler)
}

// Static files from the specified directory
func (m *Gress) Static(pattern string, dir string) *Gress {
	//append a regex to the param to match everything
	// that comes after the prefix
	pattern = pattern + "(.+)"
	return m.addRoute("GET", pattern, func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Clean(r.URL.Path)
		path = filepath.Join(dir, path)
		http.ServeFile(w, r, path)
	})
}

// addRoute Adds a new Route to the Handler
func (m *Gress) addRoute(method string, pattern string, handler http.HandlerFunc) *Gress {

	//split the url into sections
	parts := strings.Split(pattern, "/")

	//find params that start with ":"
	//replace with regular expressions
	j := 0
	params := make(map[int]string)
	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			expr := "([^/]+)"
			//a user may choose to override the defult expression
			// similar to expressjs: ‘/user/:id([0-9]+)’
			if index := strings.Index(part, "("); index != -1 {
				expr = part[index:]
				part = part[:index]
			}
			params[j] = part
			parts[i] = expr
			j++
		}
	}

	//recreate the url pattern, with parameters replaced
	//by regular expressions. then compile the regex
	pattern = strings.Join(parts, "/")
	regex, regexErr := regexp.Compile(pattern)
	if regexErr != nil {
		//TODO add error handling here to avoid panic
		panic(regexErr)
	}

	//now create the Route
	route := &route{}
	route.method = method
	route.regex = regex
	route.handler = handler
	route.params = params

	//and finally append to the list of Routes
	m.routes = append(m.routes, route)

	return m
}

// Use middleware adds the middleware filter.
func (m *Gress) Use(filter http.HandlerFunc) {
	m.mws = append(m.mws, filter)
}

// UseParam adds the middleware if the REST URL parameter exists.
func (m *Gress) UseParam(param string, filter http.HandlerFunc) {
	if !strings.HasPrefix(param, ":") {
		param = ":" + param
	}

	m.Use(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Query().Get(param)
		if len(p) > 0 {
			filter(w, r)
		}
	})
}

// ServeHTTP http.Handler interface
func (m *Gress) ServeHTTP(rw http.ResponseWriter, r *http.Request) {

	requestPath := r.URL.Path

	//wrap the response writer, in our custom interface
	w := &responseWriter{writer: rw}

	//find a matching Route
	for _, route := range m.routes {

		//if the methods don't match, skip this handler
		//i.e if request.Method is 'PUT' Route.Method must be 'PUT'
		if r.Method != route.method {
			continue
		}

		//check if Route pattern matches url
		if !route.regex.MatchString(requestPath) {
			continue
		}

		//get submatches (params)
		matches := route.regex.FindStringSubmatch(requestPath)

		//double check that the Route matches the URL pattern.
		if len(matches[0]) != len(requestPath) {
			continue
		}

		if len(route.params) > 0 {
			//add url parameters to the query param map
			values := r.URL.Query()
			for i, match := range matches[1:] {
				values.Add(route.params[i], match)
			}

			//reassemble query params and add to RawQuery
			r.URL.RawQuery = url.Values(values).Encode() + "&" + r.URL.RawQuery
			//r.URL.RawQuery = url.Values(values).Encode()
		}

		//execute middleware mws
		for _, filter := range m.mws {
			filter(w, r)
			if w.started {
				return
			}
		}

		//Invoke the request handler
		route.handler(w, r)
		break
	}

	//if no matches to url, throw a not found exception
	if w.started == false {
		http.NotFound(w, r)
	}
}

type responseWriter struct {
	writer  http.ResponseWriter
	started bool
	status  int
}

// Header returns the header map that will be sent by WriteHeader.
func (w *responseWriter) Header() http.Header {
	return w.writer.Header()
}

// Write writes the data to the connection as part of an HTTP reply,
// and sets `started` to true
func (w *responseWriter) Write(p []byte) (int, error) {
	w.started = true
	return w.writer.Write(p)
}

// WriteHeader sends an HTTP response header with status code,
// and sets `started` to true
func (w *responseWriter) WriteHeader(code int) {
	w.status = code
	w.started = true
	w.writer.WriteHeader(code)
}

// SendJSON ...
func SendJSON(w http.ResponseWriter, v interface{}) {
	content, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Length", strconv.Itoa(len(content)))
	w.Header().Set("Content-Type", "application/json")
	w.Write(content)
}

// ReadJSON ...
func ReadJSON(r *http.Request, v interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

// SendXML ...
func SendXML(w http.ResponseWriter, v interface{}) {
	content, err := xml.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Length", strconv.Itoa(len(content)))
	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	w.Write(content)
}

// ReadXML ...
func ReadXML(r *http.Request, v interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		return err
	}
	return xml.Unmarshal(body, v)
}

// Send ...
func Send(w http.ResponseWriter, r *http.Request, v interface{}) {
	accept := r.Header.Get("Accept")
	switch accept {
	case "application/json":
		SendJSON(w, v)
	case "application/xml", "text/xml":
		SendXML(w, v)
	default:
		SendJSON(w, v)
	}
}

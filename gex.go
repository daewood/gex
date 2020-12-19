package gex

import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type route struct {
	pattern string
	method  string
	handler http.Handler
	params  map[int]string
	regex   *regexp.Regexp
	filters []http.Handler
}

// Mux an extent router of http.ServeMux
type Mux struct {
	*http.ServeMux
	routes  []*route
	filters map[string][]http.Handler
}

// New allocates and returns a new Mux.
func New() *Mux {
	return &Mux{ServeMux: http.NewServeMux()}
}

// ServeHTTP dispatches the request to the handler
func (mux *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	//wrap the response writer, in our custom interface
	rw := &responseWriter{ResponseWriter: w}
	for p, filters := range mux.filters {
		if strings.HasPrefix(path, p) {
			for _, filter := range filters {
				filter.ServeHTTP(rw, r)
				if rw.started {
					return
				}
			}
		}
	}
	for _, route := range mux.routes {
		if route.regex.MatchString(path) {
			// fmt.Println(route.params)
			if len(route.params) > 0 {
				//get submatches (params)
				matches := route.regex.FindStringSubmatch(path)
				//add url parameters to the query param map
				values := r.URL.Query()
				for i, match := range matches[1:] {
					param := route.params[i][1:]
					values.Add(param, match)
				}
				r.URL.RawQuery = url.Values(values).Encode()
			}
			for _, filter := range route.filters {
				filter.ServeHTTP(rw, r)
				if rw.started {
					return
				}
			}
			if route.handler != nil {
				route.handler.ServeHTTP(rw, r)
				return
			}
		}
	}

	mux.ServeMux.ServeHTTP(rw, r)
}

// Listen run on the address
func (mux *Mux) Listen(addr string) error {
	return http.ListenAndServe(addr, mux)
}

// ListenTLS run on the address with cert and key files
func (mux *Mux) ListenTLS(addr, certFile, keyFile string) error {
	return http.ListenAndServeTLS(addr, certFile, keyFile, mux)
}

// Static files from the specified directory
func (mux *Mux) Static(pattern string, dir string) {
	mux.Handle(pattern, http.FileServer(http.Dir(dir)))
}

func (mux *Mux) addRoute(pattern string, handler http.Handler, isFilter bool) {
	for _, route := range mux.routes {
		if route.pattern == pattern { //update filter or handler
			if isFilter {
				route.filters = append(route.filters, handler)
			} else {
				route.handler = handler
			}
			return
		}
	}
	//split the url into sections
	parts := strings.Split(pattern, "/")
	//find params that start with ":"
	//replace with regular expressions
	j := 0
	params := make(map[int]string)
	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			expr := "([^/]+)"
			// similar to expressjs: ‘/user/:id([0-9]+)’, strip off
			if index := strings.Index(part, "("); index != -1 {
				expr = part[index:]
				part = part[:index]
			}
			params[j] = part
			parts[i] = expr
			j++
		}
	}

	var regex *regexp.Regexp
	if j > 0 {
		path := strings.Join(parts, "/")
		regex = regexp.MustCompile(path)
	}
	route := &route{pattern: pattern, params: params, regex: regex}
	if isFilter {
		route.filters = append(route.filters, handler)
	} else {
		route.handler = handler
	}
	mux.routes = append(mux.routes, route)
}

// Handle registers the handler for the given pattern.
func (mux *Mux) Handle(pattern string, handler http.Handler) {
	if strings.Contains(pattern, "/:") {
		mux.addRoute(pattern, handler, false)
	}
	mux.ServeMux.Handle(pattern, handler)
}

// HandleFunc registers the handler function for the given pattern.
func (mux *Mux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	if handler == nil {
		panic("http: nil handler")
	}
	mux.Handle(pattern, http.HandlerFunc(handler))
}

// Filter registers the handler for the given pattern.
func (mux *Mux) Filter(pattern string, handler http.Handler) {
	if handler == nil {
		panic("http: nil handler")
	}
	if strings.Contains(pattern, "/:") {
		mux.addRoute(pattern, handler, true)
		return
	}
	if mux.filters == nil {
		mux.filters = make(map[string][]http.Handler)
	}
	mux.filters[pattern] = append(mux.filters[pattern], handler)
}

// FilterFunc registers the handler function for the given pattern.
func (mux *Mux) FilterFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	if handler == nil {
		panic("http: nil handler")
	}
	mux.Filter(pattern, http.HandlerFunc(handler))
}

////////////////////////////
type responseWriter struct {
	http.ResponseWriter
	started bool
	status  int
}

// WriteHeader sends an HTTP response header with status code,
// and sets `started` to true
func (w *responseWriter) WriteHeader(code int) {
	w.status = code
	w.started = true
	w.ResponseWriter.WriteHeader(code)
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

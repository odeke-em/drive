// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package websrv provides a stylized way to write HTTP handlers and run an HTTP server.
//
// Overview
//
// An HTTP server using package websrv registers handlers by calling Handle(pattern, handler).
// The top-level Handler is a standard HTTP handler that dispatches to the registered websrv handlers.
// Handler itself must be registered explicitly with the standard HTTP package:
//
//	package main
//
//	import (
//		"net/http"
//		"code.google.com/p/rsc/websrv"
//	)
//
//	func main() {
//		websrv.Handle("/doc/{DocID}", new(docRequest))
//		http.Handle("/", websrv.Handler)
//		websrv.Serve()
//	}
//
// This package does not automatically register websrv.Handler because client code might
// wish to use its own top-level handler to pick off some requests, or to add custom logging
// or authentication before invoking websrv.Handler.
//
// Patterns
//
// The first argument to Handle is a pattern specifying the set of paths for which the handler
// should be invoked.
//
// The simplest pattern is a path, like /hello/world. Such a path matches exactly that URL and
// no others. Like in package http, the path may begin with a host name to restrict the matches
// to URLs on a specific host.
//
// A pattern may specify a wildcard - a name enclosed in braces - for any path element.
// For example, /hello/{Name} matches /hello/world, /hello/ken, and /hello/123, but
// not /hello/two/elements.
//
// Handlers
//
// A handler is a value with any of these four methods:
//
//	Get(ctxt *Context)
//	Post(ctxt *Context)
//	Put(ctxt *Context)
//	Delete(ctxt *Context)
//
// If an incoming request maps to a handler without the requested method (for example,
// a DELETE request arrives for a handler that implements only Get), the server replies
// with an HTTP "method not allowed" error, without using the handler at all.
//
// If the handler object is a pointer to a struct, additional processing happens before
// invoking the corresponding service method. First, the server makes a copy of the
// struct (or, if the pointer is nil, allocates a zero struct) and uses a pointer to that copy,
// not the original, when invoking the handler. This gives each request some local
// storage. Second, after making the copy, the server looks for fields with names matching
// the wildcards, and for each such field copies the corresponding path element into
// the field. The path element must be convertible into the field.
//
// For example, consider:
//
//	type Hello struct {
//		Name string
//	}
//
//	type Page struct {
//		Format string
//		ID int
//	}
//
//	websrv.Handle("/hello/{Name}", (*Hello)(nil))
//	websrv.Handle("/page/{ID}", &Page{Format: "text"})
//
// The server will prepare the receiver for the URL /hello/world by allocating a new
// Hello struct and setting its Name field to "world".
// The server will prepare the receiver for the URL /page/5 by making a copy of
// the prototype Page struct and setting ID to 5, so that the receiver is effectively
// &Page{Format: "text", ID: 5}.
//
// After preparing the receiver from the URL, the server turns to the posted form data
// and URL query parameters. A form value or query parameter with a name matching
// a struct field will be copied into that struct field. For example, the URL /page/10?Format=html
// would result in a receiver value &Page{Format: "html", ID: 10}.
// URL path wildcards have the highest priority: /page/10?ID=11 matching /page/{ID}
// ignores the ID=11, because the wildcard has already claimed ID.
// Otherwise, posted field values take precedence over query parameters,
// and the first value is the one recorded. An exception to that rule, a field of slice
// type records all vales, first from the posted form data and then from the URL query.
//
// If a struct field has a tag with a "websrv" key, that key's value specifies a comma-separated
// list in which the first item is a name to use instead of the struct field name,
// and the remainder are attributes. The only attribute is "post", which indicates that
// a field must be filled using posted data, not URL query parameters.
//
// For example:
//
//	type Page struct {
//		Format string `websrv:"fmt"`
//		ID int
//	}
//
// will match /page/10?fmt=html instead of /page/10?Format=html.
// If the tag said "fmt,post", the fmt= in the URL would be ignored entirely,
// and only a posted fmt value could be recorded.
//
// As a final step in preparing the receiver, if the posted body has content type "application/json",
// the body is treated as JSON and unmarshaled into the receiver using the standard
// encoding/json package. As such it does not respect websrv tags, and it overwrites
// values set by URL patterns or query parameters. This form is intended to support
// using jQuery to post JSON objects.
//
// After completing that initialization (or skipping it, if the receiver is not a pointer to
// a struct), the service invokes the corresponding method. During that invocation,
// the Context argument provides access to additional request details and also enables
// sending the response.
//
// Contexts
//
// A Context represents both the incoming HTTP request and the response being prepared.
// It provides access to the request as ctxt.Request, and it implements http.ResponseWriter,
// meaning it has Header, Write, and WriteHeader methods. However, by default the context
// buffers the response in memory, only sending it when the handler method returns.
// To bypass the buffer, invoke the DisableWriteBuffer method.
//
// If a handler panics, the server discards the response buffered in the Context and instead
// sends an HTTP 500 error containing the panic text and a stack trace.
//
package websrv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

// Handle registers fn to handle paths of the form given by pattern.
func Handle(pattern string, fn interface{}) {
	elems := strings.Split(pattern, "/")
	n := len(elems)
	if n == 0 {
		panic("empty pattern")
	}
	if n >= 2 && elems[n-2] == "{" && strings.HasSuffix(elems[n-1], "}") {
		elems[n-2] += "/" + elems[n-1]
		n--
		elems = elems[:n]
	}

	var vars []string
	re := "^"
	for i, elem := range elems {
		if i > 0 {
			re += "/"
		}
		wildcard := false
		if strings.HasPrefix(elem, "{") && strings.HasSuffix(elem, "}") {
			elem = elem[1 : len(elem)-1]
			wildcard = true
		}
		if strings.ContainsAny(elem, "{}") {
			panic("invalid pattern: " + pattern)
		}
		if i == 0 {
			if wildcard {
				panic("wildcard not allowed for host")
			}
			if elem == "" {
				re += "[^/]*"
			} else {
				re += `(?:[^/]+\.)?` + regexp.QuoteMeta(elem)
			}
			continue
		}
		if wildcard {
			if i == len(elems)-1 && strings.HasPrefix(elem, "/") {
				elem = elem[1:]
				re += "(.*)"
			} else {
				re += "([^/]+)"
			}
			vars = append(vars, elem)
		} else {
			re += regexp.QuoteMeta(elem)
		}
	}
	re += "$"

	h := &handler{
		regexp.MustCompile(re),
		vars,
		nil,
		fn,
	}

	fmt.Printf("h %s\n", re)

	if len(vars) > 0 {
		h.isVar = make(map[string]bool)
		for _, v := range vars {
			h.isVar[v] = true
		}
	}

	handlers = append(handlers, h)
}

type handler struct {
	re    *regexp.Regexp
	vars  []string
	isVar map[string]bool
	impl  interface{}
}

var handlers []*handler

// Handler is an http.Handler that dispatches to the registered websrv handlers.
var Handler http.Handler = http.HandlerFunc(handle)

// A Context provides access to an incoming HTTP request and collects
// the outgoing response. By default, the response is buffered in memory
// until the handler returns.
type Context struct {
	Request *http.Request
	w       http.ResponseWriter
	buf     bytes.Buffer
	status  int
	direct  bool
}

func (ctxt *Context) handlePanic() {
	// TODO: catch early writes
	if err := recover(); err != nil {
		ctxt.w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		ctxt.w.WriteHeader(500)
		fmt.Fprintf(ctxt.w, "panic: %s\n", err)
		buf := make([]byte, 10000)
		n := runtime.Stack(buf, false)
		ctxt.w.Write(buf[:n])
	}
}

// Header returns the HTTP header being prepared.
func (ctxt *Context) Header() http.Header {
	return ctxt.w.Header()
}

func (ctxt *Context) flush() {
	if ctxt.direct {
		return
	}
	if ctxt.status != 0 {
		ctxt.w.WriteHeader(ctxt.status)
	}
	ctxt.w.Write(ctxt.buf.Bytes())
	ctxt.buf.Reset()
}

// DisableWriteBuffer disables the response write buffer.
// Any buffered writes are sent to the underlying ResponseWriter,
// and any future writes bypass the buffer.
func (ctxt *Context) DisableWriteBuffer() {
	ctxt.flush()
	ctxt.direct = true
}

// Write writes HTTP response data.
func (ctxt *Context) Write(p []byte) (n int, err error) {
	if ctxt.direct {
		return ctxt.w.Write(p)
	}
	return ctxt.buf.Write(p)
}

// Redirect sends a redirect to the given URL.
func (ctxt *Context) Redirect(url string) {
	http.Redirect(ctxt, ctxt.Request, url, http.StatusFound)
}

// ServeFile serves the request using the file at the given path.
func (ctxt *Context) ServeFile(path string) {
	http.ServeFile(ctxt, ctxt.Request, path)
}

// WriteHeader writes the HTTP header, with the given status.
func (ctxt *Context) WriteHeader(status int) {
	if ctxt.status != 0 {
		return
	}
	ctxt.status = status
}

// ServeJSON serves the request using the JSON encoding of the
// given data. It sets the content type of the response to "application/json".
func (ctxt *Context) ServeJSON(data interface{}) {
	encoded, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	ctxt.Header().Set("Content-Type", "application/json")
	ctxt.Write(encoded)
}

func handle(w http.ResponseWriter, req *http.Request) {
	ctxt := &Context{Request: req, w: w}
	defer ctxt.handlePanic()

	what := req.Host + req.URL.Path
	fmt.Printf("try %s\n", what)
	for _, h := range handlers {
		m := h.re.FindStringSubmatch(what)
		if m != nil {
			fmt.Printf("found\n")
			h.serve(m, ctxt)
			ctxt.flush()
			return
		}
	}
	http.NotFound(w, req)
}

func (h *handler) serve(m []string, ctxt *Context) {
	impl := reflect.ValueOf(h.impl)
	if impl.Type().Kind() == reflect.Ptr {
		if !impl.IsNil() {
			orig := impl
			impl = reflect.New(impl.Type().Elem())
			impl.Elem().Set(orig.Elem())
		} else {
			impl = reflect.New(impl.Type().Elem())
		}
		if impl.Elem().Kind() == reflect.Struct {
			for i, v := range h.vars {
				if err := h.setField(impl.Elem(), v, m[i+1], -1, false); err != nil {
					panic(err)
				}
			}
			ctxt.Request.ParseForm()
			for key, vals := range ctxt.Request.Form {
				if h.isVar[key] {
					continue
				}
				npost := len(ctxt.Request.PostForm[key])
				for i, v := range vals {
					if err := h.setField(impl.Elem(), key, v, i, i < npost); err != nil {
						panic(err)
					}
				}
			}
			ct := ctxt.Request.Header.Get("Content-Type")
			ct, _, _ = mime.ParseMediaType(ct)
			if ct == "application/json" {
				data, err := ioutil.ReadAll(ctxt.Request.Body)
				if err != nil {
					panic(err)
				}
				println(string(data), impl.Type().String(), ctxt.Request.Method)
				if err := json.Unmarshal(data, impl.Interface()); err != nil {
					println("json", err.Error())
					panic(err)
				}
				println("did json")
			}
		}
	}

	println("dispatch")
	v := impl.Interface()
	switch ctxt.Request.Method {
	case "GET":
		if v, ok := v.(getter); ok {
			v.Get(ctxt)
			return
		}

	case "POST":
		if v, ok := v.(poster); ok {
			v.Post(ctxt)
			return
		}
		println("NO POST")

	case "PUT":
		if v, ok := v.(putter); ok {
			v.Put(ctxt)
			return
		}

	case "DELETE":
		if v, ok := v.(deleter); ok {
			v.Delete(ctxt)
			return
		}
	}

	http.Error(ctxt, "unsupported method", http.StatusMethodNotAllowed)
}

func (h *handler) setField(v reflect.Value, name, val string, index int, isPost bool) error {
	// TODO: embedded
	// TODO: do once
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" {
			continue
		}
		tag := f.Tag.Get("websrv")
		if tag == "-" {
			continue
		}
		xname := f.Name
		x := strings.Split(tag, ",")
		if x[0] != "" {
			xname = x[0]
		}
		if xname != name {
			continue
		}
		if h.isVar[name] && index != -1 {
			return nil
		}
		for _, attr := range x[1:] {
			switch attr {
			case "post":
				if !isPost {
					return nil
				}
			}
		}

		fv := v.Field(i)
		if index > 0 && fv.Kind() != reflect.Slice {
			return nil
		}

		switch fv.Kind() {
		case reflect.Slice:
			// TODO: not byte slice
			if index < 1 {
				fv.Set(reflect.MakeSlice(fv.Type().Elem(), 1, 1))
				fv = fv.Index(0)
			} else {
				n := fv.Len()
				fv.Set(reflect.Append(fv, reflect.Zero(fv.Type().Elem())))
				fv = fv.Index(n)
			}
		}

		switch fv.Kind() {
		case reflect.Ptr:
			if fv.IsNil() {
				fv.Set(reflect.New(fv.Type().Elem()))
			}
			fv = fv.Elem()
		}

		if pv, ok := fv.Addr().Interface().(parser); ok {
			return pv.Parse(val)
		}

		// TODO: interface check

		switch fv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			x, err := strconv.ParseInt(val, 0, fv.Type().Bits())
			if err != nil {
				return fmt.Errorf("invalid int %s=%s: %v", name, val, err)
			}
			println("parsed", val, "as", x)
			fv.SetInt(x)

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			x, err := strconv.ParseUint(val, 0, fv.Type().Bits())
			if err != nil {
				return fmt.Errorf("invalid uint %s=%s: %v", name, val, err)
			}
			fv.SetUint(x)

		case reflect.Float32, reflect.Float64:
			x, err := strconv.ParseFloat(val, fv.Type().Bits())
			if err != nil {
				return fmt.Errorf("invalid float %s=%s: %v", name, val, err)
			}
			fv.SetFloat(x)

		case reflect.String:
			fv.SetString(val)
		}
	}

	return nil
}

type parser interface {
	Parse(string) error
}

type getter interface {
	Get(ctxt *Context)
}

type poster interface {
	Post(ctxt *Context)
}

type putter interface {
	Put(ctxt *Context)
}

type deleter interface {
	Delete(ctxt *Context)
}

// Serve runs an HTTP server on addr.
// Serve never returns: if the HTTP server returns, Serve prints the returned error and exits.
func Serve(addr string) {
	log.Fatal(http.ListenAndServe(addr, nil))
}

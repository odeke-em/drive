// Copyright 2011 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package app

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"appengine"
	"appengine/memcache"
	"appengine/urlfetch"

	"code.google.com/p/rsc/appfs/fs"
	_ "code.google.com/p/rsc/appfs/server"
	_ "code.google.com/p/rsc/blog/post"
	"code.google.com/p/rsc/issue/dashboard"
)

func init() {
	http.HandleFunc("/admin/", Admin)
	http.HandleFunc("/admin/dashboard/", AdminDashboard)
	http.HandleFunc("/dashboard/", Dashboard)
}

func Admin(w http.ResponseWriter, req *http.Request) {
	c := appengine.NewContext(req)
	switch req.FormValue("op") {
	default:
		fmt.Fprintf(w, "unknown op %s\n", req.FormValue("op"))
	case "memcache-get":
		key := req.FormValue("key")
		item, err := memcache.Get(c, key)
		if err != nil {
			fmt.Fprintf(w, "ERROR: %s\n", err)
			return
		}
		w.Write(item.Value)
	case "memcache-delete":
		key := req.FormValue("key")
		if err := memcache.Delete(c, key); err != nil {
			fmt.Fprintf(w, "ERROR: %s\n", err)
			return
		}
		fmt.Fprintf(w, "deleted %s\n", key)
	}
}

func Dashboard(w http.ResponseWriter, req *http.Request) {
	httpCache(w, 5*time.Minute)
	ctxt := fs.NewContext(req)
	ctxt.ServeFile(w, req, "issue-dashboard/" + strings.TrimPrefix(req.URL.Path, "/dashboard/"))
}

func AdminDashboard(w http.ResponseWriter, req *http.Request) {
	version := "Go " + strings.TrimPrefix(req.URL.Path, "/admin/dashboard/")
	ctxt := fs.NewContext(req)
	ctxt.Mkdir("issue-dashboard")
	c := appengine.NewContext(req)
	if err := dashboard.Update(ctxt, urlfetch.Client(c), version); err != nil {
		fmt.Fprintf(w, "Error updating: %s\n", err)
	} else {
		fmt.Fprintf(w, "Updated.\n")
	}
}

func httpCache(w http.ResponseWriter, dt time.Duration) {
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", int(dt.Seconds())))
}

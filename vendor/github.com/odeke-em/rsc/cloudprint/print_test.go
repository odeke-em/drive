// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cloudprint

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"syscall"
	"testing"
	"time"
)

var authBlob Auth
var authClient Auth
var authServer Auth
var client *Client

func init() {
	data, err := ioutil.ReadFile("authblob")
	if err != nil {
		log.Fatal(err)
	}
	if err := json.Unmarshal(data, &authBlob); err != nil {
		log.Fatal(err)
	}

	f, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		log.Fatal(err)
	}
	syscall.Dup2(int(f.Fd()), 0)
	syscall.Dup2(int(f.Fd()), 1)
	f.Close()

	authClient, err = AuthUser(authBlob)
	if err != nil {
		log.Fatalf("AuthUser: %v", err)
	}

	authServer = authClient
	authServer.ProxyID, err = RandomID()
	if err != nil {
		log.Fatalf("AuthServer: %v", err)
	}

	c, err := NewClient(authClient)
	if err != nil {
		log.Fatalf("NewClient: %v", err)
	}
	if _, err := c.Search("", false); err != nil {
		log.Fatalf("Client.Search: %v", err)
	}
	client = c
}

func TestCreatePrinter(t *testing.T) {
	srv, err := NewServer(authServer)
	if err != nil {
		t.Fatal(err)
	}
	name := fmt.Sprintf("TestCreatePrinter at %v", time.Now().Format(time.UnixDate))
	p, err := srv.CreatePrinter(&PrinterInfo{Name: name})
	if err != nil {
		t.Fatal(err)
	}
	defer srv.DeletePrinter(p)

	testPrinter(t, srv, "CreatePrinter", p)
}

func TestOpenPrinter(t *testing.T) {
	var err error
	auth := authBlob
	auth.ProxyID, err = RandomID()
	if err != nil {
		t.Fatal(err)
	}
	name := fmt.Sprintf("TestOpenPrinter at %v", time.Now().Format(time.UnixDate))
	op, err := CreateOpenPrinter(auth, &PrinterInfo{Name: name})
	if err != nil {
		t.Fatal(err)
	}

	if err := op.VerifyClaim(); err == nil {
		t.Fatal("VerifyClaim succeeded immediately after CreateOpenPrinter")
	}

	fmt.Printf("Please visit this URL to claim the test printer:\n%s\n", op.ClaimURL())
	fmt.Printf("Type enter when done.\n")
	bufio.NewScanner(os.Stdin).Scan()

	if err := op.VerifyClaim(); err != nil {
		t.Fatalf("VerifyClaim: %v", err)
	}

	auth = op.Auth()

	srv, err := NewServer(auth)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	defer srv.DeletePrinter(op.Printer())

	testPrinter(t, srv, "CreateOpenPrinter", op.Printer())
}

func testPrinter(t *testing.T, srv *Server, how string, p *Printer) {
	data, err := ioutil.ReadFile("helloworld.pdf")
	if err != nil {
		t.Fatal(err)
	}

	p1, err := client.Printer(p.ID)
	if err != nil {
		t.Fatalf("Client.Printer: %v", err)
	}

	info := &JobInfo{
		Title: "Test Job Title",
		Tags:  []string{"test"},
	}
	j, err := client.Print(p1, info, data)
	if err != nil {
		t.Fatalf("%s, Client.Print: %v", how, err)
	}

	jobs, err := client.Jobs(p1)
	if err != nil {
		t.Fatalf("%s, Client.Jobs: %v", how, err)
	}
	if len(jobs) != 1 {
		t.Fatalf("%s, Client.Jobs() returned %d jobs, want 1", how, len(jobs))
	}
	if jobs[0].ID != j.ID {
		t.Fatalf("%s, Client.Jobs[0].ID = %s, but j.ID = %s", how, jobs[0].ID, j.ID)
	}

	jobs, err = srv.Jobs(p)
	if err != nil {
		t.Fatalf("%s, Server.Jobs: %v", how, err)
	}
	if len(jobs) != 1 {
		t.Fatalf("%s, Server.Jobs() returned %d jobs, want 1", how, len(jobs))
	}
	if jobs[0].ID != j.ID {
		t.Fatalf("%s, Server.Jobs[0].ID = %s, but j.ID = %s", how, jobs[0].ID, j.ID)
	}

	err = client.DeleteJob(j)
	if err != nil {
		t.Fatalf("%s, Client.DeleteJob: %v", how, err)
	}
	time.Sleep(2 * time.Second)

	xc, err := newXMPPClient(srv.Auth())
	if err != nil {
		t.Fatalf("%s, xc.init: %v", how, err)
	}

	info.Title += " II"
	j, err = client.Print(p1, info, data)
	if err != nil {
		t.Fatalf("%s, Client.Print 2: %v", how, err)
	}

	if err := xc.Recv(); err != nil {
		t.Fatalf("%s, xc.Recv: %v", how, err)
	}

	jobs, err = srv.Jobs(p)
	if err != nil {
		t.Fatalf("%s, Server.Jobs: %v", how, err)
	}
	if len(jobs) != 1 {
		t.Fatalf("%s, Server.Jobs() returned %d jobs, want 1", how, len(jobs))
	}
	if jobs[0].ID != j.ID {
		t.Fatalf("%s, Server.Jobs[0].ID = %s, but j.ID = %s", how, jobs[0].ID, j.ID)
	}

	data1, err := srv.ReadFile(jobs[0])
	if err != nil {
		t.Fatalf("%s, Server.ReadFile: %v", how, err)
	}

	if !bytes.Equal(data, data1) {
		ioutil.WriteFile("/tmp/wrong.pdf", data1, 0666)
		t.Fatalf("%s, did not get same PDF back %d vs %d", how, len(data), len(data1))
	}

	err = srv.UpdateJob(jobs[0], JobError, 1, "error!")
	if err != nil {
		t.Fatalf("%s, Server.UpdateJob error: %v", how, err)
	}

	err = srv.UpdateJob(jobs[0], JobDone, 0, "")
	if err != nil {
		t.Fatalf("%s, Server.UpdateJob done: %v", how, err)
	}

	jobs, err = client.Jobs(p1)
	if err != nil {
		t.Fatalf("%s, Client.Jobs: %v", how, err)
	}
	if len(jobs) != 1 {
		t.Fatalf("%s, after server.UpdateJob done, client jobs=%d", how, len(jobs))
	}
	if jobs[0].Status != "DONE" {
		t.Fatalf("%s, after server.UpdateJob status = %s, want DONE", how, jobs[0].Status)
	}

	if err := srv.DeletePrinter(p); err != nil {
		t.Fatalf("DeletePrinter: %v", err)
	}
}

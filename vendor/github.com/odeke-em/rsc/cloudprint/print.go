// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package cloudprint implements sending and receiving print jobs
// using Google Cloud Print (https://developers.google.com/cloud-print/).
package cloudprint

import (
	"fmt"
	"strings"
	"time"

	"code.google.com/p/goauth2/oauth"
)

// A Printer represents a Google Cloud Print printer.
// The fields provide a snapshot of the printer metadata.
type Printer struct {
	ID                 string
	Proxy              string
	Name               string
	DisplayName        string
	DefaultDisplayName string
	Description        string
	OwnerID            string
	CreateTime         time.Time
	AccessTime         time.Time
	UpdateTime         time.Time
	Status             string
	CapsFormat         string
	CapsHash           string
	Tags               []string
	GCPVersion         string
	IsTOSAccepted      bool
	Type               string
}

// A printerDesc is the JSON description of a printer as supplied by Google Cloud Print.
type printerDesc struct {
	ID                 string
	Proxy              string
	Name               string
	DisplayName        string
	DefaultDisplayName string
	Description        string
	OwnerID            string
	CreateTime         int64 `json:",string"`
	AccessTime         int64 `json:",string"`
	UpdateTime         int64 `json:",string"`
	Status             string
	CapsFormat         string
	CapsHash           string
	Tags               []string
	GCPVersion         string
	IsTOSAccepted      bool
	Type               string
}

func (d *printerDesc) Printer() *Printer {
	return &Printer{
		ID:                 d.ID,
		Proxy:              d.Proxy,
		Name:               d.Name,
		DisplayName:        d.DisplayName,
		DefaultDisplayName: d.DefaultDisplayName,
		Description:        d.Description,
		OwnerID:            d.OwnerID,
		CreateTime:         time.Unix(d.CreateTime, 0),
		AccessTime:         time.Unix(d.AccessTime, 0),
		UpdateTime:         time.Unix(d.UpdateTime, 0),
		Status:             d.Status,
		CapsFormat:         d.CapsFormat,
		CapsHash:           d.CapsHash,
		Tags:               d.Tags,
		GCPVersion:         d.GCPVersion,
		IsTOSAccepted:      d.IsTOSAccepted,
		Type:               d.Type,
	}
}

// A Job represents a document sent to be printed.
// The fields provide a snapshot of the job metadata.
type Job struct {
	ID          string
	PrinterID   string
	PrinterName string
	OwnerID     string
	Title       string
	Pages       int64
	CreateTime  time.Time
	UpdateTime  time.Time
	Status      string
	FileURL     string
	TicketURL   string
	PrinterType string
	ContentType string
	ErrorCode   string
	Tags        []string
}

// A jobDesc is the JSON description of a print job as supplied by Google Cloud Print.
type jobDesc struct {
	ID            string
	PrinterID     string
	PrinterName   string
	OwnerID       string
	Title         string
	NumberOfPages int64
	CreateTime    int64 `json:",string"`
	UpdateTime    int64 `json:",string"`
	Status        string
	FileURL       string
	TicketURL     string
	PrinterType   string
	ContentType   string
	ErrorCode     string
	Tags          []string
}

func (d *jobDesc) Job() *Job {
	return &Job{
		ID:          d.ID,
		PrinterID:   d.PrinterID,
		PrinterName: d.PrinterName,
		OwnerID:     d.OwnerID,
		Title:       d.Title,
		Pages:       d.NumberOfPages,
		CreateTime:  time.Unix(d.CreateTime, 0),
		UpdateTime:  time.Unix(d.UpdateTime, 0),
		Status:      d.Status,
		FileURL:     d.FileURL,
		TicketURL:   d.TicketURL,
		PrinterType: d.PrinterType,
		ContentType: d.ContentType,
		ErrorCode:   d.ErrorCode,
		Tags:        d.Tags,
	}
}

// An Auth is an authentication token that can be used to act as a
// print server or print client.
type Auth struct {
	// ClientID and ClientSecret identify the client using this code.
	// They are obtained from the Google APIs Console
	// (https://code.google.com/apis/console).
	// These fields are always required.
	APIClientID     string
	APIClientSecret string

	// ProxyID identifies a particular server, which might serve multiple printers.
	// This field is required for servers only.
	// A ProxyID can be generated using RandomID.
	ProxyID string

	// Token is an OAuth 2 token giving permission to manage or
	// print to a Google account's printers.
	// TokenUser is the email address of the corresponding email address.
	// XMPPJID is the XMPP Jabber ID to use when polling for new print jobs.
	// It overrides TokenUser and is set only when using Auths generated
	// by CreateOpenPrinter.
	Token     oauth.Token
	TokenUser string
	XMPPJID   string
}

// An OpenPrinter is a printer created without an authenticated Google account
// and therefore without an owner. A prospective owner claims the printer on the
// web using the claim URL, after which the printer implementation can obtain
// a credential corresponding to a one-printer server.
type OpenPrinter struct {
	auth     Auth
	register *registerResponse
	verify   *verifyResponse
	printer  *Printer
}

// A registerResponse is the JSON response when registering a printer.
type registerResponse struct {
	PollingURL        string `json:"polling_url"`
	InvitePageURL     string `json:"invite_page_url"`
	CompleteInviteURL string `json:"complete_invite_url"`
	TokenDuration     int64  `json:"token_duration,string"`
	RegistrationToken string `json:"registration_token"`
	InviteURL         string `json:"invite_url"`
	Printers          []printerDesc
}

// A verifyResponse is the JSON response when verifying a printer claim.
type verifyResponse struct {
	XMPPJID             string `json:"xmpp_jid"`
	ConfirmationPageURL string `json:"confirmation_page_url"`
	UserEmail           string `json:"user_email"`
	AuthorizationCode   string `json:"authorization_code"`
}

// CreateOpenPrinter registers a new open printer with Google.
//
// Only the ClientID, ClientSecret, and ProxyID fields need to be set in auth.
// The Token and TokenUser fields in the auth are ignored.
func CreateOpenPrinter(auth Auth, info *PrinterInfo) (*OpenPrinter, error) {
	var mr multipartRequest
	mr.init()
	mr.WriteField("printer", info.Name)
	w, _ := mr.CreateFormFile("capabilities", "capabilities")
	cap := info.Capabilities
	if cap == nil {
		cap = []byte(defaultPPD)
	}
	w.Write(cap)

	var resp registerResponse
	if err := jsonRPC(&auth, "POST", cloudprintURL+"/register", &mr, &resp); err != nil {
		return nil, fmt.Errorf("CreateOpenPrinter: %v", err)
	}
	if len(resp.Printers) != 1 {
		return nil, fmt.Errorf("CreateOpenPrinter: server did not create printer")
	}

	p := &OpenPrinter{
		auth:     auth,
		register: &resp,
		printer:  resp.Printers[0].Printer(),
	}
	return p, nil
}

// ClaimPDF returns the content of a PDF of instructions
// that can be printed and given to a prospective owner to claim the printer.
func (p *OpenPrinter) ClaimPDF() ([]byte, error) {
	data, err := httpGET(&p.auth, p.register.InvitePageURL)
	if err != nil {
		return nil, fmt.Errorf("ClaimPDF: %v", err)
	}
	return data, nil
}

// ClaimURL returns a URL that a prospective owner can visit
// to claim the printer.
func (p *OpenPrinter) ClaimURL() string {
	// The current implementation returns http://goo.gl/ URLs.
	// Return https://goo.gl/ URLs instead.
	url := p.register.CompleteInviteURL
	if strings.HasPrefix(url, "http://goo.gl") {
		url = "https" + url[4:]
	}
	return url
}

var ErrUnclaimed = fmt.Errorf("printer not yet claimed")

// VerifyClaim checks that the printer has been claimed.
// If the printer is claimed, VerifyClaim returns no error.
// If the printer is unclaimed, VerifyClaim retruns ErrUnclaimed.
// It is possible for VerifyClaim to return other errors, such as
// in the case of network problems.
//
// A side effect of verifying that claim is that Google creates
// a synthetic account that is only useful in a future call to
// NewServer, to manage just this one printer.
// The information about that account can be retrieved
// from the Auth, Printer, and Server methods after VerifyClaim
// succeeds.
func (p *OpenPrinter) VerifyClaim() error {
	if p.verify != nil {
		return nil
	}

	var resp verifyResponse
	if err := jsonRPC(&p.auth, "GET", p.register.PollingURL+p.auth.APIClientID, nil, &resp); err != nil {
		return fmt.Errorf("VerifyClaim: %v", err)
	}

	var tr oauth.Transport
	tr.Config = &oauth.Config{
		ClientId:     p.auth.APIClientID,
		ClientSecret: p.auth.APIClientSecret,
		Scope:        "https://www.googleapis.com/auth/cloudprint https://www.googleapis.com/auth/googletalk",
		AuthURL:      "https://accounts.google.com/o/oauth2/auth",
		TokenURL:     "https://accounts.google.com/o/oauth2/token",
		RedirectURL:  "oob",
	}
	tok, err := tr.Exchange(resp.AuthorizationCode)
	if err != nil {
		return fmt.Errorf("VerifyClaim: oauth exchange: %v", err)
	}

	p.auth.Token = *tok
	p.auth.TokenUser = resp.UserEmail
	p.auth.XMPPJID = resp.XMPPJID
	p.verify = &resp
	return nil
}

// ConfirmationPDF returns the content of a PDF that can be printed
// to confirm to the owner that the printer has been claimed.
func (p *OpenPrinter) ConfirmationPDF() ([]byte, error) {
	if p.verify == nil {
		return nil, fmt.Errorf("ConfirmationPDF: VerifyClaim has not succeeded")
	}
	data, err := httpGET(&p.auth, p.verify.ConfirmationPageURL)
	if err != nil {
		return nil, fmt.Errorf("ConfirmationPDF: %v", err)
	}
	return data, nil
}

// Auth returns a credential that can be used to manage this printer
// in future invocations of the program.
//
// Auth must be called only after VerifyClaim has succeeded.
func (p *OpenPrinter) Auth() Auth {
	if p.verify == nil {
		panic(fmt.Errorf("Auth: VerifyClaim has not succeeded"))
	}
	return p.auth
}

// Printer returns information about the newly created printer.
// This method is only a convenience: the returned printer is the
// (only) one that would be returned by p.Server().Printers().
//
// Printer must be called only after VerifyClaim has succeeded.
func (p *OpenPrinter) Printer() *Printer {
	if p.verify == nil {
		panic(fmt.Errorf("Printer: VerifyClaim has not succeeded"))
	}
	return p.printer
}

// Server returns a server that can manage the newly created printer.
// This method is only a convenience: the returned server is the one
// that would be returned by NewServer(p.Auth()).
//
// Server must be called only after VerifyClaim has succeeded.
func (p *OpenPrinter) Server() *Server {
	if p.verify == nil {
		panic(fmt.Errorf("Server: VerifyClaim has not succeeded"))
	}
	srv := &Server{
		auth: p.auth,
	}
	return srv
}

// A Server can manage one or more Google Cloud Print printers.
type Server struct {
	auth Auth
}

// Auth returns an auth token that can be passed to NewServer
// to reconnect to the server.
func (s *Server) Auth() Auth {
	return s.auth
}

// NewServer returns a print server managing the printers owned
// by the user denoted by the auth token.
func NewServer(auth Auth) (*Server, error) {
	srv := &Server{
		auth: auth,
	}
	return srv, nil
}

// A PrinterInfo describes settings for creating a new printer.
type PrinterInfo struct {
	Name string

	// Capabilities describes the printer capabilties in PPD format.
	// If Capabilities is nil, a basic default will be assumed.
	Capabilities []byte
}

// CreatePrinter creates a new printer.
// The server must be associated with a real Google account, meaning that
// it must have been created using NewServer with an auth returned by UserAuth.
// (Servers created using the auth returned by creating an open printer
// are limited to that one printer.)
func (s *Server) CreatePrinter(info *PrinterInfo) (*Printer, error) {
	var mr multipartRequest
	mr.init()
	mr.WriteField("printer", info.Name)
	w, _ := mr.CreateFormFile("capabilities", "capabilities")
	cap := info.Capabilities
	if cap == nil {
		cap = []byte(defaultPPD)
	}
	w.Write(cap)

	var resp registerResponse
	if err := jsonRPC(&s.auth, "POST", cloudprintURL+"/register", &mr, &resp); err != nil {
		return nil, fmt.Errorf("CreatePrinter: %v", err)
	}
	if len(resp.Printers) != 1 {
		return nil, fmt.Errorf("CreatePrinter: server did not create printer")
	}

	return resp.Printers[0].Printer(), nil
}

// DeletePrinter deletes the given printer.
func (s *Server) DeletePrinter(p *Printer) error {
	var mr multipartRequest
	mr.init()
	mr.WriteField("printerid", p.ID)

	err := jsonRPC(&s.auth, "POST", cloudprintURL+"/delete", &mr, nil)
	if err != nil {
		return fmt.Errorf("DeletePrinter: %v", err)
	}
	return nil
}

// Printers returns a list of printers managed by this server.
func (s *Server) Printers() ([]*Printer, error) {
	var mr multipartRequest
	mr.init()

	var resp printersResponse
	err := jsonRPC(&s.auth, "POST", cloudprintURL+"/list", &mr, &resp)
	if err != nil {
		return nil, fmt.Errorf("Printers: %v", err)
	}

	var list []*Printer
	for _, desc := range resp.Printers {
		list = append(list, desc.Printer())
	}
	return list, nil
}

/*
// UpdatePrinter updates attributes associated with the printer p.
func (s *Server) UpdatePrinter(p *Printer, info *PrinterInfo) error {
}
*/

// Jobs returns a list of jobs waiting to be printed on the given printer.
func (s *Server) Jobs(p *Printer) ([]*Job, error) {
	var mr multipartRequest
	mr.init()
	mr.WriteField("printerid", p.ID)

	var resp jobsResponse
	err := jsonRPC(&s.auth, "POST", cloudprintURL+"/fetch", &mr, &resp)
	if err != nil {
		return nil, fmt.Errorf("Jobs: %v", err)
	}

	var list []*Job
	for _, desc := range resp.Jobs {
		list = append(list, desc.Job())
	}
	return list, nil
}

// UpdateJob updates information about the given job.
// The only field that is stored back to the server is j.Status.
func (s *Server) UpdateJob(j *Job, status JobStatus, code int, message string) error {
	var mr multipartRequest
	mr.init()
	mr.WriteField("jobid", j.ID)
	mr.WriteField("printerid", j.PrinterID)
	mr.WriteField("status", string(status))
	if status == JobError {
		mr.WriteField("code", fmt.Sprint(code))
		mr.WriteField("message", message)
	}

	err := jsonRPC(&s.auth, "POST", cloudprintURL+"/control", &mr, nil)
	if err != nil {
		return fmt.Errorf("UpdateJob: %v", err)
	}
	return nil
}

// ReadFile reads the raw data to be printed by the given job.
func (s *Server) ReadFile(j *Job) ([]byte, error) {
	data, err := httpGET(&s.auth, j.FileURL)
	if err != nil {
		return nil, fmt.Errorf("ReadFile: %v", err)
	}
	return data, nil
}

// A Client can send print jobs to Google Cloud Print and manage those jobs.
type Client struct {
	auth Auth
}

// NewClient returns a new client using the given auth information.
func NewClient(auth Auth) (*Client, error) {
	c := &Client{
		auth: auth,
	}
	return c, nil
}

// Search searches for printers the client can use.
// The query can be empty to return all printers.
// By default only printers that have been online recently are returned.
// To return all printers, pass all = true.
func (c *Client) Search(query string, all bool) ([]*Printer, error) {
	var mr multipartRequest
	mr.init()
	if query != "" {
		mr.WriteField("q", query)
	}
	if all {
		mr.WriteField("connection_status", "ALL")
	}
	var resp printersResponse
	err := jsonRPC(&c.auth, "POST", cloudprintURL+"/search", &mr, &resp)
	if err != nil {
		return nil, fmt.Errorf("Search: %v", err)
	}
	var list []*Printer
	for _, desc := range resp.Printers {
		list = append(list, desc.Printer())
	}
	return list, nil
}

// Printer searches for the printer with the given ID.
func (c *Client) Printer(id string) (*Printer, error) {
	var mr multipartRequest
	mr.init()
	mr.WriteField("printerid", id)
	mr.WriteField("printer_connection_Status", "true")
	var resp printersResponse
	err := jsonRPC(&c.auth, "POST", cloudprintURL+"/printer", &mr, &resp)
	if err != nil {
		return nil, fmt.Errorf("Printer: %v", err)
	}
	if len(resp.Printers) != 1 {
		return nil, fmt.Errorf("Printer: response contained %d printers", len(resp.Printers))
	}
	return resp.Printers[0].Printer(), nil
}

// A printersResponse is the JSON response to /search and /printer.
type printersResponse struct {
	Printers []printerDesc
}

// DeleteJob deletes the given print job.
func (c *Client) DeleteJob(j *Job) error {
	var mr multipartRequest
	mr.init()
	mr.WriteField("jobid", j.ID)
	mr.WriteField("printerid", j.PrinterID)
	err := jsonRPC(&c.auth, "POST", cloudprintURL+"/deletejob", &mr, nil)
	if err != nil {
		return fmt.Errorf("DeleteJob: %v", err)
	}
	return nil
}

// Jobs returns a list of jobs waiting to be printed.
// If p is not nil, the list is restricted to jobs sent to the given printer.
func (c *Client) Jobs(p *Printer) ([]*Job, error) {
	var mr multipartRequest
	mr.init()
	if p != nil {
		mr.WriteField("printerid", p.ID)
	}
	var resp jobsResponse
	err := jsonRPC(&c.auth, "POST", cloudprintURL+"/jobs", &mr, &resp)
	if err != nil {
		return nil, fmt.Errorf("Jobs: %v", err)
	}
	var jobs []*Job
	for _, desc := range resp.Jobs {
		jobs = append(jobs, desc.Job())
	}
	return jobs, nil
}

// A jobsResponse is the JSON response to /jobs.
type jobsResponse struct {
	Jobs []jobDesc
}

// A JobInfo describes settings for creating a new print job.
type JobInfo struct {
	Title        string
	Tags         []string
	Capabilities []byte
}

// Print creates a new print job printing to p with the given job information.
// The data is the raw PDF to print.
func (c *Client) Print(p *Printer, info *JobInfo, data []byte) (*Job, error) {
	var mr multipartRequest
	mr.init()
	mr.WriteField("printerid", p.ID)
	mr.WriteField("title", info.Title)
	mr.WriteField("contentType", "application/pdf")
	for _, tag := range info.Tags {
		mr.WriteField("tag", tag)
	}
	cap := info.Capabilities
	if cap == nil {
		cap = []byte(`{"capabilities":[]}`)
	}
	w, _ := mr.CreateFormFile("capabilities", "capabilities")
	w.Write(cap)
	w, _ = mr.CreateFormFile("content", "x.pdf")
	w.Write(data)

	var resp submitResponse
	err := jsonRPC(&c.auth, "POST", cloudprintURL+"/submit", &mr, &resp)
	if err != nil {
		return nil, fmt.Errorf("Print: %v", err)
	}

	return resp.Job.Job(), nil
}

// A submitResponse is the JSON response to /submit.
type submitResponse struct {
	Job jobDesc
}

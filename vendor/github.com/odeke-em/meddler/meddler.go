package meddler

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
)

type DownloadItem struct {
	URI       string `form:"uri" binding:"required"`
	PublicKey string `form:"pubkey" binding:"-"`
	Signature string `form:"signature" binding:"-"`
}

type UriInsert struct {
	UriList []string
	Source  string
}

type Payload struct {
	URI         string `form:"uri" binding:"required"`
	PublicKey   string `form:"pubkey" binding:"-"`
	Signature   string `form:"signature" binding:"-"`
	Payload     string `form:"payload" binding:"-"`
	RequestTime int64  `form:"requesttime" binding:"required"`
	ExpiryTime  int64  `form:"expirytime" binding:"required"`
}

func (pl *Payload) RawTextForSigning() string {
	return fmt.Sprintf("%q%q%q%q%q", pl.URI, pl.RequestTime, pl.Payload, pl.PublicKey, pl.ExpiryTime)
}

func (pl *Payload) ToUrlValues(extras ...map[string]interface{}) url.Values {
	uv := url.Values{}
	uv.Set("payload", pl.Payload)
	uv.Set("pubkey", pl.PublicKey)
	uv.Set("signature", pl.Signature)
	uv.Set("uri", pl.URI)
	uv.Set("requesttime", fmt.Sprintf("%v", pl.RequestTime))
	uv.Set("expirytime", fmt.Sprintf("%v", pl.ExpiryTime))

	return uv
}

func headerShallowCopy(from, to http.Header) {
	for k, v := range from {
		to.Set(k, strings.Join(v, ","))
	}
}

func HeadGet(di DownloadItem, res http.ResponseWriter, req *http.Request) error {
	uri := di.URI
	headResponse, err := http.Head(uri)

	if err != nil {
		return err
	}

	dlHeader := headResponse.Header
	headerShallowCopy(dlHeader, res.Header())

	return nil
}

func Download(di DownloadItem, res http.ResponseWriter, req *http.Request) {
	uri := di.URI

	downloadResult, err := http.Get(uri)

	if err != nil {
		fmt.Fprintf(res, "%v", err)
		return
	}

	if downloadResult == nil || downloadResult.Body == nil {
		fmt.Fprintf(res, "could not get %q", uri)
		return
	}

	body := downloadResult.Body
	dlHeader := downloadResult.Header

	if downloadResult.Close {
		defer body.Close()
	}

	headerShallowCopy(dlHeader, res.Header())

	basename := filepath.Base(uri)
	extraHeaders := map[string][]string{
		"Content-Disposition": []string{
			fmt.Sprintf("attachment;filename=%q", basename),
		},
	}

	headerShallowCopy(extraHeaders, res.Header())

	res.WriteHeader(200)
	io.Copy(res, body)
}

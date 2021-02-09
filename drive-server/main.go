// Copyright 2015 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	qrcode "github.com/skip2/go-qrcode"

	"github.com/odeke-em/extractor"
	"github.com/odeke-em/meddler"
)

const (
	ENV_DRIVE_SERVER_PUB_KEY  = "DRIVE_SERVER_PUB_KEY"
	ENV_DRIVE_SERVER_PRIV_KEY = "DRIVE_SERVER_PRIV_KEY"
	ENV_DRIVE_SERVER_PORT     = "DRIVE_SERVER_PORT"
	ENV_DRIVE_SERVER_HOST     = "DRIVE_SERVER_HOST"
)

var envKeyAlias = &extractor.EnvKey{
	PubKeyAlias:  ENV_DRIVE_SERVER_PUB_KEY,
	PrivKeyAlias: ENV_DRIVE_SERVER_PRIV_KEY,
}

type addressInfo struct {
	port, host string
}

func envGet(varname string, placeholders ...string) string {
	v := os.Getenv(varname)
	if v == "" {
		for _, placeholder := range placeholders {
			if placeholder != "" {
				v = placeholder
				break
			}
		}
	}

	return v
}

func addressInfoFromEnv() *addressInfo {
	return &addressInfo{
		port: envGet(ENV_DRIVE_SERVER_PORT, "8010"),
		host: envGet(ENV_DRIVE_SERVER_HOST, "localhost"),
	}
}

var envKeySet = extractor.KeySetFromEnv(envKeyAlias)
var envAddrInfo = addressInfoFromEnv()

func (ai *addressInfo) ConnectionString() string {
	// TODO: ensure fields meet rubric
	return fmt.Sprintf("%s:%s", ai.host, ai.port)
}

type fileWatchInfo map[string]interface{}

func main() {
	if envKeySet.PublicKey == "" {
		errorPrint("publicKey not set. Please set %s in your env.\n", envKeyAlias.PubKeyAlias)
		return
	}

	if envKeySet.PrivateKey == "" {
		errorPrint("privateKey not set. Please set %s in your env.\n", envKeyAlias.PrivKeyAlias)
		return
	}

	m := martini.Classic()

	m.Get("/qr", binding.Bind(meddler.Payload{}), presentQRCode)
	m.Post("/watch-files", binding.Bind(fileWatchInfo{}), handleFileChanges)

	m.RunOnAddr(envAddrInfo.ConnectionString())
}

func handleFileChanges(fwi fileWatchInfo, res http.ResponseWriter, req *http.Request) {
	log.Printf("fwi: %#v", fwi)
}

func presentQRCode(pl meddler.Payload, res http.ResponseWriter, req *http.Request) {
	if pl.PublicKey != envKeySet.PublicKey {
		http.Error(res, "invalid publickey", 405)
		return
	}

	rawTextForSigning := pl.RawTextForSigning()
	if !envKeySet.Match([]byte(rawTextForSigning), []byte(pl.Signature)) {
		http.Error(res, "invalid signature", 403)
		return
	}

	curTimeUnix := time.Now().Unix()
	if pl.ExpiryTime < curTimeUnix {
		http.Error(res, fmt.Sprintf("request expired at %q, current time %q", pl.ExpiryTime, curTimeUnix), 403)
		return
	}

	uri := pl.URI
	pngImage, err := qrcode.Encode(uri, 256)
	if err != nil {
		fmt.Fprintf(res, "%s %v\n", uri, err)
		return
	}
	res.Write(pngImage)
}

func errorPrint(fmt_ string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "\033[31m")
	fmt.Fprintf(os.Stderr, fmt_, args...)
	fmt.Fprintf(os.Stderr, "\033[00m")
}

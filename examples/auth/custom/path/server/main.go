// MIT License
//
// Copyright (c) 2022 Kubeshop
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"time"

	"github.com/davecgh/go-spew/spew"
)

var (
	Username     = "kusk"
	Password     = "kusk"
	Port         = 8080
	PatternLogin = "/custom-path"
	PatternRoot  = "/"
)

type logWriter struct {
}

func (writer logWriter) Write(bytes []byte) (int, error) {
	now := time.Now().UTC()
	return fmt.Print(now.Format(time.RFC3339Nano) + " auth-custom-path: " + string(bytes))
}

func main() {
	log.SetFlags(0)
	log.SetOutput(&logWriter{})

	http.Handle(PatternLogin, http.HandlerFunc(handler))
	http.Handle(PatternRoot, http.HandlerFunc(handler))
	// http.Handle(PatternRoot, http.HandlerFunc(handlerRoot))

	address := fmt.Sprintf(":%d", Port)
	log.Printf("listening on %v at %v\n", address, PatternLogin)

	if err := http.ListenAndServe(address, nil); err != nil {
		log.Printf("http.ListenAndServe returned err=%v\n", err)
		os.Exit(1)
	}
}

func handlerRoot(w http.ResponseWriter, r *http.Request) {
	requestDumpBytes, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Printf("%v - handling request - failed to dump - err: %v\n", PatternRoot, err)
		log.Printf("%v - handling request - %+#v\n", PatternRoot, spew.Sprint(r))
	} else {
		log.Printf("%v - handling request - %+#v\n", PatternRoot, spew.Sprint(r))
		log.Printf("%v - handling request - \n%v\n", PatternRoot, string(requestDumpBytes))
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	requestDumpBytes, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Printf("handling request - failed to dump - err: %v\n", err)
		log.Printf("handling request - %+#v\n", spew.Sprint(r))
	} else {
		log.Printf("handling request - %+#v\n", spew.Sprint(r))
		log.Printf("handling request - \n%v\n", string(requestDumpBytes))
	}

	u, p, ok := r.BasicAuth()

	if !ok {
		w.WriteHeader(401)
		w.Write([]byte("expecting `Authorization: Basic a3VzazprdXNr` header\n"))

		return
	}

	if u != Username && p != Password {
		w.WriteHeader(401)
		w.Write([]byte(fmt.Sprintf("incorrect username provided: actual=%s, expected=%s\n", u, Username)))
		w.Write([]byte(fmt.Sprintf("incorrect password provided: actual=%s, expected=%s\n", p, Password)))

		return
	}

	if u != Username {
		w.WriteHeader(401)
		w.Write([]byte(fmt.Sprintf("incorrect username provided: actual=%s, expected=%s\n", u, Username)))

		return
	}

	if p != Password {
		w.WriteHeader(401)
		w.Write([]byte(fmt.Sprintf("incorrect password provided: actual=%s, expected=%s\n", p, Password)))

		return
	}

	w.WriteHeader(200)
	w.Header().Add("x-current-user", Username)

	return
}

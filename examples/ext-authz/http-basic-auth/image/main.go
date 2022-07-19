package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	Username = "kusk"
	Password = "kusk"
	Port     = 8080
)

type logWriter struct {
}

func (writer logWriter) Write(bytes []byte) (int, error) {
	now := time.Now().UTC()
	return fmt.Print(now.Format(time.RFC3339Nano) + " ext-authz-http-basic-auth: " + string(bytes))
}

func main() {
	log.SetFlags(0)
	log.SetOutput(&logWriter{})

	http.Handle("/", http.HandlerFunc(handler))

	address := fmt.Sprintf(":%d", Port)
	log.Printf("listening on %v\n", address)

	if err := http.ListenAndServe(address, nil); err != nil {
		log.Printf("http.ListenAndServe returne err=%v\n", err)
		os.Exit(1)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("handling request %+#v\n", r)

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

package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		bytes, err := httputil.DumpRequest(r, true)
		fmt.Println(string(bytes), err)
		w.Write([]byte(os.Getenv("RESPONSE")))
	})

	panic(http.ListenAndServe(":8080", nil))
}

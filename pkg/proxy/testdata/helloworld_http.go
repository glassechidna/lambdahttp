package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		bytes, err := httputil.DumpRequest(r, true)
		fmt.Println(string(bytes), err)
		w.Write([]byte(`pong`))
	})

	panic(http.ListenAndServe(":8080", nil))
}

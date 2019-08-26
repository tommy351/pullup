package main

import (
	"net/http"
	"os"
)

func main() {
	err := http.ListenAndServe(":80", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Resource-Name", os.Getenv("RESOURCE_NAME"))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))

	if err != nil {
		panic(err)
	}
}

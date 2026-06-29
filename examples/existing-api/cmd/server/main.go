package main

import (
	"log"
	"net/http"

	"example.com/existing-api/internal/httpapi"
)

func main() {
	mux := http.NewServeMux()
	httpapi.Register(mux)
	log.Fatal(http.ListenAndServe(":8080", mux))
}

package main

import (
	"log"
	"net/http"

	"github.com/KelvinHo995/crypto-strategy-lab/backend/internal/httpx"
)

func main() {
	router := httpx.NewRouter()
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

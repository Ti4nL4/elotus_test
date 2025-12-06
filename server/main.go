package main

import (
	"log"
	"net/http"
)

func main() {
	// TODO: Setup routes and middleware

	log.Println("Server starting on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}


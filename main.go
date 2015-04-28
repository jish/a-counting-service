package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		port = "3000"
	}

	fmt.Printf("Starting server on port %s\n", port)
	http.HandleFunc("/", rootRoute)
	http.ListenAndServe(":"+port, nil)
}

func rootRoute(response http.ResponseWriter, request *http.Request) {
	io.WriteString(response, "<h1>Hello Human, Welcome to A Counting Service</h1>\n")
}

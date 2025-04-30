package main

import (
	"fmt"
	"net/http"
)

const PORT_STR = ":8080"

func main() {
	fmt.Println("hello world")

	s := http.NewServeMux()
	server := http.Server{
		Addr:    ":8080",
		Handler: s,
	}

	fmt.Println("Server starting on", "http://localhost"+PORT_STR)
	err := server.ListenAndServe()
	if err != nil {
		fmt.Println(err)
	}

}

package main

import (
	"os"

	"github.com/mikhail-angelov/websignal/server"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9001"
	}
	s := &server.Server{
		Port: port,
	}
	s.Execute()
}

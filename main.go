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
	jwtSectret := os.Getenv("SECRET")
	if port == "" {
		jwtSectret = "tsjwt"
	}
	s := &server.Server{
		Port: port,
	}
	s.Run(jwtSectret)
}

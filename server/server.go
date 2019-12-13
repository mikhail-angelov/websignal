package server

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/mikhail-angelov/websignal/auth"
	"github.com/mikhail-angelov/websignal/logger"
)

// Server is http server
type Server struct {
	Port string
}

func test(w http.ResponseWriter, r *http.Request) {
	render.Status(r, http.StatusOK)
	render.PlainText(w, r, "test"+time.Now().String())
}

func (s *Server) composeRouter(jwtSectret string) *chi.Mux {
	var (
		url             = "http://localhost:9001"
		logger          = logger.New()
		auth            = auth.NewAuth(jwtSectret, logger, url)
		rooms           = NewRoomService()
		ws              = NewWsServer(rooms, auth, logger)
		roomsController = NewRoomsController(rooms)
		router          = chi.NewRouter()
	)
	auth.AddProvider("yandex", os.Getenv("YANDEX_OAUTH2_ID"), os.Getenv("YANDEX_OAUTH2_SECRET"))
	auth.AddProvider("github", os.Getenv("GITHUB_OAUTH2_ID"), os.Getenv("GITHUB_OAUTH2_SECRET"))
	auth.AddProvider("google", os.Getenv("GOOGLE_OAUTH2_ID"), os.Getenv("GOOGLE_OAUTH2_SECRET"))
	auth.AddProvider("local", "test", "test")
	AddFileServer(router, "/", http.Dir("./static"))
	router.HandleFunc("/ws", ws.SocketHandler)
	router.Mount("/auth", auth.Handlers())
	router.Route("/api", func(rapi chi.Router) {
		rapi.Group(func(r chi.Router) {
			r.Use(auth.Auth)
			r.Route("/room", roomsController.HTTPHandler)
		})
	})
	router.HandleFunc("/test", test)

	return router
}

// Run the HTTP server
func (s *Server) Run(jwtSectret string) error {
	var (
		serve  = make(chan error, 1)
		sig    = make(chan os.Signal, 1)
		router = s.composeRouter(jwtSectret)
	)

	port := s.Port
	log.Printf("listen %s port", port)
	err := http.ListenAndServe(":"+port, router)
	if err != nil {
		log.Fatalf("listen %s error: %v", port, err)
		return err
	}
	signal.Notify(sig, syscall.SIGTERM)

	select {
	case err := <-serve:
		log.Fatal(err)
	case sig := <-sig:
		log.Printf("signal %q received", sig)
	}
	log.Printf("[INFO] signaling server is terminated with error %+v", err)
	return err
}

package main

import (
	"log"
	"net"
	"net/http"
	"os"

	"github.com/SsSJKK/http/cmd/app"
	"github.com/SsSJKK/http/pkg/banners"
)

func main() {
	host := "0.0.0.0"
	port := "9999"
	log.Println("main")
	if err := execute(host, port); err != nil {
		os.Exit(1)
	}
}

func execute(host string, port string) (err error) {
	mux := http.NewServeMux()
	bannersSvc := banners.NewService()
	server := app.NewServer(mux, bannersSvc)
	srv := &http.Server{
		Addr:    net.JoinHostPort(host, port),
		Handler: server,
	}
	server.Init()
	log.Println("execute")
	return srv.ListenAndServe()
}

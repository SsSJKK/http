package main

import (
	"log"
	"net"
	"os"

	"github.com/SsSJKK/http/pkg/server"
)

func main() {
	host := "0.0.0.0"
	port := "9999"

	if err := execute(host, port); err != nil {
		os.Exit(1)
	}

}

func execute(host string, port string) (err error) {
	srv := server.NewServer(net.JoinHostPort(host, port))
	srv.Register("/payments", func(req *server.Request) {
		id := req.QueryParams["id"]
		log.Print(id)
	})
	return srv.Start()
}

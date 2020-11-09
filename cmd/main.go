package main

import (
	"log"
	"net"
	"os"
	"github.com/SsSJKK/http/pkg/server"
)

func main() {
	host := "0.0.0.0"
	port := "9998"

	if err := execute(host, port); err != nil {
		os.Exit(1)
	}

}

func execute(host string, port string) (err error) {
	srv := server.NewServer(net.JoinHostPort(host, port))
	srv.Register("/payments/{id}", func(req *server.Request) {
		id := req.PathParams["id"]
		log.Println(id)
	})
	srv.Register("/category{category}/{id}", func(req *server.Request) {
		catID := req.PathParams["category"]
		pID := req.PathParams["id"]
		log.Println(catID, pID)
	})
	return srv.Start()
}

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
	srv.Register("/payments/{id}", func(req *server.Request) {
		id := req.PathParams["id"]
		log.Print(id)
	})
	srv.Register("/category{catId}/product/{pId}", func(req *server.Request) {
		catID := req.PathParams["catId"]
		pID := req.PathParams["pId"]
		log.Println(catID, pID)
	})
	return srv.Start()
}

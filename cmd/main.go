package main

import (
	"fmt"
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
	srv.Register("/api/cards/1", func(req *server.Request) {
		fmt.Println("OK")
	})
	srv.Register("/c{catory}/{id}", func(req *server.Request) {
		fmt.Println("OK")
	})
	srv.Register("/payments/{id}", func(req *server.Request) {
		fmt.Println("OK")
	})
	return srv.Start()
}

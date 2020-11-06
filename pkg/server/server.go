package server

import (
	"bytes"
	"io"
	"log"
	"net"
	"strings"
	"sync"
)

// HandlerFunc ...
type HandlerFunc func(conn net.Conn)

// Server ...
type Server struct {
	addr     string
	mu       sync.Mutex
	handlers map[string]HandlerFunc
}

//NewServer ...
func NewServer(addr string) *Server {
	return &Server{addr: addr, handlers: make(map[string]HandlerFunc)}
}

//Register ...
func (s *Server) Register(path string, handler HandlerFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[path] = handler
}

//Start ...
func (s *Server) Start() error {

	listner, err := net.Listen("tcp", s.addr)
	if err != nil {
		log.Print(err, 1)
		return err
	}

	defer func() {
		if cerr := listner.Close(); cerr != nil {
			if err == nil {
				err = cerr
				return
			}
			log.Print(cerr)
		}
	}()

	for {
		conn, err := listner.Accept()
		if err != nil {
			log.Print(err)
			continue
		}

		go s.handle(conn)
		if err != nil {
			log.Print(err)
			continue
		}
	}
}

func (s *Server) handle(conn net.Conn) (err error) {
	mu := sync.RWMutex{}
	defer func() {
		if cerr := conn.Close(); cerr != nil {
			if err == nil {
				err = cerr
				return
			}
			log.Print(err)
		}
	}()
	//	conn.Write([]byte("Hello!\r\n"))

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err == io.EOF {
		log.Printf("%s", buf[:n])
	}
	if err != nil {
		return err
	}
	data := buf[:n]
	requestLineDelim := []byte{'\r', '\n'}
	requestLineEnd := bytes.Index(data, requestLineDelim)
	if requestLineEnd == -1 {
		return
	}
	requestLine := string(data[:requestLineEnd])
	parts := strings.Split(requestLine, " ")
	if len(parts) != 3 {
		return
	}
	path, version := parts[1], parts[2]

	if version != "HTTP/1.1" {
		return
	}
	mu.RLock()
	for k, v := range s.handlers {
		if k == path {
			v(conn)
		}
	}
	mu.RUnlock()

	return nil

}

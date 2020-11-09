package server

import (
	"bytes"
	"log"
	"net"
	"net/url"
	"strings"
	"sync"
)

// ErrBadRequest ...
const ErrBadRequest = "Error bad request !!!"

// Request ...
type Request struct {
	Conn        net.Conn
	QueryParams url.Values
	PathParams  map[string]string
}

// HandlerFunc ...
type HandlerFunc func(req *Request)

// Server ...
type Server struct {
	addr     string
	mu       sync.RWMutex
	handlers map[string]HandlerFunc
}

// NewServer ...
func NewServer(addr string) *Server {
	return &Server{addr: addr, handlers: make(map[string]HandlerFunc)}
}

// Register ...
func (s *Server) Register(path string, handler HandlerFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[path] = handler
}

// handle ...
func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, (1024 * 50))

	for {
		rbyte, err := conn.Read(buf)
		if err != nil {
			return
		}
		data := buf[:rbyte]
		ldelim := []byte{'\r', '\n'}
		index := bytes.Index(data, ldelim)
		if index == -1 {
			log.Println("chars not found")
			return
		}
		var req Request
		var good bool = true
		var path1 string = ""
		rline := string(data[:index])
		parts := strings.Split(rline, " ")
		req.PathParams = make(map[string]string)
		if len(parts) == 3 {
			path, version := parts[1], parts[2]
			decode, err := url.PathUnescape(path)
			if err != nil {
				log.Println(err)
				return
			}
			if version != "HTTP/1.1" {
				log.Println("version is not valid")
				return
			}
			url, err := url.ParseRequestURI(decode)
			if err != nil {
				log.Println(err)
				return
			}
			p := strings.Split(path, "/")
			if len(p) < 1 {
				return
			}
			if p[1] == "payments" {
				path1 = "/" + p[1] + "/{id}"
				if len(p) == 3 {
					req.PathParams["id"] = p[2]
				}
			}
			log.Println(good, path, url, p, path1)
			fn, ok := s.handlers[path1]
			if ok {
				fn(&req)
			} else {
				conn.Close()
				log.Print("conn.Close()")
			}
		}

	}
}

// Start ...
func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		log.Print(err)
		return err
	}
	defer func() {
		if cerr := listener.Close(); cerr != nil {
			if err == nil {
				err = cerr
				return
			}
			log.Print(cerr)
		}
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go s.handle(conn)
	}
}

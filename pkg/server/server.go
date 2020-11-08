package server

import (
	"bytes"
	"io"
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
	defer func() {
		if cerr := conn.Close(); cerr != nil {
			log.Println(cerr)
		}
	}()
	for {
		buf := make([]byte, 4096)
		n, err := conn.Read(buf)
		if err == io.EOF {
			log.Printf("%s", buf[:n])
		}
		if err != nil {
			log.Println(err)
			return
		}
		//log.Printf("%s", buf[:n])

		data := buf[:n]
		rnIndex := []byte{'\r', '\n'}
		reqEnd := bytes.Index(data, rnIndex)
		if reqEnd == -1 {
			log.Println(ErrBadRequest)
			return
		}
		reqLine := string(data[:reqEnd])
		parts := strings.Split(reqLine, " ")
		if len(parts) != 3 {
			log.Println(ErrBadRequest)
			return
		}
		path := parts[1]
		uri, err := url.ParseRequestURI(path)

		if err != nil {
			log.Println("error in decoding")
			return
		}
		p := strings.Split(uri.Path, "/")
		if len(p) == 3 {
			uri.RawQuery = "id=" + p[2]
			var pathParams = map[string]string{}
			pathParams["id"] = p[2]
			var req = Request{conn, uri.Query(), pathParams}
			s.mu.RLock()
			fn, ok := s.handlers["/"+p[1]+"/{id}"]
			s.mu.RUnlock()
			if ok {
				fn(&req)
			}
		}
		if len(p)==4{
			p2 := strings.Split(p[1],"category")
			var pathParams = map[string]string{}
			pathParams["catId"] = p2[1]
			pathParams["pId"] = p[3]
			var req = Request{conn, uri.Query(), pathParams}
			s.mu.RLock()
			fn, ok := s.handlers["/category{catId}/product/{pId}"]
			s.mu.RUnlock()
			if ok {
				fn(&req)
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

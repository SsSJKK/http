package server

import (
	"bytes"
	"log"
	"net"
	"net/url"
	"strconv"
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

	buf := make([]byte, 4096)

	for {
		rbyte, err := conn.Read(buf)
		if err != nil {
			return
		}
		data := buf[:rbyte]
		ldelim := []byte{'\r', '\n'}
		index := bytes.Index(data, ldelim)
		if index == -1 {
			log.Println("delim chars not found :(")
			return
		}
		var req Request
		var path1 string
		rline := string(data[:index])
		parts := strings.Split(rline, " ")
		req.PathParams = make(map[string]string)
		if len(parts) == 3 {
			_, path, version := parts[0], parts[1], parts[2]
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
			req.Conn = conn
			req.QueryParams = url.Query()
			/// PathParams...
			partsPath := strings.Split(url.Path, "/")
			for _, e := range partsPath {
				if e == "" {
					continue
				}
				_, err := strconv.Atoi(e)
				if err == nil {
					path1 += ("/{id}")
					req.PathParams["id"] = e
				} else {
					_, err := strconv.Atoi(string(e[len(e)-1]))
					if err == nil {
						var firstInt int = 0
						for i := 0; i < len(e); i++ {
							_, err := strconv.Atoi(string(e[i]))
							if err == nil {
								firstInt = i
								break
							}
						}
						path1 += ("/" + e[:firstInt] + "{" + e[:firstInt] + "Id}")
						req.PathParams[e[:firstInt]+"Id"] = e[firstInt:]
					} else {
						path1 += ("/" + e)
					}
				}
			}

		}

		var good bool = false
		var f = func(req *Request) {}

		s.mu.RLock()
		f, good = s.handlers[path1]
		s.mu.RUnlock()

		if good == false {
			conn.Close()
		} else {
			f(&req)
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

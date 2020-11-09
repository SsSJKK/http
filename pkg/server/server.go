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
	Headers     map[string]string
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
			pathSplit := strings.Split(path, "/")
			var p string
			var z string
			var zz string
			var header []byte = data[index+2:]
			req.Headers = make(map[string]string)
			//var hh = make(map[string]string)
			var pathParms = make(map[string]string)
			for _, pathPart := range pathSplit {
				b := true
				for i, x := range strings.Split(pathPart, "") {
					_, err := strconv.Atoi(x)
					if err == nil {
						if i == 0 {
							z = "id"
							zz = "{" + z + "}"
						} else {

							z = pathPart[:i]
							zz = "{" + z + "}"
						}
						p += "/" + pathPart[:i] + zz
						b = false
						pathParms[z] = pathPart[i:]
						break
					}
				}
				if b && pathPart != "" {
					p += "/" + pathPart
				}
			}
			req.PathParams = pathParms
			/// Headers .....
			if len(header) > 0 {
				ldelim := []byte{'\r', '\n', '\r', '\n'}
				index := bytes.Index(header, ldelim)
				if index == -1 {
					log.Println("index -1")
					return
				}
				data := string(header[:index])
				lheader := strings.Split(data, "\r\n")
				for _, header := range lheader {
					index := strings.Index(header, ":")
					if index == -1 {
						log.Println("index - 1")
						return
					}
					key, value := header[:index], header[index+2:]
					req.Headers[key] = value
				}
				log.Println("Headers: ")
				for k, v := range req.Headers {
					log.Println(k, v)
				}
			}
			s.mu.RLock()
			f, good := s.handlers[p]
			s.mu.RUnlock()

			if good == false {
				conn.Close()
			} else {
				f(&req)
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

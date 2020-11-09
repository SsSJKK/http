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
			log.Println("delim chars not found :(")
			return
		}
		var req Request
		var good bool = true
		var path1 string = ""
		rline := string(data[:index])
		parts := strings.Split(rline, " ")
		req.Headers = make(map[string]string)
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
			partsPath := strings.Split(url.Path, "/")
			for cur := range s.handlers {
				partsCur := strings.Split(cur, "/")
				if len(partsPath) != len(partsCur) {
					continue
				}
				var n int = len(partsPath)
				for i := 0; i < n && good == true; i++ {
					var l int = strings.Index(partsCur[i], "{")
					var r int = strings.LastIndex(partsCur[i], "}")
					var cnt int = strings.Count(partsCur[i], "{") +
						strings.Count(partsCur[i], "}")
					if cnt == 0 {
						if partsCur[i] != partsPath[i] {
							good = false
						}
					} else if cnt == 2 {
						req.PathParams[partsCur[i][l+1:r]] = partsPath[i][l:]
					} else {
						good = false
					}
				}
				if good == false {
					req.PathParams = make(map[string]string)
				} else {
					path1 = cur
					break
				}
			}
			log.Println("url.Path:", url.Path)
			log.Println("path1:", path1)
			log.Println("req.PathParams:", req.PathParams)
		}
		hLD := []byte{'\r', '\n', '\r', '\n'}
		hLE := bytes.Index(data, hLD)
		headersLine := string(data[index:hLE])
		headers := strings.Split(headersLine, "\r\n")[1:]
		mp := make(map[string]string)
		for _, v := range headers {
			headerLine := strings.Split(v, ": ")
			mp[headerLine[0]] = headerLine[1]
		}
		req.Headers = mp
		log.Println(req.Headers)
		log.Println(headersLine)
		s.mu.RLock()
		f, good := s.handlers[path1]
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

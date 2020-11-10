package server

import (
	"bytes"
	"fmt"
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
	Body        []byte
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
		rPath := ""
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
			getPath := path
			gPS := strings.Split(getPath, "/")
			for regPath := range s.handlers {
				rPS := strings.Split(regPath, "/")
				if len(rPS) == len(gPS) {
					for i, v := range rPS {
						if v == "" {
							continue
						}
						if v == gPS[i] {
							rPath += "/" + v
							_, err := strconv.Atoi(gPS[i])
							if err == nil {
								req.PathParams["id"] = gPS[i]
							}
							continue
						}
						a := strings.Index(v, "{")
						if a == -1 {
							break
						}
						if v[:a] != gPS[i][:a] {
							break
						}
						key := v[a:]
						val := gPS[i][a:]
						rPath += "/" + v
						req.PathParams[key[1:len(key)-1]] = val
					}
				}
			}
		}
		//HEADERS
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
		req.Body = data[hLE+4:]
		fmt.Println("QueryParams: ")
		for k, v := range req.QueryParams {
			fmt.Println(k, v)
		}
		fmt.Println()
		fmt.Println("PathParams: ")
		for k, v := range req.PathParams {
			fmt.Println(k, v)
		}
		fmt.Println()
		fmt.Println("Headers: ")
		for k, v := range req.Headers {
			fmt.Println(k, v)
		}
		fmt.Println()
		s.mu.RLock()
		f, good := s.handlers[rPath]
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

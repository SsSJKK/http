package app

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/SsSJKK/http/pkg/banners"
)

//Server ...
type Server struct {
	mux        *http.ServeMux
	bannersSvc *banners.Service
}

//NewServer ...
func NewServer(mux *http.ServeMux, bannersSvc *banners.Service) *Server {
	log.Println("NewServer")
	return &Server{mux: mux, bannersSvc: bannersSvc}
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	log.Println("ServeHTTP")
	s.mux.ServeHTTP(writer, request)
}

// Init all handle functions
func (s *Server) Init() {
	s.mux.HandleFunc("/banners.getAll", s.handleGetAllBanners)
	s.mux.HandleFunc("/banners.getById", s.handleGetBannerByID)
	s.mux.HandleFunc("/banners.save", s.handleSaveBanner)
	s.mux.HandleFunc("/banners.removeById", s.handleRemoveByID)
}

func (s *Server) handleGetAllBanners(writer http.ResponseWriter, request *http.Request) {
	items, err := s.bannersSvc.All(request.Context())
	data, err := json.Marshal(items)
	if err != nil {
		log.Println(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(data)
	if err != nil {
		log.Println(err)
	}

}
func (s *Server) handleGetBannerByID(writer http.ResponseWriter, request *http.Request) {
	idQuery := request.URL.Query().Get("id")
	id, err := strconv.ParseInt(idQuery, 10, 64)
	if err != nil {
		log.Println(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	item, err := s.bannersSvc.ByID(request.Context(), id)
	if err != nil {
		log.Println(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(item)
	if err != nil {
		log.Println(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(data)
	if err != nil {
		log.Println(err)
	}
}
func (s *Server) handleSaveBanner(writer http.ResponseWriter, request *http.Request) {

	idstr := request.PostFormValue("id")
	idInt, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		log.Println(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	banner, err := s.bannersSvc.Save(request.Context(),
		&banners.Banner{
			ID:      idInt,
			Title:   request.PostFormValue("title"),
			Content: request.PostFormValue("content"),
			Button:  request.PostFormValue("button"),
			Link:    request.PostFormValue("link"),
			Image:   request.PostFormValue("image"),
		}, request)
	if err != nil {
		log.Println(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(banner)
	if err != nil {
		log.Println(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(data)
	if err != nil {
		log.Println(err)
	}

}
func (s *Server) handleRemoveByID(writer http.ResponseWriter, request *http.Request) {
	idQuery := request.URL.Query().Get("id")
	id, err := strconv.ParseInt(idQuery, 10, 64)
	if err != nil {
		log.Println(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	item, err := s.bannersSvc.RemoveByID(request.Context(), id)
	if err != nil {
		log.Println(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(item)
	if err != nil {
		log.Println(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(data)
	if err != nil {
		log.Println(err)
	}
}

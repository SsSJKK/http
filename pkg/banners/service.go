package banners

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

//Service ...
type Service struct {
	mu    sync.RWMutex
	items []*Banner
}

// NewService ...
func NewService() *Service {
	return &Service{items: make([]*Banner, 0)}
}

var nextID int64 = 0

//Banner ...
type Banner struct {
	ID      int64
	Title   string
	Content string
	Button  string
	Link    string
	Image   string
}

//All ...
func (s *Service) All(ctx context.Context) ([]*Banner, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.items, nil
}

//ByID ...
func (s *Service) ByID(ctx context.Context, id int64) (*Banner, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, bnr := range s.items {
		if bnr.ID == id {
			return bnr, nil
		}
	}
	return nil, errors.New("item not found")
}

//Save ...
func (s *Service) Save(ctx context.Context, item *Banner, req *http.Request) (*Banner, error) {
	s.mu.RLock()
	fSend := true
	sF, sFH, err := req.FormFile("image")
	defer s.mu.RUnlock()
	if err != nil {
		fSend = false
	} else {
		defer sF.Close()
	}

	path := "./web/banners/"
	err = req.ParseMultipartForm(10 * 1024 * 1024)

	if item.ID == 0 {
		nextID++
		item.ID = nextID
		if fSend {
			fN := strconv.Itoa(int(nextID)) + sFH.Filename[strings.LastIndex(sFH.Filename, "."):]
			item.Image = fN
			f, err := os.OpenFile(path+fN, os.O_WRONLY|os.O_CREATE, 0666)
			if err != nil {
				return nil, err
			}
			defer f.Close()
			io.Copy(f, sF)
		}
		s.items = append(s.items, item)
		return item, nil
	}
	for i, bnr := range s.items {
		if bnr.ID == item.ID {

			if fSend {
				fN := strconv.Itoa(int(bnr.ID)) + sFH.Filename[strings.LastIndex(sFH.Filename, "."):]
				item.Image = fN
				f, err := os.OpenFile(path+fN, os.O_WRONLY|os.O_CREATE, 0666)
				if err != nil {
					return nil, err
				}

				defer f.Close()
				io.Copy(f, sF)
			} else {
				item.Image = bnr.Image
			}
			bnr = item
			s.items[i] = bnr
			return bnr, nil
		}
	}
	return nil, errors.New("item not found")
}

//RemoveByID ...
func (s *Service) RemoveByID(ctx context.Context, id int64) (*Banner, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i, bnr := range s.items {
		if bnr.ID == id {
			s.items = append(s.items[0:i], s.items[i+1:]...)
			return bnr, nil
		}
	}
	return nil, errors.New("item not found")
}

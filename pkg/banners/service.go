package banners

import (
	"context"
	"errors"
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
func (s *Service) Save(ctx context.Context, item *Banner) (*Banner, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if item.ID == 0 {
		nextID++
		item.ID = nextID
		s.items = append(s.items, item)
		return item, nil
	}
	for i, bnr := range s.items {
		if bnr.ID == item.ID {
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

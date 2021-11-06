package store

import (
	"errors"
	"log"
	"sync"
	"time"
)

type Location struct {
	Lat float64
	Lng float64
}

type LocationTTL struct {
	l []Location
	lastAccessed time.Time
}

type LocationData struct {
	m map[string]*LocationTTL
	l sync.Mutex
}

func NewLocationData(ttl int) *LocationData {
	ld := &LocationData{m: make(map[string]*LocationTTL, 0)}
	go func() {
		for now := range time.Tick(time.Second) {
			ld.l.Lock()
			for k, locations := range ld.m {
				// original non-working submission: now.Add(time.Duration(ttl) * time.Second).Before(locations.lastAccessed)
				if locations.lastAccessed.Before(now.Add(-time.Duration(ttl) * time.Second)) {
					delete(ld.m, k)
				}
			}
			ld.l.Unlock()
		}
	}()
	return ld
}

type Store interface {
	Append(orderId string, location Location) (bool, []Location, error)
	Get(orderId string, limit int) ([]Location, error)
	Delete(orderId string) error
}

type HistoryStore struct {
	data *LocationData
	l    sync.RWMutex
	log  *log.Logger
	ttl  int
}

func NewHistoryStore(log *log.Logger, ttl int) Store {
	return &HistoryStore{
		data: NewLocationData(ttl),
		log:  log,
		ttl:  ttl,
	}
}

func (s *HistoryStore) Append(orderId string, location Location) (bool, []Location, error) {
	s.l.Lock()
	defer s.l.Unlock()
	if s.data.m[orderId] == nil {
		s.data.m[orderId] = &LocationTTL{}
	}

	created := len(s.data.m[orderId].l) == 0

	s.data.m[orderId].lastAccessed = time.Now()
	s.data.m[orderId].l = append([]Location{location}, s.data.m[orderId].l...)

	return created, s.data.m[orderId].l, nil
}

func (s *HistoryStore) Get(orderId string, limit int) ([]Location, error) {
	s.l.RLock()
	defer s.l.RUnlock()

	ld, ok := s.data.m[orderId]
	if !ok {
		return nil, errors.New("order doesn't exist")
	}
	if limit > 0 && limit < len(ld.l) {
		return ld.l[:limit], nil
	}
	return ld.l, nil
}

func (s *HistoryStore) Delete(orderId string) error {
	s.l.Lock()
	defer s.l.Unlock()
	if _, ok := s.data.m[orderId]; !ok {
		return errors.New("invalid orderId")
	}
	delete(s.data.m, orderId)
	return nil
}
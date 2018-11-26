package item

import (
	"sync"
)

// LocalStore does the job of a key value store
type LocalStore struct {
	mux   sync.Mutex
	items map[string]Item
}

// NewLocalStore creates a new empty store
func NewLocalStore() *LocalStore {
	return &LocalStore{items: make(map[string]Item)}
}

// Read gets an item from the store, returning an empty Item if not present
func (s *LocalStore) Read(id string) (Item, error) {
	s.mux.Lock()
	defer s.mux.Unlock()
	return s.items[id], nil
}

// Write stores an item in the store
func (s *LocalStore) Write(item Item) error {
	if item.IsEmpty() {
		return NewEmptyItemError()
	}
	s.mux.Lock()
	defer s.mux.Unlock()
	s.items[item.ID] = item
	return nil
}

// Delete removes an item from the store if present
func (s *LocalStore) Delete(id string) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	delete(s.items, id)
	return nil
}

// Close the store
func (s *LocalStore) Close() error {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.items = make(map[string]Item)
	return nil
}

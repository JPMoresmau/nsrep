package item

import (
	"fmt"
)

// Item holds an item content
type Item struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Name     string                 `json:"name"`
	Contents map[string]interface{} `json:"contents"`
}

// IsEmpty returns true if the item is empty/not found
func (item *Item) IsEmpty() bool {
	return item.ID == ""
}

// StoreError represents a store error
type StoreError struct {
	code    string
	message string
}

func (e StoreError) Error() string {
	return fmt.Sprintf("%s: %s", e.code, e.message)
}

// NewEmptyItemError when a full item was expected
func NewEmptyItemError() error {
	return StoreError{"EMPTY_ITEM", "Empty item provided"}
}

// NewStoreCreationError when the store could not be created
func NewStoreCreationError(err error) error {
	return StoreError{"STORE_CREATION", err.Error()}
}

// NewStoreCloseError when the store could not be closed
func NewStoreCloseError(err error) error {
	return StoreError{"STORE_CLOSE", err.Error()}
}

// Store defines the interface to manipulate items
type Store interface {
	Read(id string) (Item, error)
	Write(item Item) error
	Delete(id string) (Item, error)
	Close() error
}

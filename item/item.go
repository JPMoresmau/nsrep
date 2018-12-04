package item

import (
	"fmt"
	"strings"

	"github.com/go-errors/errors"
)

// Item holds an item content
type Item struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Name     string                 `json:"name"`
	Contents map[string]interface{} `json:"contents"`
}

// ItemStatus is a item + a status
type ItemStatus struct {
	Item   Item
	Status string
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
	return errors.New(StoreError{"EMPTY_ITEM", "Empty item provided"})
}

// NewStoreCreationError when the store could not be created
func NewStoreCreationError(err error) error {
	return errors.New(StoreError{"STORE_CREATION", err.Error()})
}

// NewStoreCloseError when the store could not be closed
func NewStoreCloseError(err error) error {
	return errors.New(StoreError{"STORE_CLOSE", err.Error()})
}

// NewStoreInternalError when the store encountered some error
func NewStoreInternalError(err error) error {
	return errors.New(StoreError{"STORE_INTERNAL", err.Error()})
}

// NewStoreClosedError when the store is closed and we operate on it
func NewStoreClosedError() error {
	return errors.New(StoreError{"STORE_CLOSED", "Store is closed"})
}

// NewItemMarshallError when the item could not be marshalled properly into the store
func NewItemMarshallError(err error) error {
	return errors.New(StoreError{"ITEM_MARSHALL", err.Error()})
}

// NewMultipleItemErrors combines several error messages into one
func NewMultipleItemErrors(errs []string) error {
	if len(errs) == 0 {
		return nil
	}
	return errors.New(StoreError{"ITEM_MULTIPLE", strings.Join(errs, "\n")})
}

// NewItemUnmarshallError when the item could not be unmarshalled properly from the store
func NewItemUnmarshallError(err error) error {
	return errors.New(StoreError{"ITEM_UNMARSHALL", err.Error()})
}

// Store defines the interface to manipulate items
type Store interface {
	Read(id string) (Item, error)
	Write(item Item) error
	Delete(id string) error
	Close() error
}

// HistoryStore can provide history for a given item
type HistoryStore interface {
	History(id string, limit int) ([]ItemStatus, error)
}

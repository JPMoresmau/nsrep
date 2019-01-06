package item

import (
	"fmt"
	"strings"

	"github.com/go-errors/errors"
)

// ID is a list of string components
type ID = []string

// IDToString converts an ID into a string
func IDToString(id ID) string {
	return strings.Join(id, "/")
}

// StringToID converts a string into an ID
func StringToID(str string) ID {
	return strings.Split(str, "/")
}

// Item holds an item content
type Item struct {
	ID       ID                     `json:"id"`
	Type     string                 `json:"type"`
	Name     string                 `json:"name"`
	Contents map[string]interface{} `json:"contents"`
}

// IsEmpty returns true if the item is empty/not found
func (item *Item) IsEmpty() bool {
	return len(item.ID) == 0
}

// Flatten transforms an Item into a map of key/value pairs
func (item *Item) Flatten() map[string]interface{} {
	body := make(map[string]interface{})
	for k, v := range item.Contents {
		body[k] = v
	}
	body["item.name"] = item.Name
	body["item.type"] = item.Type
	body["item.id"] = item.ID
	return body
}

// Status is a item + a status
type Status struct {
	Item   Item
	Status string
}

// Facet is an enum of all the possible facets
type Facet int

// Different facets possible
const (
	FacetName      Facet = iota
	FacetType      Facet = iota
	FacetNamespace Facet = iota
)

// Query for searching
type Query struct {
	QueryString string
	From        int
	Length      int
	Facets      []Facet
}

// NewQuery builds a new query from the given string, returning the first 10 results
func NewQuery(queryString string) *Query {
	return &Query{queryString, 0, 10, make([]Facet, 0)}
}

// Page modifies the given query to add paging (from/length) information
func (q *Query) Page(from int, length int) *Query {
	q.From = from
	q.Length = length
	return q
}

// AddFacet add a facet to the query
func (q *Query) AddFacet(facet Facet) *Query {
	q.Facets = append(q.Facets, facet)
	return q
}

// AddAllFacets adds all possible facets
func (q *Query) AddAllFacets() *Query {
	q.Facets = []Facet{FacetName, FacetType, FacetNamespace}
	return q
}

// Score is a item + a search score
type Score struct {
	Item  Item    `json:"item"`
	Score float64 `json:"score"`
}

// SearchResult encapsulate the, ahem, search results
type SearchResult struct {
	Scores []Score                      `json:"scores"`
	Facets map[string]map[string]uint64 `json:"facets"`
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
	Read(id ID) (Item, error)
	Write(item Item) error
	Delete(id ID) error
	Close() error
}

// HistoryStore can provide history for a given item
type HistoryStore interface {
	History(id ID, limit int) ([]Status, error)
}

// SearchStore can provide full text search
type SearchStore interface {
	Search(query *Query) (SearchResult, error)
	Scroll(query string, scoreChannel chan Score, errorChannel chan error)
}

// DeleteTree deletes an item and all its children
func DeleteTree(id ID, stores []Store, searchStore SearchStore) error {

	errorC := make(chan error)
	go func() {
		defer close(errorC)
		deleteMultiple(id, stores, errorC)
		scoreC := make(chan Score)

		go searchStore.Scroll(fmt.Sprintf("item.id:%s/*", IDToString(id)), scoreC, errorC)

		for score := range scoreC {
			deleteMultiple(score.Item.ID, stores, errorC)
		}
	}()

	var errors []string
	for err := range errorC {
		errors = append(errors, err.Error())
	}
	return NewMultipleItemErrors(errors)
}

func deleteMultiple(id ID, stores []Store, errorChannel chan error) {
	for _, store := range stores {
		if store != nil {
			err := store.Delete(id)
			if err != nil {
				errorChannel <- err
			}
		}
	}
}

// AllNamespaces give the list of all namespaces parents
func AllNamespaces(id ID) []string {
	ns := make([]string, 0)
	for r := range id[:len(id)-1] {
		ns = append(ns, IDToString(id[:r+1]))
	}
	return ns
}

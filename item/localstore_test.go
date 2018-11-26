package item

import (
	"testing"
)

func TestLocalStore(t *testing.T) {
	store := NewLocalStore()
	DoTestStore(store, t)
}

func TestLocaleStoreErrors(t *testing.T) {
	store := NewLocalStore()
	DoTestStoreErrors(store, t)
}

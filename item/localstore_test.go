package item

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStore(t *testing.T) {
	require := require.New(t)
	store := NewLocalStore()
	item1 := Item{"123", "Table", "Table1", make(map[string]interface{})}

	item3, err := store.Read("123")
	require.Nil(err)
	require.Equal(Item{}, item3)

	err = store.Write(item1)
	require.Nil(err)
	item2, err := store.Read("123")
	require.Nil(err)
	require.Equal(item1, item2)

	item5, err := store.Delete(item2.ID)
	require.Nil(err)
	require.Equal(item1, item5)
	item4, err := store.Read("123")
	require.Nil(err)
	require.Equal(Item{}, item4)
}

func TestStoreErrors(t *testing.T) {
	require := require.New(t)
	store := NewLocalStore()
	err := store.Write(Item{})
	require.NotNil(err)
	require.True(strings.HasPrefix(err.Error(), "EMPTY_ITEM"))
}

package item

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJSON(t *testing.T) {
	require := require.New(t)
	item1 := Item{"123", "Table", "Table1", make(map[string]interface{})}
	b, err := json.Marshal(item1)
	require.Nil(err)
	require.Equal(`{"id":"123","type":"Table","name":"Table1","contents":{}}`, string(b))
	item2 := Item{}
	err = json.Unmarshal(b, &item2)
	require.Nil(err)
	require.Equal(item1, item2)
}

func DoTestStore(store Store, t *testing.T) {
	require := require.New(t)
	item1 := Item{"123", "Table", "Table1", make(map[string]interface{})}

	item3, err := store.Read("123")
	require.Nil(err)
	require.Equal(Item{}, item3)

	err = store.Write(item1)
	require.Nil(err)
	item2, err := store.Read("123")
	require.Nil(err)
	require.Equal(item1, item2)

	err = store.Delete(item2.ID)
	require.Nil(err)
	item4, err := store.Read("123")
	require.Nil(err)
	require.Equal(Item{}, item4)
}

func DoTestStoreErrors(store Store, t *testing.T) {
	require := require.New(t)
	err := store.Write(Item{})
	require.NotNil(err)
	require.True(strings.HasPrefix(err.Error(), "EMPTY_ITEM"))
}

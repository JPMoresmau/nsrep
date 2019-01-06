package item

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJSON(t *testing.T) {
	require := require.New(t)
	item1 := Item{[]string{"123"}, "Team", "Team1", make(map[string]interface{})}
	b, err := json.Marshal(item1)
	require.Nil(err)
	require.Equal(`{"id":["123"],"type":"Team","name":"Team1","contents":{}}`, string(b))
	item2 := Item{}
	err = json.Unmarshal(b, &item2)
	require.Nil(err)
	require.Equal(item1, item2)
}

func DoTestStore(store Store, t *testing.T) {
	require := require.New(t)
	item1 := Item{[]string{"123"}, "Team", "Team1", make(map[string]interface{})}

	item3, err := store.Read([]string{"123"})
	require.NoError(err)
	require.Equal(Item{}, item3)

	err = store.Write(item1)
	require.NoError(err)
	item2, err := store.Read([]string{"123"})
	require.NoError(err)
	require.Equal(item1, item2)

	err = store.Delete(item2.ID)
	require.NoError(err)
	item4, err := store.Read([]string{"123"})
	require.NoError(err)
	require.Equal(Item{}, item4)
}

func DoTestStoreErrors(store Store, t *testing.T) {
	require := require.New(t)
	err := store.Write(Item{})
	require.NotNil(err)
	require.True(strings.HasPrefix(err.Error(), "EMPTY_ITEM"))
}

func TestAllNamespaces(t *testing.T) {
	require := require.New(t)
	id := []string{"Organization", "Org1", "Team", "Team1"}
	ns := AllNamespaces(id)
	exp := []string{"Organization", "Organization/Org1", "Organization/Org1/Team"}
	require.Equal(exp, ns)
}

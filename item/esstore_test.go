package item

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func getEsStore(t *testing.T) *EsStore {
	require := require.New(t)
	store, err := NewElasticStore(Elastic{"http://55.0.0.2:9200", 1, 0, "items_test"})
	require.Nil(err)
	require.NotNil(store)
	return store
}

func TestEsStoreBasics(t *testing.T) {
	getEsStore(t).Close()
}

func TestEsStore(t *testing.T) {
	DoTestStore(getEsStore(t), t)
}

func TestEsStoreErrors(t *testing.T) {
	DoTestStoreErrors(getEsStore(t), t)
}

func TestEsStoreSearch(t *testing.T) {
	store := getEsStore(t)
	require := require.New(t)
	item1 := Item{"123", "Table", "Table1", map[string]interface{}{
		"field1": "value1",
		"field2": "value2",
	}}
	item2 := Item{"124", "Table", "Table2", map[string]interface{}{
		"field1": "value1",
		"field2": "value4",
	}}
	err := store.Write(item1)
	require.NoError(err)
	err = store.Write(item2)
	require.NoError(err)

	items, err := store.Search(NewQuery("value1"))
	require.NoError(err)
	require.Equal(2, len(items))
	require.Equal("123", items[0].Item.ID)
	require.Equal("124", items[1].Item.ID)

	items, err = store.Search(NewQuery("Table1"))
	require.NoError(err)
	require.Equal(1, len(items))
	require.Equal("123", items[0].Item.ID)

	items, err = store.Search(NewQuery("value2"))
	require.NoError(err)
	require.Equal(1, len(items))
	require.Equal("123", items[0].Item.ID)

	items, err = store.Search(NewQuery("Table"))
	require.NoError(err)
	require.Equal(2, len(items))
	require.Equal("123", items[0].Item.ID)
	require.Equal("124", items[1].Item.ID)

	items, err = store.Search(NewQuery("value4"))
	require.NoError(err)
	require.Equal(1, len(items))
	require.Equal("124", items[0].Item.ID)

	items, err = store.Search(Page(NewQuery("value1"), 0, 1))
	require.NoError(err)
	require.Equal(1, len(items))
	require.Equal("123", items[0].Item.ID)

	items, err = store.Search(Page(NewQuery("value1"), 1, 10))
	require.NoError(err)
	require.Equal(1, len(items))
	require.Equal("124", items[0].Item.ID)
}

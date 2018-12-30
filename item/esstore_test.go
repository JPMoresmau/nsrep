package item

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func getEsStore(t *testing.T) *EsStore {
	require := require.New(t)
	store, err := NewElasticStore(Elastic{URL: "http://55.0.0.2:9200", Shards: 1, Replicas: 0, Index: "items_test"})
	require.NoError(err)
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
	item1 := Item{[]string{"123"}, "Table", "Table1", map[string]interface{}{
		"field1": "value1",
		"field2": "value2",
	}}
	item2 := Item{[]string{"124"}, "Table", "Table2", map[string]interface{}{
		"field1": "value1",
		"field2": "value4",
	}}
	err := store.Write(item1)
	require.NoError(err)
	err = store.Write(item2)
	require.NoError(err)

	defer store.Delete(item1.ID)
	defer store.Delete(item2.ID)

	items, err := store.Search(NewQuery("value1"))
	require.NoError(err)
	require.Equal(2, len(items))
	require.Equal([]string{"123"}, items[0].Item.ID)
	require.Equal([]string{"124"}, items[1].Item.ID)

	items, err = store.Search(NewQuery("Table1"))
	require.NoError(err)
	require.Equal(1, len(items))
	require.Equal([]string{"123"}, items[0].Item.ID)

	items, err = store.Search(NewQuery("value2"))
	require.NoError(err)
	require.Equal(1, len(items))
	require.Equal([]string{"123"}, items[0].Item.ID)

	items, err = store.Search(NewQuery("Table"))
	require.NoError(err)
	require.Equal(2, len(items))
	require.Equal([]string{"123"}, items[0].Item.ID)
	require.Equal([]string{"124"}, items[1].Item.ID)

	items, err = store.Search(NewQuery("value4"))
	require.NoError(err)
	require.Equal(1, len(items))
	require.Equal([]string{"124"}, items[0].Item.ID)

	items, err = store.Search(Page(NewQuery("value1"), 0, 1))
	require.NoError(err)
	require.Equal(1, len(items))
	require.Equal([]string{"123"}, items[0].Item.ID)

	items, err = store.Search(Page(NewQuery("value1"), 1, 10))
	require.NoError(err)
	require.Equal(1, len(items))
	require.Equal([]string{"124"}, items[0].Item.ID)
}

func TestIDPrefix(t *testing.T) {
	store := getEsStore(t)
	require := require.New(t)
	item1 := Item{[]string{"DataSource", "DS1"}, "DataSource", "DS1", map[string]interface{}{
		"field1": "value1",
		"field2": "value2",
	}}
	item2 := Item{[]string{"DataSource", "DS1", "Table", "Table1"}, "Table", "Table1", map[string]interface{}{
		"field1": "value1",
		"field2": "value4",
	}}
	err := store.Write(item1)
	require.NoError(err)
	err = store.Write(item2)
	require.NoError(err)

	defer store.Delete(item1.ID)
	defer store.Delete(item2.ID)

	items, err := store.Search(NewQuery("item.id:DataSource/*"))
	require.NoError(err)
	require.Equal(2, len(items))
	require.Equal(item1.ID, items[0].Item.ID)
	require.Equal(item2.ID, items[1].Item.ID)
	items, err = store.Search(NewQuery("item.id:DataSource/DS1/*"))
	require.NoError(err)
	require.Equal(1, len(items))
	require.Equal(item2.ID, items[0].Item.ID)
}

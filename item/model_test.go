package item

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestModelAttributes(t *testing.T) {
	m0 := EmptyModel()
	require := require.New(t)
	item1 := Item{"Table/Table1", "Table", "Table1", map[string]interface{}{
		"attr1": "val1",
		"attr2": 3.14,
		"attr3": 3,
	}}
	changed, err := AddItem(item1, m0)
	require.True(changed)
	require.NoError(err)
	require.Equal([]string{"Table"}, ChildTypes(m0, ""))
	att0 := m0.TypeAttributes["Table"]
	require.NotNil(att0)
	require.Equal("string", att0["attr1"])
	require.Equal("float64", att0["attr2"])
	require.Equal("int", att0["attr3"])

	changed, err = AddItem(item1, m0)
	require.False(changed)
	require.NoError(err)
}

func TestModelWrongID(t *testing.T) {
	m0 := EmptyModel()
	require := require.New(t)
	item1 := Item{"Table1", "Table", "TableName", map[string]interface{}{
		"attr1": "val1",
		"attr2": 3.14,
		"attr3": 3,
	}}
	_, err := AddItem(item1, m0)
	require.Error(err)
	require.True(strings.Contains(err.Error(), "SHORT_ID"))
	require.True(strings.Contains(err.Error(), "Table1"))

	item1 = Item{"DataSource/Table1", "Table", "TableName", map[string]interface{}{
		"attr1": "val1",
		"attr2": 3.14,
		"attr3": 3,
	}}
	_, err = AddItem(item1, m0)
	require.Error(err)
	require.True(strings.Contains(err.Error(), "NO_TYPE"))
	require.True(strings.Contains(err.Error(), "Table"))
	require.True(strings.Contains(err.Error(), "DataSource"))
}

func TestModelAttributesTypeMismatch(t *testing.T) {
	m0 := EmptyModel()
	require := require.New(t)
	item1 := Item{"Table/Table1", "Table", "Table1", map[string]interface{}{
		"attr1": "val1",
		"attr2": 3.14,
	}}
	_, err := AddItem(item1, m0)
	require.NoError(err)

	item2 := Item{"Table/Table2", "Table", "Table2", map[string]interface{}{
		"attr1": 1,
		"attr3": true,
	}}
	_, err = AddItem(item2, m0)
	require.Error(err)
	require.Equal([]string{"Table"}, ChildTypes(m0, ""))
	att0 := m0.TypeAttributes["Table"]
	require.NotNil(att0)
	require.Equal("string", att0["attr1"])
	require.Equal("float64", att0["attr2"])
	require.Equal("bool", att0["attr3"])

	require.True(strings.Contains(err.Error(), "TYPE_MISMATCH"))
	require.True(strings.Contains(err.Error(), "attr1"))
}

func TestModelAttributesAndParent(t *testing.T) {
	m0 := EmptyModel()
	require := require.New(t)
	item1 := Item{"DataSource/DS1/Table/Table1", "Table", "Table1", map[string]interface{}{
		"attr1": "val1",
		"attr2": 3.14,
	}}
	_, err := AddItem(item1, m0)
	require.NoError(err)
	require.Equal([]string{"DataSource"}, ChildTypes(m0, ""))
	require.Equal([]string{"Table"}, ChildTypes(m0, "DataSource"))
	att0 := m0.TypeAttributes["Table"]
	require.NotNil(att0)
	require.Equal("string", att0["attr1"])
	require.Equal("float64", att0["attr2"])

}

func TestItemSerialization(t *testing.T) {
	m0 := EmptyModel()
	require := require.New(t)
	item1 := Item{"DataSource/DS1/Table/Table1", "Table", "Table1", map[string]interface{}{
		"attr1": "val1",
		"attr2": 3.14,
	}}
	_, err := AddItem(item1, m0)
	require.NoError(err)

	itemm := ToItem(m0)
	//log.Printf("%v", itemm.Contents)
	m1 := FromItem(itemm)
	require.NotNil(m1)
	require.Equal(m0.TypeAttributes, m1.TypeAttributes)
	require.Equal(ChildTypes(m0, ""), ChildTypes(m1, ""))
	require.Equal(ChildTypes(m0, "DataSource"), ChildTypes(m1, "DataSource"))
}

func TestEmptyItem(t *testing.T) {
	require := require.New(t)
	var it Item
	m0 := FromItem(it)
	require.NotNil(m0)
}

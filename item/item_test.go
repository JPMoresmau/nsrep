package item

import (
	"encoding/json"
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

package item

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var config = Cassandra{9042, "MetaRepTest", []string{"localhost"}, 1}

func TestStoreBasics(t *testing.T) {
	require := require.New(t)
	store, err := NewCqlStore(config)
	require.Nil(err)
	require.NotNil(store)
	store.Close()
}

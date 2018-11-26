package item

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var config = Cassandra{0, "MetaRepTest", []string{"localhost"}, 1}

func getStore(t *testing.T) *CqlStore {
	require := require.New(t)
	store, err := NewCqlStore(config)
	require.Nil(err)
	require.NotNil(store)
	return store
}

func TestCqlStoreBasics(t *testing.T) {
	getStore(t).Close()
}

func TestCqlStore(t *testing.T) {
	DoTestStore(getStore(t), t)
}

func TestCqlStoreErrors(t *testing.T) {
	DoTestStoreErrors(getStore(t), t)
}

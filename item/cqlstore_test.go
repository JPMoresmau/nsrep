package item

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var config = Cassandra{0, "NSRepTest", []string{"localhost"}, 1}

func getCqlStore(t *testing.T) *CqlStore {
	require := require.New(t)
	store, err := NewCqlStore(config)
	require.Nil(err)
	require.NotNil(store)
	return store
}

func TestCqlStoreBasics(t *testing.T) {
	getCqlStore(t).Close()
}

func TestCqlStore(t *testing.T) {
	DoTestStore(getCqlStore(t), t)
}

func TestCqlStoreErrors(t *testing.T) {
	DoTestStoreErrors(getCqlStore(t), t)
}

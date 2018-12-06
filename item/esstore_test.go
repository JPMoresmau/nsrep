package item

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func getEsStore(t *testing.T) *EsStore {
	require := require.New(t)
	store, err := NewElasticStore(Elastic{"http://55.0.0.2:9200", 1, 0})
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

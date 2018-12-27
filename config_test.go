package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadFileConfig(t *testing.T) {
	require := require.New(t)

	c, err := ReadFileConfig("application.yaml")
	require.Nil(err)
	require.Equal(8080, c.Port)
	require.Equal(9042, c.Cassandra.Port)
	require.Equal("NSRep", c.Cassandra.Keyspace)
	require.Equal(1, len(c.Cassandra.Endpoints))
	require.Equal("localhost", c.Cassandra.Endpoints[0])
	require.Equal("http://55.0.0.2:9200", c.Elastic.URL)
	require.Equal(1, c.Elastic.Shards)
	require.Equal(0, c.Elastic.Replicas)
	require.Equal("items_http", c.Elastic.Index)
}

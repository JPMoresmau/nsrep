package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadFileConfig(t *testing.T) {
	require := require.New(t)

	c, err := ReadFileConfig("application.yaml")
	require.Nil(err)
	require.Equal(9042, c.Cassandra.Port)
	require.Equal("LocalCluster", c.Cassandra.Name)
	require.Equal("MetaRep", c.Cassandra.Keyspace)
	require.Equal(1, len(c.Cassandra.Endpoints))
	require.Equal("localhost", c.Cassandra.Endpoints[0])
}

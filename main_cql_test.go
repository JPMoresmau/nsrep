package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	item "github.com/JPMoresmau/metarep/item"
	"github.com/stretchr/testify/require"
)

var config = item.Cassandra{Port: 0, Keyspace: "MetaRepTest", Endpoints: []string{"localhost"}, Replication: 1}

func TestCql(t *testing.T) {
	require := require.New(t)
	store, err := item.NewCqlStore(config)
	require.Nil(err)
	require.NotNil(store)
	srv := startServer(9999, store, nil)
	defer stopServer(srv)

	DoTestItem(t, "Table/tbl1")
	DoTestHistory(t, "Table/tbl1")
}

func DoTestHistory(t *testing.T, id string) {
	require := require.New(t)

	url := fmt.Sprintf("http://localhost:9999/history/%s?length=10", id)
	resp, err := http.Get(url)
	require.Nil(err)
	require.NotNil(resp)
	require.Equal(200, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	require.Nil(err)
	t.Log(string(body))
}

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	item "github.com/JPMoresmau/nsrep/item"
	"github.com/stretchr/testify/require"
)

var config = item.Cassandra{Port: 0, Keyspace: "NSRepTest", Endpoints: []string{"localhost"}, Replication: 1}
var elastic = item.Elastic{URL: "http://55.0.0.2:9200", Shards: 1, Replicas: 0, Index: "items_http_test"}

func TestCql(t *testing.T) {
	require := require.New(t)
	store, err := item.NewCqlStore(config)
	require.Nil(err)
	require.NotNil(store)
	srv, err := startServer(9999, store, nil)
	require.NoError(err)
	defer stopServer(srv)

	DoTestItem(t, "Table/tbl1")
	DoTestHistory(t, "Table/tbl1")
}

func TestCqlEs(t *testing.T) {
	require := require.New(t)
	store, err := item.NewCqlStore(config)
	require.NoError(err)
	require.NotNil(store)
	es, err := item.NewElasticStore(elastic)
	require.NoError(err)
	require.NotNil(es)
	srv, err := startServer(9999, store, es)
	require.NoError(err)
	defer stopServer(srv)

	DoTestItem(t, "Table/tbl1")
	DoTestHistory(t, "Table/tbl1")
	DoTestSearch(t)
	DoTestDeleteTree(t)
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
	// t.Log(string(body))
	var its = []item.Status{}
	require.Empty(its)
	json.Unmarshal(body, &its)
	require.NotEmpty(its)
}

func DoTestSearch(t *testing.T) {
	require := require.New(t)
	id1 := "DataSource/1"
	s1 := `{"type":"DataSource","name":"DataSource1","contents":{"field1":"value1","field2":"value2"}}`
	url1 := fmt.Sprintf("http://localhost:9999/items/%s", id1)
	id2 := "DataSource/2"
	s2 := `{"type":"DataSource","name":"DataSource2","contents":{"field1":"value1","field2":"value3"}}`
	url2 := fmt.Sprintf("http://localhost:9999/items/%s", id2)

	resp, err := http.Post(url1, "application/json", strings.NewReader(s1))
	require.Nil(err)
	require.NotNil(resp)
	require.Equal(200, resp.StatusCode)

	defer DoTestDelete(t, url1)

	resp, err = http.Post(url2, "application/json", strings.NewReader(s2))
	require.Nil(err)
	require.NotNil(resp)
	require.Equal(200, resp.StatusCode)

	defer DoTestDelete(t, url2)

	time.Sleep(time.Second)

	url := fmt.Sprintf("http://localhost:9999/search?query=%s", "value1")
	resp, err = http.Get(url)
	require.Nil(err)
	require.NotNil(resp)
	require.Equal(200, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	require.Nil(err)
	// t.Log(string(body))
	var its = []item.Score{}
	require.Empty(its)
	json.Unmarshal(body, &its)
	require.Equal(2, len(its))

}

func DoTestDeleteTree(t *testing.T) {
	require := require.New(t)
	id1 := "DataSource/1"
	s1 := `{"type":"DataSource","name":"DataSource1","contents":{"field1":"value1","field2":"value2"}}`
	url1 := fmt.Sprintf("http://localhost:9999/items/%s", id1)
	id2 := "DataSource/1/Table/1"
	s2 := `{"type":"Table","name":"Table1","contents":{"field1":"value1","field2":"value3"}}`
	url2 := fmt.Sprintf("http://localhost:9999/items/%s", id2)

	resp, err := http.Post(url1, "application/json", strings.NewReader(s1))
	require.Nil(err)
	require.NotNil(resp)
	require.Equal(200, resp.StatusCode)

	resp, err = http.Post(url2, "application/json", strings.NewReader(s2))
	require.Nil(err)
	require.NotNil(resp)
	require.Equal(200, resp.StatusCode)

	time.Sleep(time.Second)

	DoTestDelete(t, url1)

	resp, err = http.Get(url2)
	require.Nil(err)
	require.NotNil(resp)
	require.Equal(404, resp.StatusCode)

	// delete should be idempotent
	DoTestDelete(t, url1)

	resp, err = http.Get(url2)
	require.Nil(err)
	require.NotNil(resp)
	require.Equal(404, resp.StatusCode)
}
